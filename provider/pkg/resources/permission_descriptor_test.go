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

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Cardinal shape fixtures. Each function returns a freshly-allocated map so
// tests can mutate without affecting other tests.
// ----------------------------------------------------------------------------

// flatAllow: a bare Allow with no scoping.
func flatAllowKind() map[string]interface{} {
	return map[string]interface{}{
		"kind":        "allow",
		"permissions": []interface{}{"stack:read"},
	}
}
func flatAllowWire() map[string]interface{} {
	return map[string]interface{}{
		"__type":      "PermissionDescriptorAllow",
		"permissions": []interface{}{"stack:read"},
	}
}

// scopedAllow: an Allow scoped to one environment.
func scopedAllowKind() map[string]interface{} {
	return map[string]interface{}{
		"kind":        "allow",
		"on":          map[string]interface{}{"environment": "env-uuid-1"},
		"permissions": []interface{}{"environment:read"},
	}
}
func scopedAllowWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionEnvironment",
				"identity": "env-uuid-1",
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"environment:read"},
		},
	}
}

// flatGroup: a Group of two bare Allows.
func flatGroupKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "group",
		"entries": []interface{}{
			map[string]interface{}{"kind": "allow", "permissions": []interface{}{"stack:read"}},
			map[string]interface{}{"kind": "allow", "permissions": []interface{}{"stack:edit"}},
		},
	}
}
func flatGroupWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{"__type": "PermissionDescriptorAllow", "permissions": []interface{}{"stack:read"}},
			map[string]interface{}{"__type": "PermissionDescriptorAllow", "permissions": []interface{}{"stack:edit"}},
		},
	}
}

// scopedGroup: a Group with a top-level `on:` scoping every entry.
func scopedGroupKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "group",
		"on":   map[string]interface{}{"stack": "stack-id-1"},
		"entries": []interface{}{
			map[string]interface{}{"kind": "allow", "permissions": []interface{}{"stack:read"}},
			map[string]interface{}{"kind": "allow", "permissions": []interface{}{"stack:edit"}},
		},
	}
}
func scopedGroupWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionStack"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionStack",
				"identity": "stack-id-1",
			},
		},
		"subNode": map[string]interface{}{
			"__type": "PermissionDescriptorGroup",
			"entries": []interface{}{
				map[string]interface{}{"__type": "PermissionDescriptorAllow", "permissions": []interface{}{"stack:read"}},
				map[string]interface{}{"__type": "PermissionDescriptorAllow", "permissions": []interface{}{"stack:edit"}},
			},
		},
	}
}

// mixedGroup: a Group whose entries each carry their own `on:` for different
// entities. Verifies that per-entry `on:` lifts independently.
func mixedGroupKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "group",
		"entries": []interface{}{
			map[string]interface{}{
				"kind":        "allow",
				"on":          map[string]interface{}{"environment": "env-1"},
				"permissions": []interface{}{"environment:read"},
			},
			map[string]interface{}{
				"kind":        "allow",
				"on":          map[string]interface{}{"insightsAccount": "acct-1"},
				"permissions": []interface{}{"insights-account:read"},
			},
		},
	}
}
func mixedGroupWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionEqual",
					"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
					"right": map[string]interface{}{
						"__type":   "PermissionLiteralExpressionEnvironment",
						"identity": "env-1",
					},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:read"},
				},
			},
			map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionEqual",
					"left":   map[string]interface{}{"__type": "PermissionExpressionInsightsAccount"},
					"right": map[string]interface{}{
						"__type":   "PermissionLiteralExpressionInsightsAccount",
						"identity": "acct-1",
					},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"insights-account:read"},
				},
			},
		},
	}
}

// teamScopedAllow: an Allow scoped to a single team. Validates the new
// `team` entry in the on:-sugar entity-type table.
func teamScopedAllowKind() map[string]interface{} {
	return map[string]interface{}{
		"kind":        "allow",
		"on":          map[string]interface{}{"team": "team-id-1"},
		"permissions": []interface{}{"stack:edit"},
	}
}
func teamScopedAllowWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionTeam"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionTeam",
				"identity": "team-id-1",
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
	}
}

// bareCompose: a Compose descriptor — list of descriptor IDs by reference.
// The wire shape is the customer's UI-imported role.
func bareComposeKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCompose",
		"permissionDescriptors": []interface{}{
			"046f4b97-ed29-43d3-a09a-2c5d8d0e44e0",
			"7be8e8d2-9c0c-4d34-a4f5-2d0fdb5e2f1a",
		},
	}
}
func bareComposeWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCompose",
		"permissionDescriptors": []interface{}{
			"046f4b97-ed29-43d3-a09a-2c5d8d0e44e0",
			"7be8e8d2-9c0c-4d34-a4f5-2d0fdb5e2f1a",
		},
	}
}

// ifThenElseDeeplyNested: a maximally-nested IfThenElse to exercise the
// blind discriminator rename at every level (IfThenElse → And → Equal /
// Or → Not → HasTag → ContextEnvironment, with inner Allow descriptors).
func ifThenElseDeeplyNestedKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorIfThenElse",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionAnd",
			"left": map[string]interface{}{
				"kind": "PermissionExpressionEqual",
				"left": map[string]interface{}{"kind": "PermissionExpressionStack"},
				"right": map[string]interface{}{
					"kind":     "PermissionLiteralExpressionStack",
					"identity": "stack-1",
				},
			},
			"right": map[string]interface{}{
				"kind": "PermissionExpressionNot",
				"node": map[string]interface{}{
					"kind":    "PermissionExpressionHasTag",
					"context": map[string]interface{}{"kind": "PermissionExpressionEnvironment"},
					"key":     "production",
				},
			},
		},
		"subNodeForTrue": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"environment:read"},
		},
		"subNodeForFalse": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"environment:read", "environment:write"},
		},
	}
}
func ifThenElseDeeplyNestedWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorIfThenElse",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionAnd",
			"left": map[string]interface{}{
				"__type": "PermissionExpressionEqual",
				"left":   map[string]interface{}{"__type": "PermissionExpressionStack"},
				"right": map[string]interface{}{
					"__type":   "PermissionLiteralExpressionStack",
					"identity": "stack-1",
				},
			},
			"right": map[string]interface{}{
				"__type": "PermissionExpressionNot",
				"node": map[string]interface{}{
					"__type":  "PermissionExpressionHasTag",
					"context": map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
					"key":     "production",
				},
			},
		},
		"subNodeForTrue": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"environment:read"},
		},
		"subNodeForFalse": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"environment:read", "environment:write"},
		},
	}
}

// selectMixedLiteralTypes: a Select with options keyed by three different
// literal types (string, number, bool) — verifies the blind rename
// handles each literal sub-type correctly.
func selectMixedLiteralTypesKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorSelect",
		"selector": map[string]interface{}{
			"kind":    "PermissionExpressionTag",
			"context": map[string]interface{}{"kind": "PermissionExpressionEnvironment"},
			"key":     "tier",
		},
		"options": []interface{}{
			map[string]interface{}{
				"value": map[string]interface{}{
					"kind":  "PermissionLiteralExpressionString",
					"value": "prod",
				},
				"node": map[string]interface{}{
					"kind":        "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:read"},
				},
			},
			map[string]interface{}{
				"value": map[string]interface{}{
					"kind":  "PermissionLiteralExpressionNumber",
					"value": 42,
				},
				"node": map[string]interface{}{
					"kind":        "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:write"},
				},
			},
			map[string]interface{}{
				"value": map[string]interface{}{
					"kind":  "PermissionLiteralExpressionBool",
					"value": true,
				},
				"node": map[string]interface{}{
					"kind":        "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:admin"},
				},
			},
		},
	}
}
func selectMixedLiteralTypesWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorSelect",
		"selector": map[string]interface{}{
			"__type":  "PermissionExpressionTag",
			"context": map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
			"key":     "tier",
		},
		"options": []interface{}{
			map[string]interface{}{
				"value": map[string]interface{}{
					"__type": "PermissionLiteralExpressionString",
					"value":  "prod",
				},
				"node": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:read"},
				},
			},
			map[string]interface{}{
				"value": map[string]interface{}{
					"__type": "PermissionLiteralExpressionNumber",
					"value":  42,
				},
				"node": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:write"},
				},
			},
			map[string]interface{}{
				"value": map[string]interface{}{
					"__type": "PermissionLiteralExpressionBool",
					"value":  true,
				},
				"node": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:admin"},
				},
			},
		},
	}
}

// conditionAndOfEquals: a Condition whose boolean is a non-Equal And. Must
// surface as a pass-through PermissionDescriptorCondition (not collapsed).
func conditionAndOfEqualsKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionAnd",
			"left": map[string]interface{}{
				"kind": "PermissionExpressionEqual",
				"left": map[string]interface{}{"kind": "PermissionExpressionStack"},
				"right": map[string]interface{}{
					"kind":     "PermissionLiteralExpressionStack",
					"identity": "stack-1",
				},
			},
			"right": map[string]interface{}{
				"kind": "PermissionExpressionEqual",
				"left": map[string]interface{}{"kind": "PermissionExpressionTeam"},
				"right": map[string]interface{}{
					"kind":     "PermissionLiteralExpressionTeam",
					"identity": "team-1",
				},
			},
		},
		"subNode": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
	}
}
func conditionAndOfEqualsWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionAnd",
			"left": map[string]interface{}{
				"__type": "PermissionExpressionEqual",
				"left":   map[string]interface{}{"__type": "PermissionExpressionStack"},
				"right": map[string]interface{}{
					"__type":   "PermissionLiteralExpressionStack",
					"identity": "stack-1",
				},
			},
			"right": map[string]interface{}{
				"__type": "PermissionExpressionEqual",
				"left":   map[string]interface{}{"__type": "PermissionExpressionTeam"},
				"right": map[string]interface{}{
					"__type":   "PermissionLiteralExpressionTeam",
					"identity": "team-1",
				},
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
	}
}

// conditionMismatchedOperands: Equal whose left/right are wrong-pair (e.g.
// ContextEnvironment vs LitStack). Must surface as pass-through.
func conditionMismatchedOperandsKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionEqual",
			"left": map[string]interface{}{"kind": "PermissionExpressionEnvironment"},
			"right": map[string]interface{}{
				"kind":     "PermissionLiteralExpressionStack",
				"identity": "x",
			},
		},
		"subNode": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:read"},
		},
	}
}
func conditionMismatchedOperandsWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionStack",
				"identity": "x",
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:read"},
		},
	}
}

// conditionWrappingCompose: collapsible Equal but subNode is Compose
// (a pass-through kind, not a structured Allow/Group). Must surface as
// pass-through PermissionDescriptorCondition (not collapse to on:).
func conditionWrappingComposeKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionEqual",
			"left": map[string]interface{}{"kind": "PermissionExpressionEnvironment"},
			"right": map[string]interface{}{
				"kind":     "PermissionLiteralExpressionEnvironment",
				"identity": "env-1",
			},
		},
		"subNode": map[string]interface{}{
			"kind":                  "PermissionDescriptorCompose",
			"permissionDescriptors": []interface{}{"role-id-1"},
		},
	}
}
func conditionWrappingComposeWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionEnvironment",
				"identity": "env-1",
			},
		},
		"subNode": map[string]interface{}{
			"__type":                "PermissionDescriptorCompose",
			"permissionDescriptors": []interface{}{"role-id-1"},
		},
	}
}

// conditionWrappingCondition: nested scoping. Outer collapse fires only
// if subNode is Allow/Group, so this surfaces as pass-through. The inner
// Condition is inside a pass-through subtree and stays in PascalCase.
func conditionWrappingConditionKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionEqual",
			"left": map[string]interface{}{"kind": "PermissionExpressionStack"},
			"right": map[string]interface{}{
				"kind":     "PermissionLiteralExpressionStack",
				"identity": "stack-1",
			},
		},
		"subNode": map[string]interface{}{
			"kind": "PermissionDescriptorCondition",
			"condition": map[string]interface{}{
				"kind": "PermissionExpressionEqual",
				"left": map[string]interface{}{"kind": "PermissionExpressionEnvironment"},
				"right": map[string]interface{}{
					"kind":     "PermissionLiteralExpressionEnvironment",
					"identity": "env-1",
				},
			},
			"subNode": map[string]interface{}{
				"kind":        "PermissionDescriptorAllow",
				"permissions": []interface{}{"environment:read"},
			},
		},
	}
}
func conditionWrappingConditionWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionStack"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionStack",
				"identity": "stack-1",
			},
		},
		"subNode": map[string]interface{}{
			"__type": "PermissionDescriptorCondition",
			"condition": map[string]interface{}{
				"__type": "PermissionExpressionEqual",
				"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
				"right": map[string]interface{}{
					"__type":   "PermissionLiteralExpressionEnvironment",
					"identity": "env-1",
				},
			},
			"subNode": map[string]interface{}{
				"__type":      "PermissionDescriptorAllow",
				"permissions": []interface{}{"environment:read"},
			},
		},
	}
}

// composeInsideGroup: a structured Group containing a pass-through Compose
// entry. Verifies the boundary at entries[1]: the structured group recursion
// hands off to pass-through when it encounters a non-Allow/non-Group entry.
func composeInsideGroupKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "group",
		"entries": []interface{}{
			map[string]interface{}{
				"kind":        "allow",
				"permissions": []interface{}{"stack:read"},
			},
			map[string]interface{}{
				"kind": "PermissionDescriptorCompose",
				"permissionDescriptors": []interface{}{
					"046f4b97-ed29-43d3-a09a-2c5d8d0e44e0",
				},
			},
		},
	}
}
func composeInsideGroupWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"__type":      "PermissionDescriptorAllow",
				"permissions": []interface{}{"stack:read"},
			},
			map[string]interface{}{
				"__type": "PermissionDescriptorCompose",
				"permissionDescriptors": []interface{}{
					"046f4b97-ed29-43d3-a09a-2c5d8d0e44e0",
				},
			},
		},
	}
}

// ----------------------------------------------------------------------------
// Forward direction: kind → wire
// ----------------------------------------------------------------------------

func TestPermissionsKindToWire(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		{"flatAllow", flatAllowKind(), flatAllowWire()},
		{"scopedAllow", scopedAllowKind(), scopedAllowWire()},
		{"flatGroup", flatGroupKind(), flatGroupWire()},
		{"scopedGroup", scopedGroupKind(), scopedGroupWire()},
		{"mixedGroup", mixedGroupKind(), mixedGroupWire()},
		{"teamScopedAllow", teamScopedAllowKind(), teamScopedAllowWire()},
		{"bareCompose", bareComposeKind(), bareComposeWire()},
		{"ifThenElseDeeplyNested", ifThenElseDeeplyNestedKind(), ifThenElseDeeplyNestedWire()},
		{"selectMixedLiteralTypes", selectMixedLiteralTypesKind(), selectMixedLiteralTypesWire()},
		{"composeInsideGroup", composeInsideGroupKind(), composeInsideGroupWire()},
		{"conditionAndOfEquals", conditionAndOfEqualsKind(), conditionAndOfEqualsWire()},
		{"conditionMismatchedOperands", conditionMismatchedOperandsKind(), conditionMismatchedOperandsWire()},
		{"conditionWrappingCompose", conditionWrappingComposeKind(), conditionWrappingComposeWire()},
		{"conditionWrappingCondition", conditionWrappingConditionKind(), conditionWrappingConditionWire()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissionsKindToWire(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// stackOnAllow and insightsAccountOnAllow exercise the two non-environment
// entity types on a bare Allow, so every (entity-type × on-bearer) pair has
// at least one positive test.
func TestPermissionsKindToWire_StackOnAllow(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"kind":        "allow",
		"on":          map[string]interface{}{"stack": "stack-id-9"},
		"permissions": []interface{}{"stack:edit"},
	}
	want := map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionStack"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionStack",
				"identity": "stack-id-9",
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
	}
	got, err := permissionsKindToWire(in)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestPermissionsKindToWire_InsightsAccountOnAllow(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"kind":        "allow",
		"on":          map[string]interface{}{"insightsAccount": "acct-9"},
		"permissions": []interface{}{"insights-account:read"},
	}
	want := map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": "PermissionExpressionInsightsAccount"},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionInsightsAccount",
				"identity": "acct-9",
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"insights-account:read"},
		},
	}
	got, err := permissionsKindToWire(in)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// ----------------------------------------------------------------------------
// Reverse direction: wire → kind
// ----------------------------------------------------------------------------

func TestPermissionsWireToKind(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		{"flatAllow", flatAllowWire(), flatAllowKind()},
		{"scopedAllow", scopedAllowWire(), scopedAllowKind()},
		{"flatGroup", flatGroupWire(), flatGroupKind()},
		{"scopedGroup", scopedGroupWire(), scopedGroupKind()},
		{"mixedGroup", mixedGroupWire(), mixedGroupKind()},
		{"teamScopedAllow", teamScopedAllowWire(), teamScopedAllowKind()},
		{"bareCompose", bareComposeWire(), bareComposeKind()},
		{"ifThenElseDeeplyNested", ifThenElseDeeplyNestedWire(), ifThenElseDeeplyNestedKind()},
		{"selectMixedLiteralTypes", selectMixedLiteralTypesWire(), selectMixedLiteralTypesKind()},
		{"composeInsideGroup", composeInsideGroupWire(), composeInsideGroupKind()},
		{"conditionAndOfEquals", conditionAndOfEqualsWire(), conditionAndOfEqualsKind()},
		{"conditionMismatchedOperands", conditionMismatchedOperandsWire(), conditionMismatchedOperandsKind()},
		{"conditionWrappingCompose", conditionWrappingComposeWire(), conditionWrappingComposeKind()},
		{"conditionWrappingCondition", conditionWrappingConditionWire(), conditionWrappingConditionKind()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissionsWireToKind(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ----------------------------------------------------------------------------
// Round-trip: kind → wire → kind = original.
// ----------------------------------------------------------------------------

func TestPermissionsKindWireRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		kind map[string]interface{}
	}{
		{"flatAllow", flatAllowKind()},
		{"scopedAllow", scopedAllowKind()},
		{"flatGroup", flatGroupKind()},
		{"scopedGroup", scopedGroupKind()},
		{"mixedGroup", mixedGroupKind()},
		{"teamScopedAllow", teamScopedAllowKind()},
		{"bareCompose", bareComposeKind()},
		{"ifThenElseDeeplyNested", ifThenElseDeeplyNestedKind()},
		{"selectMixedLiteralTypes", selectMixedLiteralTypesKind()},
		{"composeInsideGroup", composeInsideGroupKind()},
		{"conditionAndOfEquals", conditionAndOfEqualsKind()},
		{"conditionMismatchedOperands", conditionMismatchedOperandsKind()},
		{"conditionWrappingCompose", conditionWrappingComposeKind()},
		{"conditionWrappingCondition", conditionWrappingConditionKind()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wire, err := permissionsKindToWire(tc.kind)
			require.NoError(t, err)
			back, err := permissionsWireToKind(wire)
			require.NoError(t, err)
			assert.Equal(t, tc.kind, back)
		})
	}
}

// Reverse round-trip: starting from a wire fixture, round-trip through kind
// and back must preserve the wire shape too. This catches asymmetries where
// the forward and reverse functions disagree on canonical form.
func TestPermissionsWireKindRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		wire map[string]interface{}
	}{
		{"flatAllow", flatAllowWire()},
		{"scopedAllow", scopedAllowWire()},
		{"flatGroup", flatGroupWire()},
		{"scopedGroup", scopedGroupWire()},
		{"mixedGroup", mixedGroupWire()},
		{"teamScopedAllow", teamScopedAllowWire()},
		{"bareCompose", bareComposeWire()},
		{"ifThenElseDeeplyNested", ifThenElseDeeplyNestedWire()},
		{"selectMixedLiteralTypes", selectMixedLiteralTypesWire()},
		{"composeInsideGroup", composeInsideGroupWire()},
		{"conditionAndOfEquals", conditionAndOfEqualsWire()},
		{"conditionMismatchedOperands", conditionMismatchedOperandsWire()},
		{"conditionWrappingCompose", conditionWrappingComposeWire()},
		{"conditionWrappingCondition", conditionWrappingConditionWire()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kind, err := permissionsWireToKind(tc.wire)
			require.NoError(t, err)
			back, err := permissionsKindToWire(kind)
			require.NoError(t, err)
			assert.Equal(t, tc.wire, back)
		})
	}
}

// ----------------------------------------------------------------------------
// Forward error paths.
// ----------------------------------------------------------------------------

func TestPermissionsKindToWire_Errors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		in          map[string]interface{}
		wantErrFrag string
	}{
		{
			name:        "missing kind",
			in:          map[string]interface{}{"permissions": []interface{}{"stack:read"}},
			wantErrFrag: "kind",
		},
		{
			name:        "kind is not a string",
			in:          map[string]interface{}{"kind": 42, "permissions": []interface{}{"stack:read"}},
			wantErrFrag: "kind",
		},
		{
			name:        "unknown kind",
			in:          map[string]interface{}{"kind": "bogus", "permissions": []interface{}{"stack:read"}},
			wantErrFrag: "bogus",
		},
		{
			name: "on is empty map",
			in: map[string]interface{}{
				"kind":        "allow",
				"on":          map[string]interface{}{},
				"permissions": []interface{}{"stack:read"},
			},
			wantErrFrag: "on",
		},
		{
			name: "on has multiple keys",
			in: map[string]interface{}{
				"kind":        "allow",
				"on":          map[string]interface{}{"environment": "e", "stack": "s"},
				"permissions": []interface{}{"stack:read"},
			},
			wantErrFrag: "on",
		},
		{
			name: "on has unknown entity key",
			in: map[string]interface{}{
				"kind":        "allow",
				"on":          map[string]interface{}{"unknownEntity": "x"},
				"permissions": []interface{}{"stack:read"},
			},
			wantErrFrag: "unknownEntity",
		},
		{
			name: "on value is not a string",
			in: map[string]interface{}{
				"kind":        "allow",
				"on":          map[string]interface{}{"environment": 42},
				"permissions": []interface{}{"stack:read"},
			},
			wantErrFrag: "string",
		},
		{
			name: "on is not a map",
			in: map[string]interface{}{
				"kind":        "allow",
				"on":          "environment",
				"permissions": []interface{}{"stack:read"},
			},
			wantErrFrag: "on",
		},
		{
			name:        "allow with no permissions",
			in:          map[string]interface{}{"kind": "allow"},
			wantErrFrag: "permissions",
		},
		{
			name: "allow with non-list permissions",
			in: map[string]interface{}{
				"kind":        "allow",
				"permissions": "stack:read",
			},
			wantErrFrag: "permissions",
		},
		{
			name:        "group with no entries",
			in:          map[string]interface{}{"kind": "group"},
			wantErrFrag: "entries",
		},
		{
			name: "group with non-list entries",
			in: map[string]interface{}{
				"kind":    "group",
				"entries": "not a list",
			},
			wantErrFrag: "entries",
		},
		{
			name: "group with bad entry",
			in: map[string]interface{}{
				"kind": "group",
				"entries": []interface{}{
					map[string]interface{}{"kind": "bogus"},
				},
			},
			wantErrFrag: "bogus",
		},
		{
			name: "on: on pass-through kind",
			in: map[string]interface{}{
				"kind":                  "PermissionDescriptorCompose",
				"on":                    map[string]interface{}{"environment": "e"},
				"permissionDescriptors": []interface{}{"role-1"},
			},
			wantErrFrag: "on:",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := permissionsKindToWire(tc.in)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrFrag)
		})
	}
}

// ----------------------------------------------------------------------------
// Reverse error paths.
// ----------------------------------------------------------------------------

func TestPermissionsWireToKind_Errors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		in          map[string]interface{}
		wantErrFrag string
	}{
		{
			name:        "missing __type",
			in:          map[string]interface{}{"permissions": []interface{}{"stack:read"}},
			wantErrFrag: "__type",
		},
		{
			name: "Group with non-list entries",
			in: map[string]interface{}{
				"__type":  "PermissionDescriptorGroup",
				"entries": "not a list",
			},
			wantErrFrag: "entries",
		},
		{
			name: "Allow with non-list permissions",
			in: map[string]interface{}{
				"__type":      "PermissionDescriptorAllow",
				"permissions": "stack:read",
			},
			wantErrFrag: "permissions",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := permissionsWireToKind(tc.in)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrFrag)
		})
	}
}

// ----------------------------------------------------------------------------
// renameKey — blind recursive key rename.
// ----------------------------------------------------------------------------

func TestRenameKey_FlatObject(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{"__type": "X", "other": 1}
	got := renameKey(in, "__type", "kind")
	want := map[string]interface{}{"kind": "X", "other": 1}
	assert.Equal(t, want, got)
}

func TestRenameKey_NestedObject(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"__type": "Outer",
		"inner": map[string]interface{}{
			"__type": "Inner",
			"deeper": map[string]interface{}{
				"__type": "Deepest",
			},
		},
	}
	got := renameKey(in, "__type", "kind")
	want := map[string]interface{}{
		"kind": "Outer",
		"inner": map[string]interface{}{
			"kind": "Inner",
			"deeper": map[string]interface{}{
				"kind": "Deepest",
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestRenameKey_ArrayOfObjects(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"__type": "Outer",
		"entries": []interface{}{
			map[string]interface{}{"__type": "A"},
			map[string]interface{}{"__type": "B", "nested": map[string]interface{}{"__type": "C"}},
		},
	}
	got := renameKey(in, "__type", "kind")
	want := map[string]interface{}{
		"kind": "Outer",
		"entries": []interface{}{
			map[string]interface{}{"kind": "A"},
			map[string]interface{}{"kind": "B", "nested": map[string]interface{}{"kind": "C"}},
		},
	}
	assert.Equal(t, want, got)
}

func TestRenameKey_NoMatchUnchanged(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"foo": "bar",
		"baz": map[string]interface{}{"qux": 1},
	}
	got := renameKey(in, "__type", "kind")
	assert.Equal(t, in, got)
}

func TestRenameKey_Inverse(t *testing.T) {
	t.Parallel()
	// Round trip: rename A→B then B→A returns original
	original := map[string]interface{}{
		"__type": "X",
		"sub":    map[string]interface{}{"__type": "Y"},
	}
	forward := renameKey(original, "__type", "kind")
	back := renameKey(forward, "kind", "__type")
	assert.Equal(t, original, back)
}

func TestRenameKey_PreservesNonStringValues(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"__type":  "X",
		"number":  42,
		"boolean": true,
		"null":    nil,
		"array":   []interface{}{1, "two", true},
	}
	got := renameKey(in, "__type", "kind")
	want := map[string]interface{}{
		"kind":    "X",
		"number":  42,
		"boolean": true,
		"null":    nil,
		"array":   []interface{}{1, "two", true},
	}
	assert.Equal(t, want, got)
}

// ----------------------------------------------------------------------------
// Backwards-compat: the prior helpers (in the same PR's earlier commits)
// emitted Group(entries: [Condition(...)]) — a single-entry Group wrapping
// a Condition. The reverse translator must accept that shape and produce a
// single-entry group with an `on:`-modified entry. Refresh would then detect
// drift against the new flat helper output, but Read itself must succeed.
// ----------------------------------------------------------------------------

func TestPermissionsWireToKind_LegacySingleEntryGroup(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionEqual",
					"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
					"right": map[string]interface{}{
						"__type":   "PermissionLiteralExpressionEnvironment",
						"identity": "env-uuid-1",
					},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:read"},
				},
			},
		},
	}
	want := map[string]interface{}{
		"kind": "group",
		"entries": []interface{}{
			map[string]interface{}{
				"kind":        "allow",
				"on":          map[string]interface{}{"environment": "env-uuid-1"},
				"permissions": []interface{}{"environment:read"},
			},
		},
	}
	got, err := permissionsWireToKind(in)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
