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

package resources

import "fmt"

// entityTypeWirePair maps a user-facing `on:` entity-type key to the wire-
// format pair that represents it: an Expression<Entity> for the left side of
// an Equal (the "what is in this request's context" placeholder) and a
// LiteralExpression<Entity> for the right side (the named identity to match
// against). The Pulumi Cloud REST API consumes both wire types verbatim; the
// SDK exposes only the entity-type key plus an identity string.
type entityTypeWirePair struct {
	expression string // PermissionExpression<Entity>
	literal    string // PermissionLiteralExpression<Entity>
}

var entityTypeToWire = map[string]entityTypeWirePair{
	"environment": {
		expression: "PermissionExpressionEnvironment",
		literal:    "PermissionLiteralExpressionEnvironment",
	},
	"stack": {
		expression: "PermissionExpressionStack",
		literal:    "PermissionLiteralExpressionStack",
	},
	"insightsAccount": {
		expression: "PermissionExpressionInsightsAccount",
		literal:    "PermissionLiteralExpressionInsightsAccount",
	},
	"team": {
		expression: "PermissionExpressionTeam",
		literal:    "PermissionLiteralExpressionTeam",
	},
}

// wireToEntityType is the reverse map, built from entityTypeToWire so the two
// stay in lockstep. Keyed by the (expression, literal) pair concatenated with
// a separator unlikely to appear in either name.
var wireToEntityType = func() map[string]string {
	m := make(map[string]string, len(entityTypeToWire))
	for entity, pair := range entityTypeToWire {
		m[pair.expression+"|"+pair.literal] = entity
	}
	return m
}()

// validEntityTypeNames returns the sorted list of valid `on:` keys for use in
// error messages. Computed on demand because errors are rare.
func validEntityTypeNames() []string {
	names := make([]string, 0, len(entityTypeToWire))
	for name := range entityTypeToWire {
		names = append(names, name)
	}
	// Sort for stable error output across runs.
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j-1] > names[j]; j-- {
			names[j-1], names[j] = names[j], names[j-1]
		}
	}
	return names
}

// renameKey recursively walks a JSON-shaped tree (objects, arrays, scalars
// represented as map[string]interface{} / []interface{} / primitives) and
// renames every occurrence of `from` to `to` as a key. The transformation
// is structural: it does not interpret values, so it is safe to call on
// arbitrary subtrees. Returns a new tree; does not mutate the input.
//
// Used at the wire boundary to swap the discriminator name `__type` (used
// by Pulumi Cloud's tagged-union serialization) for `kind` (used at the
// SDK boundary, where Python's RPC deserializer would otherwise strip the
// `__`-prefixed key — see pulumi/sdk/python/lib/pulumi/runtime/rpc.py:866).
func renameKey(node interface{}, from, to string) interface{} {
	switch n := node.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(n))
		for k, v := range n {
			outKey := k
			if k == from {
				outKey = to
			}
			out[outKey] = renameKey(v, from, to)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(n))
		for i, v := range n {
			out[i] = renameKey(v, from, to)
		}
		return out
	default:
		return n
	}
}

// isWireTypeName is a sanity check for pass-through kind values: we accept
// any non-empty string starting with "Permission". This catches typos
// (e.g. `kind: "deny"`) before the wire round-trip without us having to
// enumerate every valid PermissionDescriptor* / PermissionExpression* /
// PermissionLiteralExpression* type name. Cloud rejects unknown wire
// types with its own error on Create/Update.
func isWireTypeName(s string) bool {
	const prefix = "Permission"
	return len(s) > len(prefix) && s[:len(prefix)] == prefix
}

// permissionsKindToWire converts a user-facing PermissionDescriptor tree
// (kind-shaped) to the Pulumi Cloud REST API's wire shape. The user-facing
// shape has just two kinds (`allow`, `group`), each optionally carrying an
// `on:` modifier that scopes the descriptor to a single entity. The wire
// shape is the full PermissionDescriptor tagged-union tree
// (Allow / Group / Condition(Equal(Expression<E>, Literal<E>(id)), <subNode>)).
//
// Returns an error if the input is malformed: missing kind, unknown kind,
// invalid `on:` shape, or invalid Allow/Group payload.
func permissionsKindToWire(node map[string]interface{}) (map[string]interface{}, error) {
	rawKind, ok := node["kind"]
	if !ok {
		return nil, fmt.Errorf("permissions descriptor missing required `kind` field")
	}
	kind, ok := rawKind.(string)
	if !ok {
		return nil, fmt.Errorf("permissions descriptor `kind` must be a string, got %T", rawKind)
	}

	var inner map[string]interface{}
	switch kind {
	case "allow":
		permsRaw, ok := node["permissions"]
		if !ok {
			return nil, fmt.Errorf("`allow` descriptor missing required `permissions` field")
		}
		perms, ok := permsRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("`allow.permissions` must be a list, got %T", permsRaw)
		}
		inner = map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": perms,
		}
	case "group":
		entriesRaw, ok := node["entries"]
		if !ok {
			return nil, fmt.Errorf("`group` descriptor missing required `entries` field")
		}
		entries, ok := entriesRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("`group.entries` must be a list, got %T", entriesRaw)
		}
		translatedEntries := make([]interface{}, len(entries))
		for i, entry := range entries {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("`group.entries[%d]` must be an object, got %T", i, entry)
			}
			translated, err := permissionsKindToWire(entryMap)
			if err != nil {
				return nil, fmt.Errorf("group.entries[%d]: %w", i, err)
			}
			translatedEntries[i] = translated
		}
		inner = map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": translatedEntries,
		}
	default:
		// Pass-through: PascalCase wire-type names flow through with a
		// blind kind → __type rename. The SDK does not interpret the
		// node's other fields; Cloud validates the structure on
		// Create/Update.
		if !isWireTypeName(kind) {
			return nil, fmt.Errorf(
				"unknown permissions descriptor kind %q "+
					"(valid: `allow`, `group`, or a PascalCase `Permission*` wire type name)",
				kind,
			)
		}
		// Pass-through nodes do not support the `on:` modifier — that
		// sugar only wraps structured allow/group descriptors.
		if _, hasOn := node["on"]; hasOn {
			return nil, fmt.Errorf(
				"`on:` modifier is only valid on `kind: allow` and `kind: group`; "+
					"got `on:` on pass-through kind %q",
				kind,
			)
		}
		return renameKey(node, "kind", "__type").(map[string]interface{}), nil
	}

	// If `on:` is set, wrap the inner descriptor in a Condition.
	if onRaw, hasOn := node["on"]; hasOn {
		condition, err := buildOnCondition(onRaw)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"__type":    "PermissionDescriptorCondition",
			"condition": condition,
			"subNode":   inner,
		}, nil
	}
	return inner, nil
}

// buildOnCondition translates a user-facing `on:` map into a
// PermissionExpressionEqual that compares the request-context expression for
// the entity type to a literal identity. Validates that `on:` is exactly a
// single-key map keyed by a known entity type with a string value.
func buildOnCondition(onRaw interface{}) (map[string]interface{}, error) {
	on, ok := onRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("`on` must be a map, got %T", onRaw)
	}
	if len(on) == 0 {
		return nil, fmt.Errorf(
			"`on` must have exactly one entity-type key (got an empty map); valid keys: %v",
			validEntityTypeNames(),
		)
	}
	if len(on) > 1 {
		return nil, fmt.Errorf(
			"`on` must have exactly one entity-type key (got %d); valid keys: %v",
			len(on), validEntityTypeNames(),
		)
	}
	for entityType, identityRaw := range on {
		pair, known := entityTypeToWire[entityType]
		if !known {
			return nil, fmt.Errorf("`on` key %q is not a known entity type (valid: %v)", entityType, validEntityTypeNames())
		}
		identity, ok := identityRaw.(string)
		if !ok {
			return nil, fmt.Errorf("`on.%s` must be a string, got %T", entityType, identityRaw)
		}
		return map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": pair.expression},
			"right":  map[string]interface{}{"__type": pair.literal, "identity": identity},
		}, nil
	}
	// Unreachable — the loop returns on the first iteration when len(on) == 1.
	return nil, fmt.Errorf("internal error: `on` validation fell through")
}

// permissionsWireToKind converts a wire-shape PermissionDescriptor (returned
// by Pulumi Cloud's REST API) back into the user-facing kind-shape. The
// reverse of permissionsKindToWire: collapses any Condition(Equal(...))
// wrapping into an `on:` modifier on the inner descriptor, and rewrites
// `__type` to `kind` for Allow / Group.
//
// Returns an error if the input has an unrecognised __type, or if a Condition
// wraps a shape this provider doesn't recognise (the general boolean-
// expression apparatus is intentionally not exposed at the SDK boundary).
func permissionsWireToKind(node map[string]interface{}) (map[string]interface{}, error) {
	rawType, ok := node["__type"]
	if !ok {
		return nil, fmt.Errorf("permissions descriptor missing required `__type` field")
	}
	wireType, ok := rawType.(string)
	if !ok {
		return nil, fmt.Errorf("permissions descriptor `__type` must be a string, got %T", rawType)
	}
	switch wireType {
	case "PermissionDescriptorAllow":
		permsRaw, ok := node["permissions"]
		if !ok {
			return nil, fmt.Errorf("`PermissionDescriptorAllow` missing required `permissions` field")
		}
		perms, ok := permsRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("`PermissionDescriptorAllow.permissions` must be a list, got %T", permsRaw)
		}
		return map[string]interface{}{
			"kind":        "allow",
			"permissions": perms,
		}, nil
	case "PermissionDescriptorGroup":
		entriesRaw, ok := node["entries"]
		if !ok {
			return nil, fmt.Errorf("`PermissionDescriptorGroup` missing required `entries` field")
		}
		entries, ok := entriesRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("`PermissionDescriptorGroup.entries` must be a list, got %T", entriesRaw)
		}
		translatedEntries := make([]interface{}, len(entries))
		for i, entry := range entries {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("`PermissionDescriptorGroup.entries[%d]` must be an object, got %T", i, entry)
			}
			translated, err := permissionsWireToKind(entryMap)
			if err != nil {
				return nil, fmt.Errorf("PermissionDescriptorGroup.entries[%d]: %w", i, err)
			}
			translatedEntries[i] = translated
		}
		return map[string]interface{}{
			"kind":    "group",
			"entries": translatedEntries,
		}, nil
	case "PermissionDescriptorCondition":
		// Try to collapse Condition(Equal(Context<E>, Lit<E>(id)), Allow|Group)
		// into an `on:` modifier on the inner descriptor. Falls back to
		// pass-through for any wire shape outside that narrow pattern.
		if collapsed, ok := tryCollapseOn(node); ok {
			return collapsed, nil
		}
		return renameKey(node, "__type", "kind").(map[string]interface{}), nil
	default:
		// Unknown wire type — pass through verbatim with a blind
		// __type → kind rename so Python's RPC deserializer doesn't
		// strip the discriminators. Cloud validates the structure at
		// Create/Update; we don't interpret pass-through fields.
		return renameKey(node, "__type", "kind").(map[string]interface{}), nil
	}
}

// tryCollapseOn attempts to collapse a wire-shape PermissionDescriptorCondition
// into the user-facing `on:` sugar. The collapse fires only when:
//
//  1. The condition is exactly PermissionExpressionEqual whose left/right are
//     a (Context<E>, Literal<E>(identity)) pair for one of the four supported
//     entity types (environment, stack, insightsAccount, team), AND
//  2. The wrapped subNode is a structured kind (PermissionDescriptorAllow or
//     PermissionDescriptorGroup) that the SDK exposes as `kind: allow` or
//     `kind: group`.
//
// If either guard fails, returns ok=false and the caller should fall back to
// pass-through. tryCollapseOn never returns an error — any malformed condition
// just doesn't collapse.
func tryCollapseOn(node map[string]interface{}) (map[string]interface{}, bool) {
	// Guard 2 first (cheaper): subNode must be Allow or Group on the wire.
	subRaw, ok := node["subNode"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	subType, _ := subRaw["__type"].(string)
	if subType != "PermissionDescriptorAllow" && subType != "PermissionDescriptorGroup" {
		return nil, false
	}

	// Guard 1: condition must be Equal(Context<E>, Lit<E>(id)) for a known E.
	condRaw, ok := node["condition"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	if condType, _ := condRaw["__type"].(string); condType != "PermissionExpressionEqual" {
		return nil, false
	}
	left, ok := condRaw["left"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	right, ok := condRaw["right"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	leftType, _ := left["__type"].(string)
	rightType, _ := right["__type"].(string)
	entityType, known := wireToEntityType[leftType+"|"+rightType]
	if !known {
		return nil, false
	}
	identity, ok := right["identity"].(string)
	if !ok {
		return nil, false
	}

	// Recurse into the subNode (we already checked it's Allow/Group, so the
	// recursion goes through a structured arm and won't return an error
	// for valid input).
	translated, err := permissionsWireToKind(subRaw)
	if err != nil {
		// Sub-translation failed — don't swallow it; signal "no collapse" so
		// the caller falls back to pass-through, which can't fail. The
		// pass-through path won't surface the structural error, but that's
		// fine: pass-through is for shapes the SDK doesn't fully validate.
		return nil, false
	}
	translated["on"] = map[string]interface{}{entityType: identity}
	return translated, true
}
