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

// User-facing PermissionDescriptor maps use the wire-format `__type`
// discriminator at every level. Pulumi's Python SDK preserves
// `__`-prefixed keys across resource inputs as of pulumi/pulumi#22834
// (released in 3.235.0; pinned via the Python SDK's runtime requirement),
// so the language boundary is now transparent and the provider can pass
// the user's tree to the API verbatim. The only structural transform left
// is the Cloud UI workaround for top-level Conditions — see
// `permissionsForAPI` / `permissionsFromAPI` below.

package resources

import "fmt"

// validatePermissions sanity-checks a user-supplied descriptor map. The
// descriptor variants themselves are opaque to the provider (Allow, Group,
// Condition, Compose, IfThenElse, Select, the boolean operators, and any
// future variant Pulumi Cloud adds pass through unchanged); we only verify
// the top-level object carries the required `__type` discriminator key, so
// users see a clear error at preview rather than a 400 at apply.
func validatePermissions(node map[string]interface{}) error {
	if _, has := node["__type"]; !has {
		return fmt.Errorf("permissions descriptor missing required `__type` field")
	}
	return nil
}

// permissionsForAPI is the entry point used by Create/Update. If the
// top-level descriptor is a Condition, it wraps it in a single-entry
// Group: Pulumi Cloud's role-detail UI 500s on a bare top-level Condition
// descriptor — the API itself accepts the Create, only the UI breaks.
// Wrapping in a Group fixes the UI.
//
// permissionsFromAPI reverses the wrap on Read so refresh stays
// idempotent (gated on the user's prior input shape — see the comments
// there).
func permissionsForAPI(node map[string]interface{}) (map[string]interface{}, error) {
	if err := validatePermissions(node); err != nil {
		return nil, err
	}
	if node["__type"] == "PermissionDescriptorCondition" {
		return map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": []interface{}{node},
		}, nil
	}
	return node, nil
}

// permissionsFromAPI reads a descriptor as it came back from Pulumi
// Cloud. At the top level, it optionally collapses a single-entry Group
// whose only entry is a Condition — the artefact of the API-boundary
// wrap added by permissionsForAPI. The collapse is gated on `prior` so
// the round-trip is non-lossy:
//
//   - If the user authored a top-level Condition (or imported a role
//     with no prior input), collapse — the wrapped Group(Condition) wire
//     shape reads back as Condition. Matches helper output.
//   - If the user authored a top-level Group, do not collapse —
//     preserve their Group(Condition) shape verbatim.
//   - All other prior shapes (Allow, Compose, IfThenElse, …) cannot
//     produce a single-entry Group(Condition) on the wire, so the gate
//     has no effect on them.
func permissionsFromAPI(
	node map[string]interface{},
	prior map[string]interface{},
) (map[string]interface{}, error) {
	preserveGroupShape := prior != nil && prior["__type"] == "PermissionDescriptorGroup"
	if preserveGroupShape {
		return node, nil
	}
	if node["__type"] != "PermissionDescriptorGroup" {
		return node, nil
	}
	entries, ok := node["entries"].([]interface{})
	if !ok || len(entries) != 1 {
		return node, nil
	}
	entry, ok := entries[0].(map[string]interface{})
	if !ok || entry["__type"] != "PermissionDescriptorCondition" {
		return node, nil
	}
	return entry, nil
}
