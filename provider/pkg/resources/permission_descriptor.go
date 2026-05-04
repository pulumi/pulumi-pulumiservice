// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The descriptor translator is structurally blind: it walks the JSON-ish
// tree and renames the discriminator field at every level. The user-facing
// SDK boundary uses `discriminator`; the Pulumi Cloud REST API uses
// `__type`. The translator's two directions both reduce to a single
// recursive rename via renameDiscriminator.
//
// Why recursive (not top-only): Pulumi's Python SDK strips `__`-prefixed
// keys from resource inputs at *every nesting level* (pulumi/pulumi#22738),
// not just the top. Helpers like `buildEnvironmentScopedPermissions` produce
// trees with discriminators at four levels (Condition → Equal → Operand →
// Allow); if those used `__type` directly, Python would strip the nested
// `__type` keys before the input reached the provider, leaving an empty
// discriminator the typed API cannot dispatch on. So helpers and direct
// authors both use `discriminator` at every level, and the provider does
// the wholesale rename. The schema doesn't type the nested fields — but
// the Python SDK does transform them, which is what compels the recursion.

package resources

import "fmt"

// permissionsToWire converts a user-facing PermissionDescriptor tree into
// the Pulumi Cloud REST API's wire shape. Recursive rename: every map node
// with a `discriminator` key has it replaced by `__type`. No descriptor
// variant is hard-coded — Allow, Group, Condition, Compose, IfThenElse,
// Select, the boolean operators, and any future variant Pulumi Cloud adds
// pass through unchanged.
//
// Returns an error if the input contains a `__type` key anywhere — that
// almost always means the user copied raw wire format from the REST API
// docs and is one rebuild away from a Python SDK that silently drops the
// field. Pointing them at `discriminator` here gives a clear signal.
func permissionsToWire(node map[string]interface{}) (map[string]interface{}, error) {
	if err := assertNoUnderscoreType(node); err != nil {
		return nil, err
	}
	out := renameDiscriminator(node, "discriminator", "__type")
	outMap, ok := out.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("permissions descriptor must be an object, got %T", node)
	}
	if _, hasDiscriminator := outMap["__type"]; !hasDiscriminator {
		return nil, fmt.Errorf("permissions descriptor missing required `discriminator` field")
	}
	return outMap, nil
}

// permissionsToWireForAPI is the entry point used by Create/Update. It runs
// permissionsToWire and, if the top-level descriptor is a Condition, wraps
// it in a single-entry Group. Pulumi Cloud's role-detail UI 500s on a bare
// top-level Condition descriptor — the API itself accepts the Create, it's
// just the UI that breaks. Wrapping in a Group fixes the UI.
//
// permissionsFromWire reverses the wrap on Read so refresh stays idempotent
// (gated on the user's prior input shape — see comments there).
func permissionsToWireForAPI(node map[string]interface{}) (map[string]interface{}, error) {
	wire, err := permissionsToWire(node)
	if err != nil {
		return nil, err
	}
	if wire["__type"] == "PermissionDescriptorCondition" {
		return map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": []interface{}{wire},
		}, nil
	}
	return wire, nil
}

// permissionsFromWire converts a wire-shape PermissionDescriptor tree
// returned by Pulumi Cloud's REST API back into the user-facing shape.
// Reverse of permissionsToWire: a recursive rename of `__type` to
// `discriminator`.
//
// At the top level, optionally collapses a single-entry Group whose only
// entry is a Condition — the artefact of the API-boundary wrap added by
// permissionsToWireForAPI. The collapse is gated on `prior` so the
// round-trip is non-lossy:
//
//   - If the user authored a top-level Condition (or imported a role with
//     no prior input), collapse — the wrapped Group(Condition) wire shape
//     reads back as Condition. Matches helper output.
//   - If the user authored a top-level Group, do not collapse — preserve
//     their Group(Condition) shape verbatim.
//   - All other prior shapes (Allow, Compose, IfThenElse, …) cannot produce
//     a single-entry Group(Condition) on the wire, so the gate has no
//     effect on them.
func permissionsFromWire(
	node map[string]interface{},
	prior map[string]interface{},
) (map[string]interface{}, error) {
	out := renameDiscriminator(node, "__type", "discriminator")
	outMap, ok := out.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("permissions descriptor must be an object, got %T", node)
	}

	preserveGroupShape := prior != nil && prior["discriminator"] == "PermissionDescriptorGroup"
	if preserveGroupShape {
		return outMap, nil
	}
	if outMap["discriminator"] != "PermissionDescriptorGroup" {
		return outMap, nil
	}
	entries, ok := outMap["entries"].([]interface{})
	if !ok || len(entries) != 1 {
		return outMap, nil
	}
	entry, ok := entries[0].(map[string]interface{})
	if !ok || entry["discriminator"] != "PermissionDescriptorCondition" {
		return outMap, nil
	}
	return entry, nil
}

// renameDiscriminator walks a JSON-ish tree (map[string]interface{} /
// []interface{} / scalars) and returns a deep copy with every occurrence of
// the `from` key on a map node replaced by `to`. Other keys, values, and
// nesting are preserved verbatim.
func renameDiscriminator(v interface{}, from, to string) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, val := range x {
			outKey := k
			if k == from {
				outKey = to
			}
			out[outKey] = renameDiscriminator(val, from, to)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, item := range x {
			out[i] = renameDiscriminator(item, from, to)
		}
		return out
	default:
		return v
	}
}

// assertNoUnderscoreType walks a JSON-ish tree and returns an error if any
// map node contains a `__type` key. Defensive: Pulumi's Python SDK strips
// `__`-prefixed input keys (pulumi/pulumi#22738), so a Python user pasting
// raw wire format from the REST API docs would have the discriminator
// quietly disappear at the language boundary and the role would be created
// with a malformed descriptor. Rejecting `__type` at the SDK boundary
// surfaces a clear error pointing at `discriminator`.
func assertNoUnderscoreType(v interface{}) error {
	switch x := v.(type) {
	case map[string]interface{}:
		if _, has := x["__type"]; has {
			return fmt.Errorf(
				"permissions descriptor uses `__type` field — use `discriminator` instead. " +
					"Pulumi's Python SDK strips `__`-prefixed input keys at every nesting level " +
					"(pulumi/pulumi#22738), so the SDK boundary uses `discriminator` for every " +
					"language. The field's values are unchanged " +
					"(`PermissionDescriptorAllow`, `PermissionDescriptorGroup`, etc.)",
			)
		}
		for _, val := range x {
			if err := assertNoUnderscoreType(val); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range x {
			if err := assertNoUnderscoreType(item); err != nil {
				return err
			}
		}
	}
	return nil
}
