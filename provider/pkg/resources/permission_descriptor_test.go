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
//                     wrapped by permissionsToWireForAPI.
//   - Compose       — references other roles by ID. Pass-through only.
//   - And/Or/Not    — boolean operators. Pass-through only.
//   - IfThenElse    — variant we don't understand structurally; pass-through
//                     proves the translator is variant-agnostic.
// ----------------------------------------------------------------------------

func flatAllowKind() map[string]interface{} {
	return map[string]interface{}{
		"discriminator": "PermissionDescriptorAllow",
		"permissions":   []interface{}{"stack:read"},
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
		"discriminator": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"discriminator": "PermissionDescriptorAllow",
				"permissions":   []interface{}{"stack:read"},
			},
			map[string]interface{}{
				"discriminator": "PermissionDescriptorAllow",
				"permissions":   []interface{}{"stack:edit"},
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
// the helpers emit. permissionsToWireForAPI wraps this in a single-entry
// Group for the Cloud UI; permissionsFromWire collapses the wrap on Read
// when the user's prior input was not a Group.
func scopedConditionKind() map[string]interface{} {
	return map[string]interface{}{
		"discriminator": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"discriminator": "PermissionExpressionEqual",
			"left": map[string]interface{}{
				"discriminator": "PermissionExpressionEnvironment",
			},
			"right": map[string]interface{}{
				"discriminator": "PermissionLiteralExpressionEnvironment",
				"identity":      "env-uuid-1",
			},
		},
		"subNode": map[string]interface{}{
			"discriminator": "PermissionDescriptorAllow",
			"permissions":   []interface{}{"environment:read"},
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
		"discriminator": "PermissionDescriptorCompose",
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
		"discriminator": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"discriminator": "PermissionExpressionAnd",
			"left": map[string]interface{}{
				"discriminator": "PermissionExpressionEqual",
				"left": map[string]interface{}{
					"discriminator": "PermissionExpressionTeam",
				},
				"right": map[string]interface{}{
					"discriminator": "PermissionLiteralExpressionTeam",
					"identity":      "team-a",
				},
			},
			"right": map[string]interface{}{
				"discriminator": "PermissionExpressionEqual",
				"left": map[string]interface{}{
					"discriminator": "PermissionExpressionTeam",
				},
				"right": map[string]interface{}{
					"discriminator": "PermissionLiteralExpressionTeam",
					"identity":      "team-b",
				},
			},
		},
		"subNode": map[string]interface{}{
			"discriminator": "PermissionDescriptorAllow",
			"permissions":   []interface{}{"stack:edit"},
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
		"discriminator": "PermissionDescriptorIfThenElse",
		"if": map[string]interface{}{
			"discriminator": "PermissionExpressionTeam",
		},
		"then": map[string]interface{}{
			"discriminator": "PermissionDescriptorAllow",
			"permissions":   []interface{}{"stack:edit"},
		},
		"else": map[string]interface{}{
			"discriminator": "PermissionDescriptorAllow",
			"permissions":   []interface{}{"stack:read"},
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
// Forward direction: discriminator → wire. Pure blind rename.
// ----------------------------------------------------------------------------

func TestPermissionsToWire(t *testing.T) {
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
			got, err := permissionsToWire(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// permissionsToWire validates the input is at least a map with a
// `discriminator` field at the top level. Anything below the top level can
// be arbitrary Pulumi Cloud wire grammar — we don't second-guess.
func TestPermissionsToWire_TopLevelValidation(t *testing.T) {
	t.Parallel()
	t.Run("missing top-level discriminator", func(t *testing.T) {
		t.Parallel()
		_, err := permissionsToWire(map[string]interface{}{
			"permissions": []interface{}{"stack:read"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "discriminator")
	})
}

// permissionsToWire rejects `__type` at every nesting level — Python's SDK
// strips `__`-prefixed keys from inputs (pulumi/pulumi#22738) at every
// level, not just the top, so a Python user pasting raw wire format
// would silently lose those discriminators at the language boundary
// and the role would be created with a malformed body. Surfacing the
// error here points the user at `discriminator` instead.
func TestPermissionsToWire_RejectsUnderscoreType(t *testing.T) {
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
			name: "nested __type inside subNode",
			in: map[string]interface{}{
				"discriminator": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"discriminator": "PermissionExpressionEqual",
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"stack:read"},
				},
			},
		},
		{
			name: "nested __type inside entries list",
			in: map[string]interface{}{
				"discriminator": "PermissionDescriptorGroup",
				"entries": []interface{}{
					map[string]interface{}{
						"__type":      "PermissionDescriptorAllow",
						"permissions": []interface{}{"stack:read"},
					},
				},
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := permissionsToWire(tc.in)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "__type")
			assert.Contains(t, err.Error(), "discriminator",
				"error must point the user at the correct field name")
		})
	}
}

// ----------------------------------------------------------------------------
// API-boundary wrap: permissionsToWireForAPI is what Create/Update call.
// It runs permissionsToWire and then wraps the result in a single-entry
// Group only when the top-level wire shape is a Condition. Pulumi Cloud's
// role-detail UI 500s on a bare top-level Condition; the wrap fixes it.
// ----------------------------------------------------------------------------

func TestPermissionsToWireForAPI(t *testing.T) {
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
			got, err := permissionsToWireForAPI(tc.in)
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
			got, err := permissionsFromWire(tc.in, nil)
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
		got, err := permissionsFromWire(wrap, nil)
		require.NoError(t, err)
		assert.Equal(t, scopedConditionKind(), got)
	})

	t.Run("prior Condition collapses to Condition", func(t *testing.T) {
		t.Parallel()
		got, err := permissionsFromWire(wrap, scopedConditionKind())
		require.NoError(t, err)
		assert.Equal(t, scopedConditionKind(), got)
	})

	t.Run("prior Group preserves Group(Condition)", func(t *testing.T) {
		t.Parallel()
		// User input uses `discriminator` at every level (recursive
		// translation). The collapse heuristic must not fire because
		// prior is a Group.
		groupOfCondition := map[string]interface{}{
			"discriminator": "PermissionDescriptorGroup",
			"entries":       []interface{}{scopedConditionKind()},
		}
		got, err := permissionsFromWire(wrap, groupOfCondition)
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
		got, err := permissionsFromWire(multi, nil)
		require.NoError(t, err)
		assert.Equal(t, "PermissionDescriptorGroup", got["discriminator"])
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
		got, err := permissionsFromWire(nonCondition, nil)
		require.NoError(t, err)
		assert.Equal(t, "PermissionDescriptorGroup", got["discriminator"])
	})
}

// ----------------------------------------------------------------------------
// Round-trip: every shape the customer can author survives a Create + Read
// without drift.
//
// The full path on Create/Update is permissionsToWireForAPI, which adds
// the single-entry-Group wrap for top-level Conditions. On Read,
// permissionsFromWire reverses the wrap iff the prior input wasn't a Group.
// This block tests the *whole* pipeline — input → wire → input = original —
// as the customer experiences it.
// ----------------------------------------------------------------------------

func TestPermissionsRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
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
			wire, err := permissionsToWireForAPI(tc.in)
			require.NoError(t, err)
			back, err := permissionsFromWire(wire, tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.in, back)
		})
	}
}

// TestPermissionsRoundTrip_GroupOfConditionPriorIsHonored proves that a
// customer who deliberately authors a single-entry Group of Condition gets
// it back unchanged on Read. The collapse heuristic is suppressed by
// prior == Group.
func TestPermissionsRoundTrip_GroupOfConditionPriorIsHonored(t *testing.T) {
	t.Parallel()
	groupOfCondition := map[string]interface{}{
		"discriminator": "PermissionDescriptorGroup",
		"entries":       []interface{}{scopedConditionKind()},
	}
	wire, err := permissionsToWireForAPI(groupOfCondition)
	require.NoError(t, err)
	// Top-level was already Group, so no extra wrap.
	assert.Equal(t, "PermissionDescriptorGroup", wire["__type"])
	back, err := permissionsFromWire(wire, groupOfCondition)
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
	got, err := permissionsFromWire(wire, nil)
	require.NoError(t, err)
	// The provider hands the Compose tree to the user's program verbatim
	// (modulo the __type → discriminator rename). No "unknown __type" error.
	assert.Equal(t, composeKind(), got)
	// A subsequent up round-trips cleanly.
	wire2, err := permissionsToWireForAPI(got)
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

// TestRoundTrip_DeepNesting builds an arbitrary deep tree using the SDK
// boundary's `discriminator` at every level (matching helper output and
// hand-author guidance) and verifies it round-trips losslessly through the
// recursive translator.
func TestRoundTrip_DeepNesting(t *testing.T) {
	t.Parallel()
	deep := map[string]interface{}{
		"discriminator": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"discriminator": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"discriminator": "PermissionExpressionAnd",
					"left": map[string]interface{}{
						"discriminator": "PermissionExpressionOr",
						"left": map[string]interface{}{
							"discriminator": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"discriminator": "PermissionExpressionStack",
							},
							"right": map[string]interface{}{
								"discriminator": "PermissionLiteralExpressionStack",
								"identity":      "s1",
							},
						},
						"right": map[string]interface{}{
							"discriminator": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"discriminator": "PermissionExpressionStack",
							},
							"right": map[string]interface{}{
								"discriminator": "PermissionLiteralExpressionStack",
								"identity":      "s2",
							},
						},
					},
					"right": map[string]interface{}{
						"discriminator": "PermissionExpressionNot",
						"operand": map[string]interface{}{
							"discriminator": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"discriminator": "PermissionExpressionEnvironment",
							},
							"right": map[string]interface{}{
								"discriminator": "PermissionLiteralExpressionEnvironment",
								"identity":      "env-evil",
							},
						},
					},
				},
				"subNode": map[string]interface{}{
					"discriminator": "PermissionDescriptorAllow",
					"permissions": []interface{}{"stack:edit"},
				},
			},
		},
	}
	wire, err := permissionsToWireForAPI(deep)
	require.NoError(t, err)
	assert.Equal(t, "PermissionDescriptorGroup", wire["__type"],
		"top must be the wire-format `__type`")
	_, hasTopDisc := wire["discriminator"]
	assert.False(t, hasTopDisc, "top `discriminator` must have been promoted to `__type`")
	back, err := permissionsFromWire(wire, deep)
	require.NoError(t, err)
	assert.Equal(t, deep, back)
}

// TestRoundTrip_IdentityValueLooksLikeDiscriminator pins down that the
// translator only renames map *keys*, not values. An identity field whose
// string value is "discriminator" must survive untouched.
func TestRoundTrip_IdentityValueLooksLikeDiscriminator(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"discriminator": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"discriminator": "PermissionExpressionEqual",
			"left": map[string]interface{}{
				"discriminator": "PermissionExpressionStack",
			},
			"right": map[string]interface{}{
				"discriminator": "PermissionLiteralExpressionStack",
				"identity":      "discriminator", // looks like a discriminator key
			},
		},
		"subNode": map[string]interface{}{
			"discriminator": "PermissionDescriptorAllow",
			"permissions":   []interface{}{"discriminator"}, // looks like a discriminator key
		},
	}
	wire, err := permissionsToWireForAPI(in)
	require.NoError(t, err)
	back, err := permissionsFromWire(wire, in)
	require.NoError(t, err)
	assert.Equal(t, in, back)
}
