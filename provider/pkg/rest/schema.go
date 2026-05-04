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
	"fmt"
	"slices"
	"strings"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

// BuildSchema produces a Pulumi PackageSpec describing every resource declared
// in metadata, using spec for type information.
//
// pkg is the schema package name (e.g. "pulumicloud") used to qualify token
// references emitted into the schema.
//
// Errors aggregate per-resource validation failures (missing operationIds,
// unresolvable $refs, unsupported constructs) so callers can see every
// problem from a single GetSchema call.
func BuildSchema(spec *Spec, metadata *Metadata, pkg string) (*schema.PackageSpec, error) {
	out := &schema.PackageSpec{
		Name:      pkg,
		Resources: map[string]schema.ResourceSpec{},
		Types:     map[string]schema.ComplexTypeSpec{},
		Functions: map[string]schema.FunctionSpec{},
	}

	var errs []string
	for token, rm := range metadata.Resources {
		rs, err := buildResource(spec, metadata, token, rm)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", token, err))
			continue
		}
		out.Resources[token] = *rs
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("rest: build schema:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return out, nil
}

func buildResource(spec *Spec, _ *Metadata, _ string, rm ResourceMeta) (*schema.ResourceSpec, error) {
	createID := rm.Operations.Create
	readID := rm.Operations.Read
	if createID == "" {
		return nil, fmt.Errorf("operations.create is required")
	}

	create, ok := spec.Op(createID)
	if !ok {
		return nil, fmt.Errorf("operations.create %q not found in spec", createID)
	}
	var read *Operation
	if readID != "" {
		read, ok = spec.Op(readID)
		if !ok {
			return nil, fmt.Errorf("operations.read %q not found in spec", readID)
		}
	}
	for verb, opID := range map[string]string{
		"update": rm.Operations.Update,
		"delete": rm.Operations.Delete,
	} {
		if opID == "" {
			continue
		}
		if _, ok := spec.Op(opID); !ok {
			return nil, fmt.Errorf("operations.%s %q not found in spec", verb, opID)
		}
	}

	// Inputs come from the create op: path params + request body schema.
	// Path params from read/update/delete are also exposed as inputs since
	// users supply them on create (the resource's identifier is part of the
	// state thereafter). Path params are forceNew by default.
	inputs, requiredInputs, err := operationInputs(spec, create, rm)
	if err != nil {
		return nil, fmt.Errorf("inputs: %w", err)
	}
	for _, op := range []*Operation{read} {
		if op == nil {
			continue
		}
		// Read-path params (e.g. server-generated IDs like issuerId,
		// poolId, scheduleID) surface as inputs so users _can_ supply
		// them on import, but they're not required on create — the
		// server returns them in the create response.
		if err := mergePathParamsAsInputs(inputs, nil, op, rm); err != nil {
			return nil, fmt.Errorf("inputs (read path params): %w", err)
		}
	}

	// Outputs come from the read op's response schema (read is the source of
	// truth for State); fall back to create's response if read has none.
	outputs, requiredOutputs, err := operationOutputs(spec, read, rm)
	if err != nil {
		return nil, fmt.Errorf("outputs: %w", err)
	}
	if len(outputs) == 0 {
		outputs, requiredOutputs, err = operationOutputs(spec, create, rm)
		if err != nil {
			return nil, fmt.Errorf("outputs (fallback to create): %w", err)
		}
	}
	// Path parameters need to round-trip through state so Delete (which
	// reads from saved state, not from inputs) can construct its URL.
	// Without this, deleting a resource fails with "path parameter X
	// missing from inputs" because X never made it into outputs.
	if outputs == nil {
		outputs = map[string]schema.PropertySpec{}
	}
	for _, op := range []*Operation{create, read} {
		if op == nil {
			continue
		}
		if err := mergePathParamsAsInputs(outputs, &requiredOutputs, op, rm); err != nil {
			return nil, fmt.Errorf("outputs (path params): %w", err)
		}
	}

	// Validate that every metadata.fields key matches an input or output.
	for fieldName := range rm.Fields {
		_, inInputs := inputs[fieldName]
		_, inOutputs := outputs[fieldName]
		if !inInputs && !inOutputs {
			return nil, fmt.Errorf("metadata.fields[%q] does not match any input or output field", fieldName)
		}
	}

	desc := rm.Description
	if desc == "" {
		desc = create.Description
	}
	if len(rm.Examples) > 0 {
		desc = appendExamples(desc, rm.Examples)
	}

	rs := &schema.ResourceSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type:        "object",
			Description: desc,
			Properties:  outputs,
			Required:    requiredOutputs,
		},
		InputProperties: inputs,
		RequiredInputs:  requiredInputs,
	}
	for _, alias := range rm.Aliases {
		rs.Aliases = append(rs.Aliases, schema.AliasSpec{Type: alias})
	}
	return rs, nil
}

// operationInputs builds the input PropertySpec map for a create operation:
// path/query parameters plus the request body schema's top-level properties.
func operationInputs(spec *Spec, op *Operation, rm ResourceMeta) (map[string]schema.PropertySpec, []string, error) {
	props := map[string]schema.PropertySpec{}
	required := map[string]bool{}

	// Path/query params first. These are always required (per OpenAPI's
	// constraints for path) and default to forceNew=true since they typically
	// participate in URL identity.
	for _, p := range op.Parameters {
		if p.In != "path" && p.In != "query" {
			continue
		}
		name := pulumiName(p.Name, rm.Renames, true)
		ps := schema.PropertySpec{
			TypeSpec:    schema.TypeSpec{Type: defaultParamType(p.SchemaType)},
			Description: p.Description,
		}
		if p.In == "path" {
			ps.WillReplaceOnChanges = true
		}
		applyFieldMeta(&ps, rm.Fields[name], true)
		props[name] = ps
		if p.Required || p.In == "path" {
			required[name] = true
		}
	}

	// Body schema (request).
	if op.RequestRef != "" {
		bodyProps, bodyRequired, err := flattenObjectSchema(spec, op.RequestRef)
		if err != nil {
			return nil, nil, fmt.Errorf("request body: %w", err)
		}
		for k, p := range bodyProps {
			name := pulumiName(k, rm.Renames, false)
			ps := openAPIToProperty(p)
			applyFieldMeta(&ps, rm.Fields[name], false)
			props[name] = ps
		}
		for _, r := range bodyRequired {
			required[pulumiName(r, rm.Renames, false)] = true
		}
	}

	return props, sortedKeys(required), nil
}

// mergePathParamsAsInputs adds path parameters from op to the inputs map if
// they aren't already present. Covers cases like read/update/delete using
// {id}-style path params that don't appear on create.
func mergePathParamsAsInputs(inputs map[string]schema.PropertySpec, required *[]string, op *Operation, rm ResourceMeta) error {
	for _, pp := range op.Parameters {
		if pp.In != "path" {
			continue
		}
		name := pulumiName(pp.Name, rm.Renames, true)
		if _, ok := inputs[name]; ok {
			continue
		}
		ps := schema.PropertySpec{
			TypeSpec:             schema.TypeSpec{Type: defaultParamType(pp.SchemaType)},
			Description:          pp.Description,
			WillReplaceOnChanges: true,
			ReplaceOnChanges:     true,
		}
		applyFieldMeta(&ps, rm.Fields[name], true)
		inputs[name] = ps
		if required != nil {
			*required = append(*required, name)
		}
	}
	return nil
}

// operationOutputs builds the State output PropertySpec map from an op's
// response body schema, then applies the metadata.outputs allowlist or
// metadata.outputsExclude denylist.
func operationOutputs(spec *Spec, op *Operation, rm ResourceMeta) (map[string]schema.PropertySpec, []string, error) {
	if op == nil || op.ResponseRef == "" {
		return nil, nil, nil
	}
	bodyProps, bodyRequired, err := flattenObjectSchema(spec, op.ResponseRef)
	if err != nil {
		return nil, nil, fmt.Errorf("response body: %w", err)
	}

	allowlist := stringSet(rm.Outputs)
	denylist := stringSet(rm.OutputsExclude)

	props := map[string]schema.PropertySpec{}
	required := map[string]bool{}

	for k, p := range bodyProps {
		name := pulumiName(k, rm.Renames, false)
		// "id" is reserved by Pulumi for resources — the resource ID is
		// extracted from the response separately via Ops.IDField. Skip it
		// from outputs so SDK codegen doesn't emit a colliding property.
		if name == "id" {
			continue
		}
		if len(allowlist) > 0 {
			if _, ok := allowlist[name]; !ok {
				continue
			}
		} else if _, blocked := denylist[name]; blocked {
			continue
		}
		ps := openAPIToProperty(p)
		applyFieldMeta(&ps, rm.Fields[name], false)
		if looksSecret(name) {
			ps.Secret = true
		}
		props[name] = ps
	}
	for _, r := range bodyRequired {
		name := pulumiName(r, rm.Renames, false)
		if _, kept := props[name]; kept {
			required[name] = true
		}
	}
	return props, sortedKeys(required), nil
}

// flattenObjectSchema resolves a $ref and walks any allOf chain, producing the
// merged set of top-level properties and the union of required fields.
//
// Returns an error for OpenAPI constructs we don't yet support (oneOf, anyOf,
// nested $refs in the union, additionalProperties, polymorphism). Aggregates
// require fields across all allOf branches.
func flattenObjectSchema(spec *Spec, ref string) (map[string]any, []string, error) {
	visited := map[string]bool{}
	props := map[string]any{}
	required := map[string]bool{}

	var walk func(node map[string]any, source string) error
	walk = func(node map[string]any, source string) error {
		if r, ok := node["$ref"].(string); ok {
			if visited[r] {
				return fmt.Errorf("cyclic $ref chain at %s", r)
			}
			visited[r] = true
			resolved, ok := spec.ResolveSchema(r)
			if !ok {
				return fmt.Errorf("unresolvable $ref %q (referenced from %s)", r, source)
			}
			return walk(resolved, r)
		}
		if all, ok := node["allOf"].([]any); ok {
			for i, m := range all {
				mm, ok := m.(map[string]any)
				if !ok {
					return fmt.Errorf("allOf[%d] in %s is not an object", i, source)
				}
				if err := walk(mm, fmt.Sprintf("%s.allOf[%d]", source, i)); err != nil {
					return err
				}
			}
			// allOf members may also carry properties at the same level.
		}
		if _, ok := node["oneOf"]; ok {
			return fmt.Errorf("oneOf not yet supported (at %s)", source)
		}
		if _, ok := node["anyOf"]; ok {
			return fmt.Errorf("anyOf not yet supported (at %s)", source)
		}
		if t, ok := node["type"].(string); ok && t != "object" && t != "" {
			return fmt.Errorf("expected object schema at %s, got %q", source, t)
		}
		if rr, ok := node["required"].([]any); ok {
			for _, r := range rr {
				if rs, ok := r.(string); ok {
					required[rs] = true
				}
			}
		}
		if pp, ok := node["properties"].(map[string]any); ok {
			for k, v := range pp {
				vm, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("property %q in %s is not an object", k, source)
				}
				props[k] = vm
			}
		}
		return nil
	}

	root, ok := spec.ResolveSchema(ref)
	if !ok {
		return nil, nil, fmt.Errorf("unresolvable $ref %q", ref)
	}
	visited[ref] = true
	if err := walk(root, ref); err != nil {
		return nil, nil, err
	}

	// Drop any required entries we didn't see as properties (e.g. inherited
	// requirements that another allOf branch satisfied; we union by name).
	finalRequired := []string{}
	for r := range required {
		if _, ok := props[r]; ok {
			finalRequired = append(finalRequired, r)
		}
	}
	slices.Sort(finalRequired)
	return props, finalRequired, nil
}

// openAPIToProperty converts an OpenAPI property object to a Pulumi
// PropertySpec.
func openAPIToProperty(node any) schema.PropertySpec {
	nm, ok := node.(map[string]any)
	if !ok {
		return schema.PropertySpec{TypeSpec: schema.TypeSpec{Ref: "pulumi.json#/Any"}}
	}
	ts := openAPIToType(nm)
	desc, _ := nm["description"].(string)
	return schema.PropertySpec{TypeSpec: ts, Description: desc}
}

// openAPIToType converts an OpenAPI schema node to a Pulumi TypeSpec.
//
// Constructs that aren't yet handled (nested $ref, oneOf, additionalProperties)
// degrade to "pulumi.json#/Any" so the schema is still serializable. The
// resource-level validator catches metadata that references fields whose
// shapes can't be modeled; for unrecognized payload fields we accept and
// pass through as untyped values.
func openAPIToType(node map[string]any) schema.TypeSpec {
	if _, ok := node["$ref"].(string); ok {
		return schema.TypeSpec{Ref: "pulumi.json#/Any"}
	}
	t, _ := node["type"].(string)
	switch t {
	case "string":
		return schema.TypeSpec{Type: "string"}
	case "integer":
		return schema.TypeSpec{Type: "integer"}
	case "number":
		return schema.TypeSpec{Type: "number"}
	case "boolean":
		return schema.TypeSpec{Type: "boolean"}
	case "array":
		items, _ := node["items"].(map[string]any)
		var itemTS schema.TypeSpec
		if items != nil {
			itemTS = openAPIToType(items)
		} else {
			itemTS = schema.TypeSpec{Ref: "pulumi.json#/Any"}
		}
		return schema.TypeSpec{Type: "array", Items: &itemTS}
	case "object", "":
		// Anonymous object — emit as a free-form object for now.
		return schema.TypeSpec{
			Type:                 "object",
			AdditionalProperties: &schema.TypeSpec{Ref: "pulumi.json#/Any"},
		}
	default:
		return schema.TypeSpec{Ref: "pulumi.json#/Any"}
	}
}

func applyFieldMeta(ps *schema.PropertySpec, fm FieldMeta, isPathParam bool) {
	if fm.ForceNew || isPathParam {
		ps.WillReplaceOnChanges = true
		ps.ReplaceOnChanges = true
	}
	if fm.Secret {
		ps.Secret = true
	}
	if fm.Description != "" {
		ps.Description = fm.Description
	}
}

// looksSecret heuristically detects whether a field name implies the
// value is sensitive. Catches OrganizationWebhook's `secret` and
// `secretCiphertext`, OrgToken/TeamToken/PersonalToken's `tokenValue`,
// and the AgentPool's response `tokenValue`. The OpenAPI spec doesn't
// carry an x-secret extension we can rely on, so this fills the gap.
func looksSecret(name string) bool {
	lower := strings.ToLower(name)
	for _, sub := range []string{"secret", "tokenvalue", "password", "apikey", "accesstoken", "ciphertext"} {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

func defaultParamType(t string) string {
	if t == "" {
		return "string"
	}
	return t
}

// pulumiName applies the metadata.renames map. The map's *keys* are
// Pulumi-side names (the SDK-facing identifier); values are OpenAPI-side
// names (matching wire). When wireSide is true the input is an OpenAPI name;
// when false (e.g. for body field names that the schema spec exposes) we
// invert.
func pulumiName(name string, renames map[string]string, wireSide bool) string {
	if wireSide {
		// path/query param names are wire-side; flip if a rename targets them.
		for pul, wire := range renames {
			if wire == name {
				return pul
			}
		}
		return name
	}
	// body-side names: same logic since renames maps Pulumi → wire.
	for pul, wire := range renames {
		if wire == name {
			return pul
		}
	}
	return name
}

func stringSet(ss []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		out[s] = struct{}{}
	}
	return out
}

// appendExamples concatenates an `## Example Usage` section onto a resource
// description, with each PCL snippet wrapped in a ```pulumi``` fenced block.
// Pulumi's SDK codegen recognizes the pulumi-tagged fence and runs
// `pulumi convert` per target language at gen time.
func appendExamples(desc string, examples []string) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(desc, "\n"))
	if b.Len() > 0 {
		b.WriteString("\n\n")
	}
	b.WriteString("## Example Usage\n")
	for _, ex := range examples {
		b.WriteString("\n```pulumi\n")
		b.WriteString(strings.TrimSpace(ex))
		b.WriteString("\n```\n")
	}
	return b.String()
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
