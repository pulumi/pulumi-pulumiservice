// Copyright 2026, Pulumi Corporation.
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

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestPolicyGroupResourceID(t *testing.T) {
	assert.Equal(t, "org/group", policyGroupResourceID("org", "group"))
}

func TestStackReferencesEqual(t *testing.T) {
	a := []PolicyGroupStackReference{
		{Name: "s1", RoutingProject: "p1"},
		{Name: "s2", RoutingProject: "p2"},
	}
	b := []PolicyGroupStackReference{
		{Name: "s2", RoutingProject: "p2"},
		{Name: "s1", RoutingProject: "p1"},
	}
	assert.True(t, stackReferencesEqual(a, b), "order-independent equal")

	c := []PolicyGroupStackReference{
		{Name: "s1", RoutingProject: "different"},
		{Name: "s2", RoutingProject: "p2"},
	}
	assert.False(t, stackReferencesEqual(a, c), "different routing project")

	assert.True(t, stackReferencesEqual(nil, nil))
	assert.True(t, stackReferencesEqual([]PolicyGroupStackReference{}, nil))
}

func TestPolicyPackInputsEqualState(t *testing.T) {
	inputs := []PolicyGroupPolicyPackReferenceInput{
		{Name: "pp1", VersionTag: "1.0.0"},
		{Name: "pp2", VersionTag: "2.0.0"},
	}
	state := []PolicyGroupPolicyPackReference{
		{Name: "pp2", VersionTag: "2.0.0", Version: 7},
		{Name: "pp1", VersionTag: "1.0.0", Version: 3},
	}
	assert.True(t, policyPackInputsEqualState(inputs, state),
		"matches by (name, versionTag) regardless of order and server-derived version")

	differentTag := []PolicyGroupPolicyPackReference{
		{Name: "pp1", VersionTag: "1.0.1", Version: 3},
		{Name: "pp2", VersionTag: "2.0.0", Version: 7},
	}
	assert.False(t, policyPackInputsEqualState(inputs, differentTag))

	differentLen := []PolicyGroupPolicyPackReference{
		{Name: "pp1", VersionTag: "1.0.0"},
	}
	assert.False(t, policyPackInputsEqualState(inputs, differentLen))

	// Config is compared, with nil and empty treated as equal.
	emptyConfigInput := []PolicyGroupPolicyPackReferenceInput{
		{Name: "pp1", VersionTag: "1.0.0", Config: map[string]interface{}{}},
	}
	nilConfigState := []PolicyGroupPolicyPackReference{
		{Name: "pp1", VersionTag: "1.0.0"},
	}
	assert.True(t, policyPackInputsEqualState(emptyConfigInput, nilConfigState),
		"nil and empty config are equivalent")

	configInput := []PolicyGroupPolicyPackReferenceInput{
		{Name: "pp1", VersionTag: "1.0.0", Config: map[string]interface{}{"all": "mandatory"}},
	}
	sameConfigState := []PolicyGroupPolicyPackReference{
		{Name: "pp1", VersionTag: "1.0.0", Config: map[string]interface{}{"all": "mandatory"}},
	}
	assert.True(t, policyPackInputsEqualState(configInput, sameConfigState),
		"matching config compares equal")

	differentConfigState := []PolicyGroupPolicyPackReference{
		{Name: "pp1", VersionTag: "1.0.0", Config: map[string]interface{}{"all": "advisory"}},
	}
	assert.False(t, policyPackInputsEqualState(configInput, differentConfigState),
		"differing config compares unequal")
}

func TestHasParentAccount(t *testing.T) {
	tests := []struct {
		name     string
		account  string
		accounts []string
		expected bool
	}{
		{"child account with parent in list", "parent/child", []string{"parent", "other"}, true},
		{"child account without parent in list", "parent/child", []string{"other"}, false},
		{"parent account (no parent exists)", "parent", []string{"other"}, false},
		{"empty accounts list", "parent/child", nil, false},
		{"similar prefix but not parent", "parent-other/child", []string{"parent"}, false},
		{"exact match is not a parent", "parent", []string{"parent"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, hasParentAccount(tt.account, tt.accounts))
		})
	}
}

func TestBuildUpdateBatch(t *testing.T) {
	state := PolicyGroupState{
		Stacks: []PolicyGroupStackReference{
			{Name: "s1", RoutingProject: "p1"},
		},
		PolicyPacks: []PolicyGroupPolicyPackReference{
			{Name: "pp-old", VersionTag: "1.0.0", Version: 3},
		},
		Accounts: []string{"parent", "parent/child"},
	}
	// Parent stays in the new inputs — the child should not be removed even
	// though it isn't in the inputs list (it's auto-managed by the API).
	inputs := PolicyGroupInput{
		Stacks: []PolicyGroupStackReference{
			{Name: "s2", RoutingProject: "p2"},
		},
		PolicyPacks: []PolicyGroupPolicyPackReferenceInput{
			{Name: "pp-new", VersionTag: "2.0.0"},
		},
		Accounts: []string{"parent"},
	}
	batch := buildUpdateBatch(state, inputs)

	// Expected ops: remove s1, add s2, remove pp-old, add pp-new. No account
	// changes (child stays because parent stays).
	require.Len(t, batch, 4)

	var removedStack, addedStack *pulumiapi.StackReference
	var removedPack, addedPack *pulumiapi.PolicyPackMetadata
	for _, op := range batch {
		switch {
		case op.RemoveStack != nil:
			removedStack = op.RemoveStack
		case op.AddStack != nil:
			addedStack = op.AddStack
		case op.RemovePolicyPack != nil:
			removedPack = op.RemovePolicyPack
		case op.AddPolicyPack != nil:
			addedPack = op.AddPolicyPack
		case op.RemoveInsightsAccount != nil:
			t.Fatalf("should not remove child account whose parent stays: %s", op.RemoveInsightsAccount.Name)
		case op.AddInsightsAccount != nil:
			t.Fatalf("unexpected AddInsightsAccount: %s", op.AddInsightsAccount.Name)
		}
	}
	require.NotNil(t, removedStack)
	require.NotNil(t, addedStack)
	require.NotNil(t, removedPack)
	require.NotNil(t, addedPack)
	assert.Equal(t, "s1", removedStack.Name)
	assert.Equal(t, "s2", addedStack.Name)
	assert.Equal(t, "pp-old", removedPack.Name)
	assert.Equal(t, "pp-new", addedPack.Name)
}

func TestBuildUpdateBatch_NoChanges(t *testing.T) {
	state := PolicyGroupState{
		Stacks: []PolicyGroupStackReference{{Name: "s1", RoutingProject: "p1"}},
		PolicyPacks: []PolicyGroupPolicyPackReference{
			{Name: "pp1", VersionTag: "1.0.0", Version: 5},
		},
		Accounts: []string{"a1"},
	}
	inputs := PolicyGroupInput{
		Stacks: []PolicyGroupStackReference{{Name: "s1", RoutingProject: "p1"}},
		PolicyPacks: []PolicyGroupPolicyPackReferenceInput{
			// No `Version` here — inputs only carry versionTag.
			{Name: "pp1", VersionTag: "1.0.0"},
		},
		Accounts: []string{"a1"},
	}
	assert.Empty(t, buildUpdateBatch(state, inputs),
		"identical inputs (modulo server-derived version) produce no API ops")
}

func TestBuildUpdateBatch_RemovesChildWhenParentRemoved(t *testing.T) {
	state := PolicyGroupState{
		Accounts: []string{"parent", "parent/child"},
	}
	inputs := PolicyGroupInput{Accounts: nil}
	batch := buildUpdateBatch(state, inputs)

	require.Len(t, batch, 2)
	names := []string{}
	for _, op := range batch {
		require.NotNil(t, op.RemoveInsightsAccount)
		names = append(names, op.RemoveInsightsAccount.Name)
	}
	assert.ElementsMatch(t, []string{"parent", "parent/child"}, names)
}

func TestPolicyGroupCheck_DefaultsAndValidation(t *testing.T) {
	pg := &PolicyGroup{}

	t.Run("applies defaults", func(t *testing.T) {
		req := infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"name":             property.New("g"),
				"organizationName": property.New("org"),
			}),
		}
		resp, err := pg.Check(t.Context(), req)
		require.NoError(t, err)
		assert.Empty(t, resp.Failures)
		assert.Equal(t, "stacks", resp.Inputs.EntityType)
		assert.Equal(t, "audit", resp.Inputs.Mode)
	})

	t.Run("strips server-derived version from policy packs", func(t *testing.T) {
		req := infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"name":             property.New("g"),
				"organizationName": property.New("org"),
				"policyPacks": property.New(property.NewArray([]property.Value{
					property.New(property.NewMap(map[string]property.Value{
						"name":       property.New("pp1"),
						"version":    property.New(5.0),
						"versionTag": property.New("1.0.0"),
					})),
				})),
			}),
		}
		resp, err := pg.Check(t.Context(), req)
		require.NoError(t, err)
		assert.Empty(t, resp.Failures)
		require.Len(t, resp.Inputs.PolicyPacks, 1)
		assert.Equal(t, "pp1", resp.Inputs.PolicyPacks[0].Name)
		assert.Equal(t, "1.0.0", resp.Inputs.PolicyPacks[0].VersionTag)
		// The Input type has no Version field — the strip already happened
		// before the decode, so it cannot leak through.
	})

	t.Run("rejects invalid entityType", func(t *testing.T) {
		req := infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"name":             property.New("g"),
				"organizationName": property.New("org"),
				"entityType":       property.New("invalid"),
				"mode":             property.New("audit"),
			}),
		}
		resp, err := pg.Check(t.Context(), req)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Failures)
		assert.Equal(t, p.CheckFailure{
			Property: "entityType",
			Reason:   "entityType must be either 'stacks' or 'accounts'",
		}, resp.Failures[0])
	})

	t.Run("rejects invalid mode", func(t *testing.T) {
		req := infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"name":             property.New("g"),
				"organizationName": property.New("org"),
				"entityType":       property.New("stacks"),
				"mode":             property.New("preventive"),
			}),
		}
		resp, err := pg.Check(t.Context(), req)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Failures)
		assert.Equal(t, p.CheckFailure{
			Property: "mode",
			Reason:   "mode must be either 'audit' or 'preventative'",
		}, resp.Failures[0])
	})
}

func TestPolicyGroupDiff(t *testing.T) {
	pg := &PolicyGroup{}

	mkReq := func(state PolicyGroupState, inputs PolicyGroupInput) infer.DiffRequest[PolicyGroupInput, PolicyGroupState] {
		return infer.DiffRequest[PolicyGroupInput, PolicyGroupState]{State: state, Inputs: inputs}
	}

	t.Run("entityType change replaces", func(t *testing.T) {
		resp, err := pg.Diff(t.Context(), mkReq(
			PolicyGroupState{Name: "g", OrganizationName: "o", EntityType: "stacks", Mode: "audit"},
			PolicyGroupInput{Name: "g", OrganizationName: "o", EntityType: "accounts", Mode: "audit"},
		))
		require.NoError(t, err)
		assert.True(t, resp.HasChanges)
		assert.Equal(t, p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}, resp.DetailedDiff["entityType"])
	})

	t.Run("mode change replaces", func(t *testing.T) {
		resp, err := pg.Diff(t.Context(), mkReq(
			PolicyGroupState{Mode: "audit"},
			PolicyGroupInput{Mode: "preventative"},
		))
		require.NoError(t, err)
		assert.Equal(t, p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}, resp.DetailedDiff["mode"])
	})

	t.Run("order-independent stacks: no diff", func(t *testing.T) {
		stacks1 := []PolicyGroupStackReference{
			{Name: "a", RoutingProject: "p"}, {Name: "b", RoutingProject: "p"},
		}
		stacks2 := []PolicyGroupStackReference{
			{Name: "b", RoutingProject: "p"}, {Name: "a", RoutingProject: "p"},
		}
		resp, err := pg.Diff(t.Context(), mkReq(
			PolicyGroupState{Stacks: stacks1},
			PolicyGroupInput{Stacks: stacks2},
		))
		require.NoError(t, err)
		assert.False(t, resp.HasChanges)
	})

	t.Run("order-independent accounts: no diff", func(t *testing.T) {
		resp, err := pg.Diff(t.Context(), mkReq(
			PolicyGroupState{Accounts: []string{"b", "a"}},
			PolicyGroupInput{Accounts: []string{"a", "b"}},
		))
		require.NoError(t, err)
		assert.False(t, resp.HasChanges)
	})

	t.Run("policy pack input matches state ignoring server version", func(t *testing.T) {
		resp, err := pg.Diff(t.Context(), mkReq(
			PolicyGroupState{
				PolicyPacks: []PolicyGroupPolicyPackReference{
					{Name: "pp1", VersionTag: "1.0.0", Version: 42},
				},
			},
			PolicyGroupInput{
				PolicyPacks: []PolicyGroupPolicyPackReferenceInput{
					{Name: "pp1", VersionTag: "1.0.0"},
				},
			},
		))
		require.NoError(t, err)
		assert.False(t, resp.HasChanges, "version-only difference must not diff")
	})
}
