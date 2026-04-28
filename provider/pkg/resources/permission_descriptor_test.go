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
			name:        "unknown __type",
			in:          map[string]interface{}{"__type": "PermissionDescriptorBogus"},
			wantErrFrag: "PermissionDescriptorBogus",
		},
		{
			name: "Condition with non-Equal condition",
			in: map[string]interface{}{
				"__type":    "PermissionDescriptorCondition",
				"condition": map[string]interface{}{"__type": "PermissionExpressionAnd"},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"stack:read"},
				},
			},
			wantErrFrag: "PermissionExpressionAnd",
		},
		{
			name: "Equal with mismatched left/right entity types",
			in: map[string]interface{}{
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
			},
			wantErrFrag: "mismatched",
		},
		{
			name: "Equal with right missing identity",
			in: map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionEqual",
					"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
					"right":  map[string]interface{}{"__type": "PermissionLiteralExpressionEnvironment"},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"stack:read"},
				},
			},
			wantErrFrag: "identity",
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
		{
			name: "Condition wraps another Condition (nested scoping)",
			in: map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionEqual",
					"left":   map[string]interface{}{"__type": "PermissionExpressionStack"},
					"right": map[string]interface{}{
						"__type":   "PermissionLiteralExpressionStack",
						"identity": "s1",
					},
				},
				"subNode": map[string]interface{}{
					"__type": "PermissionDescriptorCondition",
					"condition": map[string]interface{}{
						"__type": "PermissionExpressionEqual",
						"left":   map[string]interface{}{"__type": "PermissionExpressionEnvironment"},
						"right": map[string]interface{}{
							"__type":   "PermissionLiteralExpressionEnvironment",
							"identity": "e1",
						},
					},
					"subNode": map[string]interface{}{
						"__type":      "PermissionDescriptorAllow",
						"permissions": []interface{}{"environment:read"},
					},
				},
			},
			wantErrFrag: "nested",
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
