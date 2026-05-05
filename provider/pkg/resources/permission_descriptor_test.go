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
// Wire format and SDK boundary both use `__type` (Pulumi's Python SDK
// preserves `__`-prefixed keys as of pulumi/pulumi#22834, pinned via the
// Python SDK's runtime requirement). Each fixture returns a freshly
// allocated map so tests can mutate without affecting other tests.
//
// Variants exercised:
//   - Allow         — leaf grant.
//   - Group         — composition.
//   - Condition     — gate on an Equal expression. Top-level instances are
//                     wrapped by permissionsForAPI for the Cloud UI.
//   - Compose       — references other roles by ID.
//   - And/Or/Not    — boolean operators.
//   - IfThenElse    — variant the provider has no specific knowledge of;
//                     pass-through proves the provider is variant-agnostic.
// ----------------------------------------------------------------------------

func flatAllow() map[string]interface{} {
	return map[string]interface{}{
		"__type":      "PermissionDescriptorAllow",
		"permissions": []interface{}{"stack:read"},
	}
}

func flatGroup() map[string]interface{} {
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

// scopedCondition: the post-collapse Condition shape the helpers emit.
// permissionsForAPI wraps this in a single-entry Group for the Cloud UI;
// permissionsFromAPI collapses the wrap on Read when the user's prior
// input was not a Group.
func scopedCondition() map[string]interface{} {
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

// compose: the customer-import case — a top-level
// PermissionDescriptorCompose tree the provider has no specific knowledge
// of. Pass-through round-trip proves the provider is variant-agnostic.
func compose() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorCompose",
		"permissionDescriptors": []interface{}{
			"role-id-base-a",
			"role-id-base-b",
		},
	}
}

// andCondition: a Condition gated on And(Equal, Equal). Both Equal operands
// target the same team, which is tautological but valid; the point is to
// exercise a non-Equal boolean operator end to end.
func andCondition() map[string]interface{} {
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

// ifThenElse: a fictional descriptor the provider has no knowledge of,
// structured to look like a plausible future Cloud variant. Round-tripping
// this fixture proves the provider is structurally agnostic — adding a new
// variant to Pulumi Cloud requires zero provider changes.
func ifThenElse() map[string]interface{} {
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
// validatePermissions: top-level descriptor must carry `__type`.
// ----------------------------------------------------------------------------

func TestValidatePermissions(t *testing.T) {
	t.Parallel()
	t.Run("missing top-level __type", func(t *testing.T) {
		t.Parallel()
		err := validatePermissions(map[string]interface{}{
			"permissions": []interface{}{"stack:read"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "__type")
	})

	t.Run("present top-level __type", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, validatePermissions(flatAllow()))
	})
}

// ----------------------------------------------------------------------------
// API-boundary wrap: permissionsForAPI is what Create/Update call.
// It wraps the input in a single-entry Group only when the top-level shape
// is a Condition. Pulumi Cloud's role-detail UI 500s on a bare top-level
// Condition; the wrap fixes it.
// ----------------------------------------------------------------------------

func TestPermissionsForAPI(t *testing.T) {
	t.Parallel()
	wrappedScopedCondition := func() map[string]interface{} {
		return map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": []interface{}{scopedCondition()},
		}
	}
	wrappedAndCondition := func() map[string]interface{} {
		return map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": []interface{}{andCondition()},
		}
	}

	cases := []struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		// Wraps when the top-level shape is a Condition.
		{"scopedCondition wraps", scopedCondition(), wrappedScopedCondition()},
		{"andCondition wraps", andCondition(), wrappedAndCondition()},

		// Pass-through — top-level is already Allow / Group / Compose / IfThenElse.
		{"flatAllow no wrap", flatAllow(), flatAllow()},
		{"flatGroup no wrap", flatGroup(), flatGroup()},
		{"compose no wrap", compose(), compose()},
		{"ifThenElse no wrap", ifThenElse(), ifThenElse()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissionsForAPI(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ----------------------------------------------------------------------------
// permissionsFromAPI: the API-boundary wrap is reversed on Read iff the
// user's prior input wasn't already a Group.
// ----------------------------------------------------------------------------

// TestPermissionsFromAPI_PassThrough covers shapes whose Read is independent
// of any prior-input gating — the collapse heuristic is irrelevant because
// the top-level shape isn't a single-entry Group(Condition).
func TestPermissionsFromAPI_PassThrough(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
	}{
		{"flatAllow", flatAllow()},
		{"flatGroup", flatGroup()},
		{"compose", compose()},
		{"andCondition (bare; no wrap)", andCondition()},
		{"ifThenElse", ifThenElse()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissionsFromAPI(tc.in, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.in, got)
		})
	}
}

// TestPermissionsFromAPI_CollapseHeuristic exercises the gating rule on
// the API-boundary wrap collapse. The same wire shape — a single-entry
// Group whose only entry is a Condition — is interpreted differently
// depending on the user's prior input:
//   - prior == Condition (or empty): collapse to the entry. This matches
//     the helpers' output and produces clean state for `pulumi import`.
//   - prior == Group: preserve. The user explicitly wrote a single-entry
//     Group of Condition, so we hand it back the same way.
func TestPermissionsFromAPI_CollapseHeuristic(t *testing.T) {
	t.Parallel()
	wrap := map[string]interface{}{
		"__type":  "PermissionDescriptorGroup",
		"entries": []interface{}{scopedCondition()},
	}

	t.Run("nil prior collapses to Condition", func(t *testing.T) {
		t.Parallel()
		got, err := permissionsFromAPI(wrap, nil)
		require.NoError(t, err)
		assert.Equal(t, scopedCondition(), got)
	})

	t.Run("prior Condition collapses to Condition", func(t *testing.T) {
		t.Parallel()
		got, err := permissionsFromAPI(wrap, scopedCondition())
		require.NoError(t, err)
		assert.Equal(t, scopedCondition(), got)
	})

	t.Run("prior Group preserves Group(Condition)", func(t *testing.T) {
		t.Parallel()
		groupOfCondition := map[string]interface{}{
			"__type":  "PermissionDescriptorGroup",
			"entries": []interface{}{scopedCondition()},
		}
		got, err := permissionsFromAPI(wrap, groupOfCondition)
		require.NoError(t, err)
		assert.Equal(t, groupOfCondition, got)
	})

	t.Run("multi-entry Group never collapses", func(t *testing.T) {
		t.Parallel()
		multi := map[string]interface{}{
			"__type": "PermissionDescriptorGroup",
			"entries": []interface{}{
				scopedCondition(),
				flatAllow(),
			},
		}
		got, err := permissionsFromAPI(multi, nil)
		require.NoError(t, err)
		assert.Equal(t, "PermissionDescriptorGroup", got["__type"])
		entries := got["entries"].([]interface{})
		assert.Len(t, entries, 2)
	})

	t.Run("single-entry Group of non-Condition never collapses", func(t *testing.T) {
		t.Parallel()
		nonCondition := map[string]interface{}{
			"__type": "PermissionDescriptorGroup",
			"entries": []interface{}{
				flatAllow(),
			},
		}
		got, err := permissionsFromAPI(nonCondition, nil)
		require.NoError(t, err)
		assert.Equal(t, "PermissionDescriptorGroup", got["__type"])
	})
}

// ----------------------------------------------------------------------------
// Round-trip: every shape the customer can author survives Create + Read
// without drift.
//
// On Create/Update permissionsForAPI adds the single-entry-Group wrap for
// top-level Conditions. On Read permissionsFromAPI reverses the wrap iff
// the prior input wasn't a Group. This block tests the *whole* pipeline —
// input → wire → input = original — as the customer experiences it.
// ----------------------------------------------------------------------------

func TestPermissionsRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]interface{}
	}{
		{"flatAllow", flatAllow()},
		{"flatGroup", flatGroup()},
		{"scopedCondition", scopedCondition()},
		{"compose", compose()},
		{"andCondition", andCondition()},
		{"ifThenElse", ifThenElse()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wire, err := permissionsForAPI(tc.in)
			require.NoError(t, err)
			back, err := permissionsFromAPI(wire, tc.in)
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
		"__type":  "PermissionDescriptorGroup",
		"entries": []interface{}{scopedCondition()},
	}
	wire, err := permissionsForAPI(groupOfCondition)
	require.NoError(t, err)
	// Top-level was already Group, so no extra wrap.
	assert.Equal(t, "PermissionDescriptorGroup", wire["__type"])
	back, err := permissionsFromAPI(wire, groupOfCondition)
	require.NoError(t, err)
	assert.Equal(t, groupOfCondition, back)
}

// TestImportRepro_Compose proves the headline use case — a customer role
// authored in the Pulumi Cloud UI as a PermissionDescriptorCompose tree
// imports cleanly without provider changes.
func TestImportRepro_Compose(t *testing.T) {
	t.Parallel()
	// Wire shape as the Cloud REST API would return it.
	wire := compose()
	// On `pulumi import`, prior is empty.
	got, err := permissionsFromAPI(wire, nil)
	require.NoError(t, err)
	// The provider hands the Compose tree to the user's program verbatim.
	assert.Equal(t, compose(), got)
	// A subsequent up round-trips cleanly.
	wire2, err := permissionsForAPI(got)
	require.NoError(t, err)
	assert.Equal(t, compose(), wire2)
}

// ----------------------------------------------------------------------------
// Property tests. The provider passes the descriptor through verbatim
// (modulo the top-level Condition wrap). These verify that property holds
// even on deeply-nested expressions.
// ----------------------------------------------------------------------------

// TestRoundTrip_DeepNesting builds an arbitrary deep tree and verifies it
// round-trips losslessly.
func TestRoundTrip_DeepNesting(t *testing.T) {
	t.Parallel()
	deep := map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionAnd",
					"left": map[string]interface{}{
						"__type": "PermissionExpressionOr",
						"left": map[string]interface{}{
							"__type": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"__type": "PermissionExpressionStack",
							},
							"right": map[string]interface{}{
								"__type":   "PermissionLiteralExpressionStack",
								"identity": "s1",
							},
						},
						"right": map[string]interface{}{
							"__type": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"__type": "PermissionExpressionStack",
							},
							"right": map[string]interface{}{
								"__type":   "PermissionLiteralExpressionStack",
								"identity": "s2",
							},
						},
					},
					"right": map[string]interface{}{
						"__type": "PermissionExpressionNot",
						"operand": map[string]interface{}{
							"__type": "PermissionExpressionEqual",
							"left": map[string]interface{}{
								"__type": "PermissionExpressionEnvironment",
							},
							"right": map[string]interface{}{
								"__type":   "PermissionLiteralExpressionEnvironment",
								"identity": "env-evil",
							},
						},
					},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"stack:edit"},
				},
			},
		},
	}
	wire, err := permissionsForAPI(deep)
	require.NoError(t, err)
	assert.Equal(t, "PermissionDescriptorGroup", wire["__type"])
	back, err := permissionsFromAPI(wire, deep)
	require.NoError(t, err)
	assert.Equal(t, deep, back)
}

// TestRoundTrip_IdentityValueLooksLikeDiscriminator pins down that the
// provider treats descriptor *values* as opaque. An identity field whose
// string value is "__type" must survive untouched.
func TestRoundTrip_IdentityValueLooksLikeDiscriminator(t *testing.T) {
	t.Parallel()
	in := map[string]interface{}{
		"__type": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left": map[string]interface{}{
				"__type": "PermissionExpressionStack",
			},
			"right": map[string]interface{}{
				"__type":   "PermissionLiteralExpressionStack",
				"identity": "__type", // value that looks like a discriminator key
			},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"__type"}, // ditto
		},
	}
	wire, err := permissionsForAPI(in)
	require.NoError(t, err)
	back, err := permissionsFromAPI(wire, in)
	require.NoError(t, err)
	assert.Equal(t, in, back)
}
