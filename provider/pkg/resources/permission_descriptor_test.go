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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Cardinal shape fixtures.
//
// Each fixture pair (kind, wire) round-trips through the translator. Each
// function returns a freshly-allocated map so tests can mutate without
// affecting other tests.
//
// The translator has no descriptor-variant knowledge; these fixtures
// exercise:
//   - Allow         — leaf grant.
//   - Group         — composition.
//   - Condition     — gate on an Equal expression. Top-level instances are
//                     wrapped by permissionsKindToWireForAPI.
//   - Compose       — references other roles by ID. Pass-through only.
//   - And/Or/Not    — boolean operators. Pass-through only.
//   - IfThenElse    — variant we don't understand structurally; pass-through
//                     proves the translator is variant-agnostic.
// ----------------------------------------------------------------------------

func flatAllowKind() map[string]interface{} {
	return map[string]interface{}{
		"kind":        "PermissionDescriptorAllow",
		"permissions": []interface{}{"stack:read"},
	}
}
func flatAllowWire() map[string]interface{} {
	return map[string]interface{}{
		"__type":      "PermissionDescriptorAllow",
		"permissions": []interface{}{"stack:read"},
	}
}

func flatGroupKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"kind":        "PermissionDescriptorAllow",
				"permissions": []interface{}{"stack:read"},
			},
			map[string]interface{}{
				"kind":        "PermissionDescriptorAllow",
				"permissions": []interface{}{"stack:edit"},
			},
		},
	}
}
func flatGroupWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"__type":      "PermissionDescriptorAllow",
				"permissions": []interface{}{"stack:read"},
			},
			map[string]interface{}{
				"__type":      "PermissionDescriptorAllow",
				"permissions": []interface{}{"stack:edit"},
			},
		},
	}
}

// scopedConditionKind / scopedConditionWire: the post-collapse Condition shape
// the helpers emit. permissionsKindToWireForAPI wraps this in a single-entry
// Group for the Cloud UI; permissionsWireToKind collapses the wrap on Read
// when the user's prior input was not a Group.
func scopedConditionKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionEqual",
			"left": map[string]interface{}{
				"kind": "PermissionExpressionEnvironment",
			},
			"right": map[string]interface{}{
				"kind":     "PermissionLiteralExpressionEnvironment",
				"identity": "env-uuid-1",
			},
		},
		"subNode": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"environment:read"},
		},
	}
}
func scopedConditionWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left": map[string]interface{}{
				"__type": "PermissionExpressionEnvironment",
			},
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

// composeKind / composeWire: the customer-import case (Webflow). The provider
// has no Compose-specific code; it passes through verbatim.
func composeKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCompose",
		"permissionDescriptors": []interface{}{
			"role-id-base-a",
			"role-id-base-b",
		},
	}
}
func composeWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCompose",
		"permissionDescriptors": []interface{}{
			"role-id-base-a",
			"role-id-base-b",
		},
	}
}

// andConditionKind / andConditionWire: a Condition gated on And(Equal, Equal).
// Both Equal operands target the same team, which is tautological but valid;
// the point is to exercise a non-Equal boolean operator end to end.
func andConditionKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionAnd",
			"left": map[string]interface{}{
				"kind": "PermissionExpressionEqual",
				"left": map[string]interface{}{
					"kind": "PermissionExpressionTeam",
				},
				"right": map[string]interface{}{
					"kind":     "PermissionLiteralExpressionTeam",
					"identity": "team-a",
				},
			},
			"right": map[string]interface{}{
				"kind": "PermissionExpressionEqual",
				"left": map[string]interface{}{
					"kind": "PermissionExpressionTeam",
				},
				"right": map[string]interface{}{
					"kind":     "PermissionLiteralExpressionTeam",
					"identity": "team-b",
				},
			},
		},
		"subNode": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
	}
}
func andConditionWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionAnd",
			"left": map[string]interface{}{
				"__type": "PermissionExpressionEqual",
				"left": map[string]interface{}{
					"__type": "PermissionExpressionTeam",
				},
				"right": map[string]interface{}{
					"__type":   "PermissionLiteralExpressionTeam",
					"identity": "team-a",
				},
			},
			"right": map[string]interface{}{
				"__type": "PermissionExpressionEqual",
				"left": map[string]interface{}{
					"__type": "PermissionExpressionTeam",
				},
				"right": map[string]interface{}{
					"__type":   "PermissionLiteralExpressionTeam",
					"identity": "team-b",
				},
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
	}
}

// ifThenElseKind / ifThenElseWire: a fictional descriptor the provider has
// no knowledge of, structured to look like a plausible future Cloud variant.
// Round-trips this fixture proves the translator is structurally agnostic
// — adding a new variant to Pulumi Cloud requires zero provider changes.
func ifThenElseKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "PermissionDescriptorIfThenElse",
		"if": map[string]interface{}{
			"kind": "PermissionExpressionTeam",
		},
		"then": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
		"else": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:read"},
		},
	}
}
func ifThenElseWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorIfThenElse",
		"if": map[string]interface{}{
			"__type": "PermissionExpressionTeam",
		},
		"then": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:edit"},
		},
		"else": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:read"},
		},
	}
}

// ----------------------------------------------------------------------------
// Forward direction: kind → wire. Pure blind rename.
// ----------------------------------------------------------------------------

func TestPermissionsKindToWire(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		{"flatAllow", flatAllowKind(), flatAllowWire()},
		{"flatGroup", flatGroupKind(), flatGroupWire()},
		{"scopedCondition", scopedConditionKind(), scopedConditionWire()},
		{"compose", composeKind(), composeWire()},
		{"andCondition", andConditionKind(), andConditionWire()},
		{"ifThenElse", ifThenElseKind(), ifThenElseWire()},
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

// permissionsKindToWire validates the input is at least a map with a `kind`
// field at the top level. Anything below the top level can be arbitrary
// Pulumi Cloud wire grammar — we don't second-guess.
func TestPermissionsKindToWire_TopLevelValidation(t *testing.T) {
	t.Parallel()
	t.Run("missing top-level kind", func(t *testing.T) {
		t.Parallel()
		_, err := permissionsKindToWire(map[string]interface{}{
			"permissions": []interface{}{"stack:read"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kind")
	})
}

// permissionsKindToWire rejects any `__type` key in the input — defensive
// guard against users pasting raw wire format from the REST API docs. See
// the function docstring and pulumi/pulumi#22738 for the rationale.
func TestPermissionsKindToWire_RejectsUnderscoreType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
	}{
		{
			name: "top-level __type",
			in: map[string]interface{}{
				"__type":      "PermissionDescriptorAllow",
				"permissions": []interface{}{"stack:read"},
			},
		},
		{
			name: "nested __type in subNode",
			in: map[string]interface{}{
				"kind": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"kind": "PermissionExpressionEqual",
					"left": map[string]interface{}{"kind": "PermissionExpressionStack"},
					"right": map[string]interface{}{
						"kind":     "PermissionLiteralExpressionStack",
						"identity": "s",
					},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"stack:read"},
				},
			},
		},
		{
			name: "nested __type inside an entries list",
			in: map[string]interface{}{
				"kind": "PermissionDescriptorGroup",
				"entries": []interface{}{
					map[string]interface{}{
						"__type":      "PermissionDescriptorAllow",
						"permissions": []interface{}{"stack:read"},
					},
				},
			},
		},
		{
			name: "__type at the same level as kind (mixed paste)",
			in: map[string]interface{}{
				"kind":        "PermissionDescriptorAllow",
				"__type":      "PermissionDescriptorAllow",
				"permissions": []interface{}{"stack:read"},
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := permissionsKindToWire(tc.in)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "__type")
			assert.Contains(t, err.Error(), "kind",
				"error must point the user at the correct field name")
		})
	}
}

// ----------------------------------------------------------------------------
// API-boundary wrap: permissionsKindToWireForAPI is what Create/Update call.
// It runs permissionsKindToWire and then wraps the result in a single-entry
// Group only when the top-level wire shape is a Condition. Pulumi Cloud's
// role-detail UI 500s on a bare top-level Condition; the wrap fixes it.
// ----------------------------------------------------------------------------

func TestPermissionsKindToWireForAPI(t *testing.T) {
	t.Parallel()
	wrappedScopedCondition := func() map[string]interface{} {
		return map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": []interface{}{scopedConditionWire()},
		}
	}
	wrappedAndCondition := func() map[string]interface{} {
		return map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": []interface{}{andConditionWire()},
		}
	}
	wrappedIfThenElse := func() map[string]interface{} {
		// IfThenElse is not a Condition, so it must NOT be wrapped — the
		// wrap rule is precisely "top-level == PermissionDescriptorCondition".
		return ifThenElseWire()
	}

	cases := []struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		// Wraps when the top-level wire shape is a Condition.
		{"scopedCondition wraps", scopedConditionKind(), wrappedScopedCondition()},
		{"andCondition wraps", andConditionKind(), wrappedAndCondition()},

		// Pass-through — top-level is already Allow / Group / Compose / etc.
		{"flatAllow no wrap", flatAllowKind(), flatAllowWire()},
		{"flatGroup no wrap", flatGroupKind(), flatGroupWire()},
		{"compose no wrap", composeKind(), composeWire()},
		{"ifThenElse no wrap", ifThenElseKind(), wrappedIfThenElse()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissionsKindToWireForAPI(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ----------------------------------------------------------------------------
// Reverse direction: wire → kind. Pure blind rename, plus an optional top-
// level Group(Condition) → Condition collapse keyed off the user's prior
// input shape.
// ----------------------------------------------------------------------------

// TestPermissionsWireToKind covers shapes whose round-trip is independent of
// any prior-input gating — the collapse heuristic is irrelevant because the
// top-level wire shape isn't a single-entry Group(Condition).
func TestPermissionsWireToKind(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		{"flatAllow", flatAllowWire(), flatAllowKind()},
		{"flatGroup", flatGroupWire(), flatGroupKind()},
		{"compose", composeWire(), composeKind()},
		{"andCondition (bare; no wrap)", andConditionWire(), andConditionKind()},
		{"ifThenElse", ifThenElseWire(), ifThenElseKind()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissionsWireToKind(tc.in, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestPermissionsWireToKind_CollapseHeuristic exercises the gating rule on
// the API-boundary wrap collapse. The same wire shape — a single-entry
// Group whose only entry is a Condition — is interpreted differently
// depending on the user's prior input:
//   - prior == Condition (or empty): collapse to the entry. This matches
//     the helpers' output and produces clean state for `pulumi import`.
//   - prior == Group: preserve. The user explicitly wrote a single-entry
//     Group of Condition, so we hand it back the same way.
func TestPermissionsWireToKind_CollapseHeuristic(t *testing.T) {
	t.Parallel()
	wrap := map[string]interface{}{
		"__type":  "PermissionDescriptorGroup",
		"entries": []interface{}{scopedConditionWire()},
	}

	t.Run("nil prior collapses to Condition", func(t *testing.T) {
		t.Parallel()
		got, err := permissionsWireToKind(wrap, nil)
		require.NoError(t, err)
		assert.Equal(t, scopedConditionKind(), got)
	})

	t.Run("prior Condition collapses to Condition", func(t *testing.T) {
		t.Parallel()
		got, err := permissionsWireToKind(wrap, scopedConditionKind())
		require.NoError(t, err)
		assert.Equal(t, scopedConditionKind(), got)
	})

	t.Run("prior Group preserves Group(Condition)", func(t *testing.T) {
		t.Parallel()
		groupOfCondition := map[string]interface{}{
			"kind":    "PermissionDescriptorGroup",
			"entries": []interface{}{scopedConditionKind()},
		}
		got, err := permissionsWireToKind(wrap, groupOfCondition)
		require.NoError(t, err)
		assert.Equal(t, groupOfCondition, got)
	})

	t.Run("multi-entry Group never collapses", func(t *testing.T) {
		t.Parallel()
		multi := map[string]interface{}{
			"__type": "PermissionDescriptorGroup",
			"entries": []interface{}{
				scopedConditionWire(),
				flatAllowWire(),
			},
		}
		got, err := permissionsWireToKind(multi, nil)
		require.NoError(t, err)
		assert.Equal(t, "PermissionDescriptorGroup", got["kind"])
		entries := got["entries"].([]interface{})
		assert.Len(t, entries, 2)
	})

	t.Run("single-entry Group of non-Condition never collapses", func(t *testing.T) {
		t.Parallel()
		nonCondition := map[string]interface{}{
			"__type": "PermissionDescriptorGroup",
			"entries": []interface{}{
				flatAllowWire(),
			},
		}
		got, err := permissionsWireToKind(nonCondition, nil)
		require.NoError(t, err)
		assert.Equal(t, "PermissionDescriptorGroup", got["kind"])
	})
}

// ----------------------------------------------------------------------------
// Round-trip: every shape the customer can author survives a Create + Read
// without drift.
//
// The full path on Create/Update is permissionsKindToWireForAPI, which adds
// the single-entry-Group wrap for top-level Conditions. On Read,
// permissionsWireToKind reverses the wrap iff the prior input wasn't a Group.
// This block tests the *whole* pipeline — kind → wire → kind = original —
// as the customer experiences it.
// ----------------------------------------------------------------------------

func TestKindWireKindRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		kind map[string]interface{}
	}{
		{"flatAllow", flatAllowKind()},
		{"flatGroup", flatGroupKind()},
		{"scopedCondition", scopedConditionKind()},
		{"compose", composeKind()},
		{"andCondition", andConditionKind()},
		{"ifThenElse", ifThenElseKind()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wire, err := permissionsKindToWireForAPI(tc.kind)
			require.NoError(t, err)
			back, err := permissionsWireToKind(wire, tc.kind)
			require.NoError(t, err)
			assert.Equal(t, tc.kind, back)
		})
	}
}

// TestKindWireKindRoundTrip_GroupOfConditionPriorIsHonored proves that a
// customer who deliberately authors a single-entry Group of Condition gets
// it back unchanged on Read. The wrap fires on the inner Condition's
// passage to the wire (via permissionsKindToWire's rename only — the outer
// Group prevents permissionsKindToWireForAPI's wrap from firing); the
// collapse heuristic is suppressed by prior == Group.
func TestKindWireKindRoundTrip_GroupOfConditionPriorIsHonored(t *testing.T) {
	t.Parallel()
	groupOfCondition := map[string]interface{}{
		"kind":    "PermissionDescriptorGroup",
		"entries": []interface{}{scopedConditionKind()},
	}
	wire, err := permissionsKindToWireForAPI(groupOfCondition)
	require.NoError(t, err)
	// Top-level was already Group, so no extra wrap.
	assert.Equal(t, "PermissionDescriptorGroup", wire["__type"])
	back, err := permissionsWireToKind(wire, groupOfCondition)
	require.NoError(t, err)
	assert.Equal(t, groupOfCondition, back)
}

// TestImportRepro_Compose proves the headline use case — Webflow's role,
// authored in the Pulumi Cloud UI as a PermissionDescriptorCompose tree,
// imports cleanly without provider changes. This is the regression that
// motivated rewriting the translator.
func TestImportRepro_Compose(t *testing.T) {
	t.Parallel()
	// Wire shape as the Cloud REST API would return it.
	wire := composeWire()
	// On `pulumi import`, prior is empty.
	got, err := permissionsWireToKind(wire, nil)
	require.NoError(t, err)
	// The provider hands the Compose tree to the user's program verbatim
	// (modulo the __type → kind rename). No "unknown __type" error.
	assert.Equal(t, composeKind(), got)
	// A subsequent up round-trips cleanly.
	wire2, err := permissionsKindToWireForAPI(got)
	require.NoError(t, err)
	assert.Equal(t, composeWire(), wire2)
}

// ----------------------------------------------------------------------------
// Property tests. The blind rename is a structural deep-copy with the
// discriminator key swapped. These verify that property holds across two
// canaries that the structural code in earlier translator versions was apt
// to break: deeply-nested expressions, and identity-string values that
// happen to look like a discriminator.
// ----------------------------------------------------------------------------

// TestRoundTrip_DeepNesting builds an arbitrary deep tree and verifies it
// round-trips losslessly — every level swaps the discriminator key and
// recurses into every map and list value.
func TestRoundTrip_DeepNesting(t *testing.T) {
	t.Parallel()
	deep := map[string]interface{}{
		"kind": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"kind": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"kind": "PermissionExpressionAnd",
					"left": map[string]interface{}{
						"kind": "PermissionExpressionOr",
						"left": map[string]interface{}{
							"kind": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"kind": "PermissionExpressionStack",
							},
							"right": map[string]interface{}{
								"kind":     "PermissionLiteralExpressionStack",
								"identity": "s1",
							},
						},
						"right": map[string]interface{}{
							"kind": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"kind": "PermissionExpressionStack",
							},
							"right": map[string]interface{}{
								"kind":     "PermissionLiteralExpressionStack",
								"identity": "s2",
							},
						},
					},
					"right": map[string]interface{}{
						"kind": "PermissionExpressionNot",
						"operand": map[string]interface{}{
							"kind": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"kind": "PermissionExpressionEnvironment",
							},
							"right": map[string]interface{}{
								"kind":     "PermissionLiteralExpressionEnvironment",
								"identity": "env-evil",
							},
						},
					},
				},
				"subNode": map[string]interface{}{
					"kind":        "PermissionDescriptorAllow",
					"permissions": []interface{}{"stack:edit"},
				},
			},
		},
	}
	wire, err := permissionsKindToWireForAPI(deep)
	require.NoError(t, err)
	assert.NotContains(t, mustJSON(t, wire), `"kind":`,
		"after kind→wire, no `kind` key may remain anywhere")
	back, err := permissionsWireToKind(wire, deep)
	require.NoError(t, err)
	assert.Equal(t, deep, back)
}

// TestRoundTrip_IdentityValueLooksLikeDiscriminator pins down that the
// translator only renames map *keys*, not values. An identity field whose
// string value is "kind" or "__type" must survive untouched.
func TestRoundTrip_IdentityValueLooksLikeDiscriminator(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind": "PermissionExpressionEqual",
			"left": map[string]interface{}{
				"kind": "PermissionExpressionStack",
			},
			"right": map[string]interface{}{
				"kind":     "PermissionLiteralExpressionStack",
				"identity": "kind", // looks like a discriminator
			},
		},
		"subNode": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": []interface{}{"__type"}, // looks like a discriminator
		},
	}
	wire, err := permissionsKindToWireForAPI(in)
	require.NoError(t, err)
	back, err := permissionsWireToKind(wire, in)
	require.NoError(t, err)
	assert.Equal(t, in, back)
}

// mustJSON marshals a value to compact JSON for substring assertions.
func mustJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return strings.ReplaceAll(string(b), " ", "")
}
