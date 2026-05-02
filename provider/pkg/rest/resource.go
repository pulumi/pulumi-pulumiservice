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
	"io"
	"net/http"
	"regexp"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// pathParamRE matches OpenAPI {param} placeholders in path strings.
var pathParamRE = regexp.MustCompile(`\{([^/{}]+)\}`)

// Resources builds dispatchable handlers — one per metadata.resources entry —
// suitable for registration with a runtime dispatcher.
//
// Mapping validation does NOT happen here. Operation IDs are resolved
// lazily at call time so that a half-broken metadata document doesn't fail
// the running provider; broken mappings surface either via BuildSchema
// (during GetSchema) or as runtime errors on the affected resource's
// first CRUD call.
func Resources(spec *Spec, metadata *Metadata) map[string]*DynamicResource {
	out := make(map[string]*DynamicResource, len(metadata.Resources))
	for token, rm := range metadata.Resources {
		out[token] = &DynamicResource{meta: rm, spec: spec}
	}
	return out
}

// DynamicResource is a metadata-driven resource handler. It satisfies
// pulumi-go-provider/middleware.CustomResource. Operation IDs are resolved
// against the spec at call time, not at construction.
type DynamicResource struct {
	meta ResourceMeta
	spec *Spec
}

// resolveOp looks up an operation ID against the bound spec. Returns
// (nil, nil) when id is empty (verb not declared on this resource).
func (r *DynamicResource) resolveOp(verb, id string) (*Operation, error) {
	if id == "" {
		return nil, nil
	}
	op, ok := r.spec.Op(id)
	if !ok {
		return nil, fmt.Errorf("rest: operations.%s = %q not found in spec", verb, id)
	}
	return op, nil
}

// Check is a passthrough; the schema's `replaceOnChanges` tags drive
// engine-side replacement decisions.
func (r *DynamicResource) Check(_ context.Context, req p.CheckRequest) (p.CheckResponse, error) {
	return p.CheckResponse{Inputs: req.Inputs}, nil
}

// Diff reports a coarse changes/no-changes outcome by comparing inputs.
func (r *DynamicResource) Diff(_ context.Context, req p.DiffRequest) (p.DiffResponse, error) {
	if mapEqual(req.OldInputs, req.Inputs) {
		return p.DiffResponse{}, nil
	}
	return p.DiffResponse{HasChanges: true}, nil
}

// Create executes the create operation: substitutes path params, JSON-encodes
// the inputs as the request body, fires the request, decodes the response,
// extracts the resource ID, returns the new state.
func (r *DynamicResource) Create(ctx context.Context, req p.CreateRequest) (p.CreateResponse, error) {
	if req.DryRun {
		return p.CreateResponse{Properties: req.Properties}, nil
	}
	op, err := r.resolveOp("create", r.meta.Operations.Create)
	if err != nil {
		return p.CreateResponse{}, err
	}
	if op == nil {
		return p.CreateResponse{}, fmt.Errorf("create: resource has no create operation declared")
	}
	body, state, err := r.execAndDecode(ctx, op, req.Properties)
	if err != nil {
		return p.CreateResponse{}, err
	}
	id, err := extractID(body, r.meta.IDField)
	if err != nil {
		return p.CreateResponse{}, fmt.Errorf("create: %w", err)
	}
	if id == "" {
		return p.CreateResponse{}, fmt.Errorf("create: response did not contain an ID at %q", idFieldOrDefault(r.meta.IDField))
	}
	return p.CreateResponse{ID: id, Properties: state}, nil
}

// Read fetches the current state. Path parameters come from the inputs (which
// the engine threads through; for fresh imports the user must supply enough
// inputs to identify the resource).
func (r *DynamicResource) Read(ctx context.Context, req p.ReadRequest) (p.ReadResponse, error) {
	op, err := r.resolveOp("read", r.meta.Operations.Read)
	if err != nil {
		return p.ReadResponse{}, err
	}
	if op == nil {
		return p.ReadResponse{}, fmt.Errorf("read: resource has no read operation declared")
	}
	source := req.Inputs
	if source.Len() == 0 {
		source = req.Properties
	}
	_, state, err := r.execAndDecode(ctx, op, source)
	if err != nil {
		return p.ReadResponse{}, err
	}
	return p.ReadResponse{ID: req.ID, Properties: state, Inputs: req.Inputs}, nil
}

// Update fires the update op (if declared).
func (r *DynamicResource) Update(ctx context.Context, req p.UpdateRequest) (p.UpdateResponse, error) {
	op, err := r.resolveOp("update", r.meta.Operations.Update)
	if err != nil {
		return p.UpdateResponse{}, err
	}
	if op == nil {
		return p.UpdateResponse{}, fmt.Errorf("update: resource has no update operation declared")
	}
	if req.DryRun {
		return p.UpdateResponse{Properties: req.Inputs}, nil
	}
	_, state, err := r.execAndDecode(ctx, op, req.Inputs)
	if err != nil {
		return p.UpdateResponse{}, err
	}
	return p.UpdateResponse{Properties: state}, nil
}

// Delete fires the delete op (if declared). Resources without a delete op
// quietly succeed; the engine drops the state.
func (r *DynamicResource) Delete(ctx context.Context, req p.DeleteRequest) error {
	op, err := r.resolveOp("delete", r.meta.Operations.Delete)
	if err != nil {
		return err
	}
	if op == nil {
		return nil
	}
	_, _, err = r.execAndDecode(ctx, op, req.Properties)
	return err
}

// execAndDecode performs the HTTP round-trip. The returned state Map is the
// JSON response body decoded as Pulumi properties; the returned []byte is
// the raw response body (used for ID extraction).
func (r *DynamicResource) execAndDecode(ctx context.Context, op *Operation, inputs property.Map) ([]byte, property.Map, error) {
	transport, err := resolveTransport(ctx)
	if err != nil {
		return nil, property.Map{}, err
	}

	url, err := r.buildURL(op, inputs)
	if err != nil {
		return nil, property.Map{}, err
	}

	var body io.Reader
	if needsBody(op.Method) {
		bodyJSON, err := json.Marshal(propertyMapToAny(inputs))
		if err != nil {
			return nil, property.Map{}, fmt.Errorf("rest: marshal request body for %s: %w", op.ID, err)
		}
		body = bytes.NewReader(bodyJSON)
	}

	httpReq, err := http.NewRequestWithContext(ctx, op.Method, url, body)
	if err != nil {
		return nil, property.Map{}, fmt.Errorf("rest: build HTTP request for %s: %w", op.ID, err)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := transport.Do(ctx, httpReq)
	if err != nil {
		return nil, property.Map{}, fmt.Errorf("rest: %s %s: %w", op.Method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, property.Map{}, fmt.Errorf("rest: read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, property.Map{}, fmt.Errorf("rest: %s %s returned %d: %s", op.Method, url, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if len(respBody) == 0 || resp.StatusCode == http.StatusNoContent {
		return respBody, property.NewMap(nil), nil
	}
	var raw map[string]any
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return respBody, property.Map{}, fmt.Errorf("rest: decode response for %s: %w", op.ID, err)
	}
	return respBody, anyMapToPropertyMap(raw), nil
}

// buildURL substitutes {path} placeholders from inputs. Returns an absolute
// URL; the spec's first server is used as the base. The transport may
// override the host before sending.
func (r *DynamicResource) buildURL(op *Operation, inputs property.Map) (string, error) {
	if len(r.spec.Servers) == 0 {
		return "", fmt.Errorf("rest: spec has no servers; cannot build URL for %s", op.ID)
	}
	matches := pathParamRE.FindAllStringSubmatchIndex(op.Path, -1)
	var b strings.Builder
	last := 0
	for _, m := range matches {
		b.WriteString(op.Path[last:m[0]])
		wireName := op.Path[m[2]:m[3]]
		pulName := pulumiName(wireName, r.meta.Renames, true)
		v, ok := inputs.GetOk(pulName)
		if !ok {
			return "", fmt.Errorf("rest: path parameter %q (Pulumi name %q) missing from inputs", wireName, pulName)
		}
		b.WriteString(propertyValueToString(v))
		last = m[1]
	}
	b.WriteString(op.Path[last:])
	return strings.TrimRight(r.spec.Servers[0], "/") + b.String(), nil
}

func needsBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	}
	return false
}

// extractID resolves a JSON-pointer-ish path against a JSON-decoded body.
func extractID(body []byte, ptr string) (string, error) {
	ptr = idFieldOrDefault(ptr)
	if len(body) == 0 {
		return "", nil
	}
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("decode response for ID extraction: %w", err)
	}
	cursor := raw
	for _, seg := range strings.Split(strings.TrimPrefix(ptr, "/"), "/") {
		if seg == "" {
			continue
		}
		m, ok := cursor.(map[string]any)
		if !ok {
			return "", nil
		}
		cursor, ok = m[seg]
		if !ok {
			return "", nil
		}
	}
	switch v := cursor.(type) {
	case string:
		return v, nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func idFieldOrDefault(p string) string {
	if p == "" {
		return "/id"
	}
	return p
}

// propertyValueToString stringifies a Value for path/URL substitution.
func propertyValueToString(v property.Value) string {
	switch {
	case v.IsString():
		return v.AsString()
	case v.IsNumber():
		return fmt.Sprintf("%v", v.AsNumber())
	case v.IsBool():
		return fmt.Sprintf("%v", v.AsBool())
	case v.IsNull():
		return ""
	default:
		return fmt.Sprintf("%v", propertyValueToAny(v))
	}
}

// propertyMapToAny converts a property.Map to a generic any tree suitable for
// json.Marshal. Secrets are unwrapped (the wire shape doesn't preserve them).
func propertyMapToAny(m property.Map) map[string]any {
	out := make(map[string]any, m.Len())
	for k, v := range m.AllStable {
		out[k] = propertyValueToAny(v)
	}
	return out
}

func propertyValueToAny(v property.Value) any {
	switch {
	case v.IsNull():
		return nil
	case v.IsBool():
		return v.AsBool()
	case v.IsNumber():
		return v.AsNumber()
	case v.IsString():
		return v.AsString()
	case v.IsArray():
		arr := v.AsArray()
		out := make([]any, 0, arr.Len())
		for _, e := range arr.AsSlice() {
			out = append(out, propertyValueToAny(e))
		}
		return out
	case v.IsMap():
		return propertyMapToAny(v.AsMap())
	default:
		return nil
	}
}

// anyMapToPropertyMap converts a JSON-decoded map to a property.Map.
func anyMapToPropertyMap(m map[string]any) property.Map {
	out := make(map[string]property.Value, len(m))
	for k, v := range m {
		out[k] = anyToPropertyValue(v)
	}
	return property.NewMap(out)
}

func anyToPropertyValue(v any) property.Value {
	switch x := v.(type) {
	case nil:
		return property.New(property.Null)
	case bool:
		return property.New(x)
	case float64:
		return property.New(x)
	case string:
		return property.New(x)
	case []any:
		arr := make([]property.Value, len(x))
		for i, e := range x {
			arr[i] = anyToPropertyValue(e)
		}
		return property.New(property.NewArray(arr))
	case map[string]any:
		return property.New(anyMapToPropertyMap(x))
	default:
		return property.New(property.Null)
	}
}

// mapEqual is a structural comparison helper for Diff.
func mapEqual(a, b property.Map) bool {
	if a.Len() != b.Len() {
		return false
	}
	for k, v := range a.AllStable {
		bv, ok := b.GetOk(k)
		if !ok {
			return false
		}
		if !v.Equals(bv) {
			return false
		}
	}
	return true
}
