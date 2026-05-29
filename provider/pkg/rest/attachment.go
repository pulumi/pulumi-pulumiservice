// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// Attachment resources model one membership edge of a parent resource whose
// membership API is imperative: a single mutation op (e.g. UpdatePolicyGroup)
// whose request body carries paired add/remove fields rather than a declarative
// list. The generic 1:1-CRUD Resource path can't express "send addStack on
// create, removeStack on delete, and read membership out of a list," so an
// AttachmentMeta on the entry routes Create/Read/Delete/Diff here instead.
//
// The design is deliberately diff-free: each edge is a single value, so there's
// no list reconciliation — Create adds it, Delete removes it, every input is
// replace-on-change, and Read is a membership test.
//
// Concurrency note: the mutation op sends only the single edge (server-side
// merge), not a full-list replace, so parallel attachments to the same parent
// are additive. The endpoint is imperative with no optimistic concurrency,
// though, so heavily interleaved add/remove against one parent ultimately
// relies on the server serializing membership updates.

// createAttachment adds the edge via MutationOp with the AddField body, then
// reads it back out of the parent's membership list to produce canonical state.
func (r *Resource) createAttachment(ctx context.Context, req p.CreateRequest) (p.CreateResponse, error) {
	if req.DryRun {
		return p.CreateResponse{Properties: req.Properties}, nil
	}
	am := r.meta.Attachment
	op, ok := r.spec.Op(am.MutationOp)
	if !ok {
		return p.CreateResponse{}, fmt.Errorf("attachment: mutationOp %q not found in spec", am.MutationOp)
	}
	if _, _, err := r.execAttachment(ctx, op, req.Properties, am.AddField); err != nil {
		return p.CreateResponse{}, fmt.Errorf("attachment create: %w", err)
	}
	state, found, err := r.readAttachmentState(ctx, req.Properties)
	if err != nil {
		return p.CreateResponse{}, fmt.Errorf("attachment read-after-create: %w", err)
	}
	if !found {
		// The add succeeded but the edge isn't visible in the parent yet;
		// fall back to the user's inputs (path params + edge fields), which
		// already match the resource's declared shape.
		state = req.Properties
	}
	id, err := r.synthesizeID(state, req.Properties)
	if err != nil {
		return p.CreateResponse{}, fmt.Errorf("attachment create: %w", err)
	}
	return p.CreateResponse{ID: id, Properties: state}, nil
}

// readAttachment GETs the parent and reports the edge present/absent. An absent
// edge returns an empty ReadResponse, signaling the engine to drop the resource
// from state (the membership equivalent of a 404 on a normal read).
func (r *Resource) readAttachment(ctx context.Context, req p.ReadRequest) (p.ReadResponse, error) {
	parsed := r.parseIDIntoInputs(req.ID, req.Inputs)
	source := mergeMaps(parsed, req.Properties)
	state, found, err := r.readAttachmentState(ctx, source)
	if err != nil {
		return p.ReadResponse{}, err
	}
	if !found {
		return p.ReadResponse{}, nil
	}
	returnedInputs := req.Inputs
	if req.Inputs.Len() == 0 {
		returnedInputs = parsed
	}
	return p.ReadResponse{ID: req.ID, Properties: state, Inputs: returnedInputs}, nil
}

// deleteAttachment removes the edge via MutationOp with the RemoveField body.
// A 404 (parent already gone) is success, mirroring the default Delete.
func (r *Resource) deleteAttachment(ctx context.Context, req p.DeleteRequest) error {
	am := r.meta.Attachment
	op, ok := r.spec.Op(am.MutationOp)
	if !ok {
		return fmt.Errorf("attachment: mutationOp %q not found in spec", am.MutationOp)
	}
	source := r.parseIDIntoInputs(req.ID, mergeMaps(req.Properties, req.OldInputs))
	if _, _, err := r.execAttachment(ctx, op, source, am.RemoveField); err != nil && !IsNotFound(err) {
		return err
	}
	return nil
}

// diffAttachment treats every input change as a replace: an edge has no update
// path, so any change is delete-the-old-edge / add-the-new-edge.
func (r *Resource) diffAttachment(req p.DiffRequest) p.DiffResponse {
	if mapEqual(req.OldInputs, req.Inputs) {
		return p.DiffResponse{}
	}
	detailed := map[string]p.PropertyDiff{}
	for k, newV := range req.Inputs.AllStable {
		oldV, ok := req.OldInputs.GetOk(k)
		if !ok {
			detailed[k] = p.PropertyDiff{Kind: p.AddReplace}
			continue
		}
		if !newV.Equals(oldV) {
			detailed[k] = p.PropertyDiff{Kind: p.UpdateReplace}
		}
	}
	for k := range req.OldInputs.AllStable {
		if _, ok := req.Inputs.GetOk(k); !ok {
			detailed[k] = p.PropertyDiff{Kind: p.DeleteReplace}
		}
	}
	return p.DiffResponse{
		HasChanges:          true,
		DeleteBeforeReplace: r.meta.DeleteBeforeReplace,
		DetailedDiff:        detailed,
	}
}

// execAttachment PATCHes the parent with a body nesting the edge under field
// (AddField or RemoveField). Path params come from urlSrc; the edge object is
// every non-path-param input.
func (r *Resource) execAttachment(
	ctx context.Context, op *Operation, urlSrc property.Map, field string,
) ([]byte, property.Map, error) {
	url, err := r.buildURL(op, urlSrc)
	if err != nil {
		return nil, property.Map{}, err
	}
	body := map[string]any{field: r.attachmentEdge(urlSrc)}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, property.Map{}, fmt.Errorf("rest: marshal attachment body for %s: %w", op.ID, err)
	}
	return r.roundTrip(ctx, op, url, bytes.NewReader(bodyJSON), contentJSON)
}

// readAttachmentState GETs the parent and returns the edge's state if it's a
// member. A 404 on the parent means the edge is gone too.
func (r *Resource) readAttachmentState(ctx context.Context, source property.Map) (property.Map, bool, error) {
	am := r.meta.Attachment
	op, ok := r.spec.Op(am.ReadOp)
	if !ok {
		return property.Map{}, false, fmt.Errorf("attachment: readOp %q not found in spec", am.ReadOp)
	}
	_, parent, err := r.execAndDecode(ctx, op, source)
	if err != nil {
		if IsNotFound(err) {
			return property.Map{}, false, nil
		}
		return property.Map{}, false, err
	}
	listVal, ok := parent.GetOk(am.MembershipField)
	if !ok || !listVal.IsArray() {
		return property.Map{}, false, nil
	}
	for _, elem := range listVal.AsArray().AsSlice() {
		// Object membership (e.g. stacks: [{name, routingProject}]) matches on
		// every MatchKey field; scalar membership (e.g. accounts: ["name"])
		// matches the single MatchKey field against the element directly.
		if elem.IsMap() {
			if r.attachmentElementMatches(elem.AsMap(), source) {
				return r.attachmentState(source, elem.AsMap()), true, nil
			}
			continue
		}
		if r.attachmentScalarMatches(elem, source) {
			return r.attachmentScalarState(source), true, nil
		}
	}
	return property.Map{}, false, nil
}

// edgeValuesEqual compares an identity value from the parent's membership list
// against a resource input by value alone. property.Value.Equals treats a
// secret value as unequal to its plain counterpart and an input carrying
// dependencies as unequal to one without, so a secret-wrapped or output-derived
// MatchKey would never match its own (always-plain) membership element. Strip
// both before comparing so identity is the underlying value, not its provenance.
func edgeValuesEqual(a, b property.Value) bool {
	return a.WithSecret(false).WithDependencies(nil).
		Equals(b.WithSecret(false).WithDependencies(nil))
}

// attachmentScalarMatches matches a scalar membership element (e.g. an account
// name) against the single MatchKey input field.
func (r *Resource) attachmentScalarMatches(elem property.Value, source property.Map) bool {
	mk := r.meta.Attachment.MatchKey
	if len(mk) != 1 {
		return false
	}
	want, ok := source.GetOk(pulumiName(mk[0], r.meta.Renames))
	return ok && edgeValuesEqual(elem, want)
}

// attachmentScalarState builds state for a scalar-membership match: the parent
// path params plus the edge's MatchKey field, all carried in source.
func (r *Resource) attachmentScalarState(source property.Map) property.Map {
	out := map[string]property.Value{}
	for name := range r.attachmentPathParamNames() {
		if v, ok := source.GetOk(name); ok {
			out[name] = v
		}
	}
	for _, k := range r.meta.Attachment.MatchKey {
		pk := pulumiName(k, r.meta.Renames)
		if v, ok := source.GetOk(pk); ok {
			out[pk] = v
		}
	}
	return property.NewMap(out)
}

// attachmentElementMatches reports whether a membership element identifies the
// same edge as source, comparing every MatchKey field.
func (r *Resource) attachmentElementMatches(elem, source property.Map) bool {
	for _, wireKey := range r.meta.Attachment.MatchKey {
		want, ok := source.GetOk(pulumiName(wireKey, r.meta.Renames))
		if !ok {
			return false
		}
		got, ok := elem.GetOk(wireKey)
		if !ok || !edgeValuesEqual(got, want) {
			return false
		}
	}
	return true
}

// attachmentState builds resource state from the parent path params (carried in
// source) plus the matched element's MatchKey fields. Only MatchKey is copied —
// it is exactly the edge's declared input/output set — so a richer membership
// element (server fields beyond the schema) can't leak undeclared outputs into
// state. Mirrors attachmentScalarState, which reads the same identity from
// source rather than a (scalar) element.
func (r *Resource) attachmentState(source, elem property.Map) property.Map {
	out := map[string]property.Value{}
	for name := range r.attachmentPathParamNames() {
		if v, ok := source.GetOk(name); ok {
			out[name] = v
		}
	}
	for _, k := range r.meta.Attachment.MatchKey {
		if v, ok := elem.GetOk(k); ok {
			out[pulumiName(k, r.meta.Renames)] = v
		}
	}
	return property.NewMap(out)
}

// attachmentEdge builds the edge object sent in the mutation body: every input
// that isn't a parent path param, keyed by its wire-side name.
func (r *Resource) attachmentEdge(inputs property.Map) map[string]any {
	pathParams := r.attachmentPathParamNames()
	edge := map[string]any{}
	for k, v := range inputs.AllStable {
		if pathParams[k] {
			continue
		}
		edge[wireSideName(k, r.meta.Renames)] = propertyValueToAny(v)
	}
	return edge
}

// attachmentPathParamNames returns the Pulumi-side names of the mutation op's
// path parameters — the inputs that identify the parent (URL), not the edge.
func (r *Resource) attachmentPathParamNames() map[string]bool {
	out := map[string]bool{}
	op, ok := r.spec.Op(r.meta.Attachment.MutationOp)
	if !ok {
		return out
	}
	for _, pp := range op.Parameters {
		if pp.In == inPath {
			out[pulumiName(pp.Name, r.meta.Renames)] = true
		}
	}
	return out
}
