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

// The provider owns only the top of the permission descriptor tree.
// `OrganizationRole.permissions` is typed `map[string]Any` in the schema, so
// the Pulumi infra has no opinion about anything below the outer map.
// Translation is one rename at the top:
//
//	SDK boundary  →  wire (Create/Update)    discriminator → __type
//	wire          →  SDK boundary (Read)     __type        → discriminator
//
// Everything below the top is opaque pass-through. Helper functions like
// `buildEnvironmentScopedPermissions` and direct authors put the wire-format
// (`__type` at every nested level) inside; the provider forwards it to
// Pulumi Cloud verbatim. The SDK boundary uses `discriminator` rather than
// `__type` for two reasons:
//
//  1. Pulumi's Python SDK strips `__`-prefixed input keys
//     (pulumi/pulumi#22738), so `__type` would silently disappear at the
//     language boundary.
//  2. `discriminator` is reserved against future Pulumi Cloud models that may
//     legitimately carry a domain `kind` field (we already have `TeamKind`,
//     `PolicyIssueKind`, `ScheduledActionKind`, … in the public types).
//
// The reason this stays a single rename and not a recursive walk: the schema
// does not type the nested fields. They are not part of the provider's
// contract; they are payload the user assembled (often via a helper) that
// the provider hands off to the API. Any structural opinion on nested levels
// belongs in the helper that builds the tree, not in the translator.

package resources

import "fmt"

// permissionsToWire promotes the SDK-boundary `discriminator` key to the
// wire's `__type`. Top-level only: nested fields pass through unchanged.
// Returns an error if the input is missing `discriminator` or is using
// `__type` directly (a clear signal the user pasted raw wire format).
func permissionsToWire(node map[string]interface{}) (map[string]interface{}, error) {
	if node == nil {
		return nil, fmt.Errorf("permissions descriptor must be an object")
	}
	if _, hasUnderscore := node["__type"]; hasUnderscore {
		return nil, fmt.Errorf(
			"permissions descriptor uses `__type` at the top — use `discriminator` " +
				"instead at the SDK boundary. (Nested levels of the descriptor tree " +
				"continue to use the wire-format `__type`; only the top is renamed. " +
				"Pulumi's Python SDK strips `__`-prefixed input keys — pulumi/pulumi#22738.)",
		)
	}
	discriminator, ok := node["discriminator"]
	if !ok {
		return nil, fmt.Errorf("permissions descriptor missing required `discriminator` field at the top")
	}
	return swapTopKey(node, "discriminator", "__type", discriminator), nil
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

// permissionsFromWire promotes the wire's `__type` key back to the SDK
// boundary's `discriminator`. Top-level only — the rest of the tree carries
// `__type` verbatim from Pulumi Cloud's response.
//
// At the top, optionally collapses a single-entry Group whose only entry is
// a Condition — the artefact of permissionsToWireForAPI's UI-workaround
// wrap. The collapse is gated on the user's prior input shape so the
// round-trip is non-lossy:
//
//   - If the user authored a top-level Condition (or imported a role with
//     no prior input), collapse — the wrapped Group(Condition) wire shape
//     reads back as Condition, matching helper output.
//   - If the user authored a top-level Group, do not collapse — preserve
//     their Group(Condition) shape verbatim.
//   - All other prior shapes (Allow, Compose, IfThenElse, …) cannot produce
//     a single-entry Group(Condition) on the wire, so the gate has no
//     effect on them.
func permissionsFromWire(
	node map[string]interface{},
	prior map[string]interface{},
) (map[string]interface{}, error) {
	if node == nil {
		return nil, fmt.Errorf("permissions descriptor must be an object")
	}
	out := topWireToSDK(node)

	preserveGroupShape := prior != nil && prior["discriminator"] == "PermissionDescriptorGroup"
	if preserveGroupShape {
		return out, nil
	}
	if out["discriminator"] != "PermissionDescriptorGroup" {
		return out, nil
	}
	entries, ok := out["entries"].([]interface{})
	if !ok || len(entries) != 1 {
		return out, nil
	}
	entry, ok := entries[0].(map[string]interface{})
	// The entry is at a nested level, so it still carries `__type` (we
	// don't recurse on Read either). Promote it through topWireToSDK now
	// that it's becoming the new top.
	if !ok || entry["__type"] != "PermissionDescriptorCondition" {
		return out, nil
	}
	return topWireToSDK(entry), nil
}

// topWireToSDK is the shared top-level rename used by permissionsFromWire
// and the collapse path. `__type` → `discriminator`; everything else
// passes through unchanged.
func topWireToSDK(node map[string]interface{}) map[string]interface{} {
	t, ok := node["__type"]
	if !ok {
		// No discriminator on the wire — nothing to promote. Return a copy
		// so callers can safely mutate.
		out := make(map[string]interface{}, len(node))
		for k, v := range node {
			out[k] = v
		}
		return out
	}
	return swapTopKey(node, "__type", "discriminator", t)
}

// swapTopKey returns a copy of node with `from` removed and `to` set to
// value. Other keys pass through. The deep contents of the tree are not
// touched — top-level only by design.
func swapTopKey(node map[string]interface{}, from, to string, value interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(node))
	for k, v := range node {
		if k == from {
			continue
		}
		out[k] = v
	}
	out[to] = value
	return out
}
