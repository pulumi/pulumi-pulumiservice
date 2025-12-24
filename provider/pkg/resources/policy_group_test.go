// Copyright 2016-2025, Pulumi Corporation.
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
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock client for PolicyGroup tests
type PolicyGroupClientMock struct {
	getPolicyGroupFunc    func() (*pulumiapi.PolicyGroup, error)
	createPolicyGroupFunc func(ctx context.Context, orgName, policyGroupName, entityType, mode string) error
}

func (c *PolicyGroupClientMock) ListPolicyGroups(ctx context.Context, orgName string) ([]pulumiapi.PolicyGroupSummary, error) {
	return nil, nil
}

func (c *PolicyGroupClientMock) GetPolicyGroup(ctx context.Context, orgName string, policyGroupName string) (*pulumiapi.PolicyGroup, error) {
	if c.getPolicyGroupFunc != nil {
		return c.getPolicyGroupFunc()
	}
	return nil, nil
}

func (c *PolicyGroupClientMock) CreatePolicyGroup(ctx context.Context, orgName, policyGroupName, entityType, mode string) error {
	if c.createPolicyGroupFunc != nil {
		return c.createPolicyGroupFunc(ctx, orgName, policyGroupName, entityType, mode)
	}
	return nil
}

func (c *PolicyGroupClientMock) UpdatePolicyGroup(ctx context.Context, orgName, policyGroupName string, req pulumiapi.UpdatePolicyGroupRequest) error {
	return nil
}

func (c *PolicyGroupClientMock) BatchUpdatePolicyGroup(ctx context.Context, orgName, policyGroupName string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
	return nil
}

func (c *PolicyGroupClientMock) DeletePolicyGroup(ctx context.Context, orgName, policyGroupName string) error {
	return nil
}

// TestPolicyGroup_Check_Defaults tests that Check applies default values when entityType and mode are not provided
func TestPolicyGroup_Check_Defaults(t *testing.T) {
	provider := PulumiServicePolicyGroupResource{
		Client: &PolicyGroupClientMock{},
	}

	// Create input without entityType and mode
	inputs := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures, "Check should succeed with no failures")

	// Verify defaults are applied
	outputMap := resource.PropertyMap{}
	for k, v := range resp.Inputs.GetFields() {
		outputMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
	}

	assert.True(t, outputMap["entityType"].HasValue(), "entityType should have a value")
	assert.Equal(t, "stacks", outputMap["entityType"].StringValue(), "entityType should default to 'stacks'")

	assert.True(t, outputMap["mode"].HasValue(), "mode should have a value")
	assert.Equal(t, "audit", outputMap["mode"].StringValue(), "mode should default to 'audit'")
}

// TestPolicyGroup_Check_ExplicitValues tests that Check preserves explicit entityType and mode values
func TestPolicyGroup_Check_ExplicitValues(t *testing.T) {
	testCases := []struct {
		name       string
		entityType string
		mode       string
	}{
		{
			name:       "stacks and audit",
			entityType: "stacks",
			mode:       "audit",
		},
		{
			name:       "stacks and preventative",
			entityType: "stacks",
			mode:       "preventative",
		},
		{
			name:       "accounts and audit",
			entityType: "accounts",
			mode:       "audit",
		},
		{
			name:       "accounts and preventative",
			entityType: "accounts",
			mode:       "preventative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := PulumiServicePolicyGroupResource{
				Client: &PolicyGroupClientMock{},
			}

			inputs := resource.PropertyMap{
				"name":             resource.NewStringProperty("test-policy-group"),
				"organizationName": resource.NewStringProperty("test-org"),
				"entityType":       resource.NewStringProperty(tc.entityType),
				"mode":             resource.NewStringProperty(tc.mode),
				"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
				"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
			}

			inputsStruct, err := structpb.NewStruct(inputs.Mappable())
			require.NoError(t, err)

			req := &pulumirpc.CheckRequest{
				Urn:  "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
				News: inputsStruct,
			}

			resp, err := provider.Check(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Empty(t, resp.Failures, "Check should succeed with no failures")

			// Verify explicit values are preserved
			outputMap := resource.PropertyMap{}
			for k, v := range resp.Inputs.GetFields() {
				outputMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
			}

			assert.Equal(t, tc.entityType, outputMap["entityType"].StringValue(), "entityType should preserve explicit value")
			assert.Equal(t, tc.mode, outputMap["mode"].StringValue(), "mode should preserve explicit value")
		})
	}
}

// TestPolicyGroup_Check_InvalidEntityType tests that Check validates entityType enum values
func TestPolicyGroup_Check_InvalidEntityType(t *testing.T) {
	testCases := []struct {
		name       string
		entityType string
	}{
		{name: "empty string", entityType: ""},
		{name: "invalid value", entityType: "invalid"},
		{name: "wrong case", entityType: "Stacks"},
		{name: "typo", entityType: "stack"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := PulumiServicePolicyGroupResource{
				Client: &PolicyGroupClientMock{},
			}

			inputs := resource.PropertyMap{
				"name":             resource.NewStringProperty("test-policy-group"),
				"organizationName": resource.NewStringProperty("test-org"),
				"entityType":       resource.NewStringProperty(tc.entityType),
				"mode":             resource.NewStringProperty("audit"),
				"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
				"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
			}

			inputsStruct, err := structpb.NewStruct(inputs.Mappable())
			require.NoError(t, err)

			req := &pulumirpc.CheckRequest{
				Urn:  "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
				News: inputsStruct,
			}

			resp, err := provider.Check(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Failures, "Check should fail with validation error")

			// Verify failure message mentions entityType
			found := false
			for _, failure := range resp.Failures {
				if failure.Property == "entityType" {
					found = true
					assert.Contains(t, failure.Reason, "entityType must be either 'stacks' or 'accounts'")
					break
				}
			}
			assert.True(t, found, "Should have failure for entityType property")
		})
	}
}

// TestPolicyGroup_Check_InvalidMode tests that Check validates mode enum values
func TestPolicyGroup_Check_InvalidMode(t *testing.T) {
	testCases := []struct {
		name string
		mode string
	}{
		{name: "empty string", mode: ""},
		{name: "invalid value", mode: "invalid"},
		{name: "wrong case", mode: "Audit"},
		{name: "typo", mode: "audits"},
		{name: "preventive (typo)", mode: "preventive"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := PulumiServicePolicyGroupResource{
				Client: &PolicyGroupClientMock{},
			}

			inputs := resource.PropertyMap{
				"name":             resource.NewStringProperty("test-policy-group"),
				"organizationName": resource.NewStringProperty("test-org"),
				"entityType":       resource.NewStringProperty("stacks"),
				"mode":             resource.NewStringProperty(tc.mode),
				"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
				"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
			}

			inputsStruct, err := structpb.NewStruct(inputs.Mappable())
			require.NoError(t, err)

			req := &pulumirpc.CheckRequest{
				Urn:  "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
				News: inputsStruct,
			}

			resp, err := provider.Check(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Failures, "Check should fail with validation error")

			// Verify failure message mentions mode
			found := false
			for _, failure := range resp.Failures {
				if failure.Property == "mode" {
					found = true
					assert.Contains(t, failure.Reason, "mode must be either 'audit' or 'preventative'")
					break
				}
			}
			assert.True(t, found, "Should have failure for mode property")
		})
	}
}

// TestPolicyGroup_Diff_EntityTypeChange tests that changing entityType triggers replacement
func TestPolicyGroup_Diff_EntityTypeChange(t *testing.T) {
	provider := PulumiServicePolicyGroupResource{
		Client: &PolicyGroupClientMock{},
	}

	// Old state: entityType = "stacks"
	oldInputs := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"entityType":       resource.NewStringProperty("stacks"),
		"mode":             resource.NewStringProperty("audit"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	// New state: entityType = "accounts"
	newInputs := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"entityType":       resource.NewStringProperty("accounts"),
		"mode":             resource.NewStringProperty("audit"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/test-policy-group",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify that entityType change triggers replacement
	assert.Contains(t, resp.Replaces, "entityType", "Changing entityType should trigger replacement")
}

// TestPolicyGroup_Diff_ModeChange tests that changing mode triggers replacement
func TestPolicyGroup_Diff_ModeChange(t *testing.T) {
	provider := PulumiServicePolicyGroupResource{
		Client: &PolicyGroupClientMock{},
	}

	// Old state: mode = "audit"
	oldInputs := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"entityType":       resource.NewStringProperty("stacks"),
		"mode":             resource.NewStringProperty("audit"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	// New state: mode = "preventative"
	newInputs := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"entityType":       resource.NewStringProperty("stacks"),
		"mode":             resource.NewStringProperty("preventative"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/test-policy-group",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify that mode change triggers replacement
	assert.Contains(t, resp.Replaces, "mode", "Changing mode should trigger replacement")
}

// TestPolicyGroup_Diff_ArrayOrderIndependent tests that arrays with same elements in different order don't cause diffs
func TestPolicyGroup_Diff_ArrayOrderIndependent(t *testing.T) {
	provider := PulumiServicePolicyGroupResource{
		Client: &PolicyGroupClientMock{},
	}

	// Old state (from refresh): accounts in one order
	oldInputs := resource.PropertyMap{
		"accounts": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("account-b"),
			resource.NewStringProperty("account-a"),
		}),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	// New inputs (from program): accounts in different order
	newInputs := resource.PropertyMap{
		"accounts": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("account-a"),
			resource.NewStringProperty("account-b"),
		}),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:        "test-org/test-policy-group",
		Urn:       "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
		OldInputs: oldState,
		News:      newState,
	}

	resp, err := provider.Diff(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify that there is NO diff (order of array elements should not matter)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes,
		"Expected no diff when array elements are in different order but contain same values")
	assert.Empty(t, resp.Replaces, "Expected no replacements")
}

// TestPolicyGroup_Diff_NullVsEmptyArray tests that null arrays in inputs don't cause diffs against empty arrays in state
func TestPolicyGroup_Diff_NullVsEmptyArray(t *testing.T) {
	provider := PulumiServicePolicyGroupResource{
		Client: &PolicyGroupClientMock{},
	}

	// Old state (from refresh): empty arrays
	oldInputs := resource.PropertyMap{
		"stacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
		"accounts":    resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks": resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	// New inputs (from program): null values (property not specified)
	newInputs := resource.PropertyMap{
		"stacks":      resource.NewNullProperty(),
		"accounts":    resource.NewNullProperty(),
		"policyPacks": resource.NewNullProperty(),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:        "test-org/test-policy-group",
		Urn:       "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
		OldInputs: oldState,
		News:      newState,
	}

	resp, err := provider.Diff(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify that there is NO diff (null should be treated same as empty array)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes,
		"Expected no diff when comparing null arrays with empty arrays")
	assert.Empty(t, resp.Replaces, "Expected no replacements")
}

// TestPolicyGroup_ToPropertyMap tests that ToPropertyMap includes entityType and mode
func TestPolicyGroup_ToPropertyMap(t *testing.T) {
	input := PulumiServicePolicyGroupInput{
		Name:             "test-policy-group",
		OrganizationName: "test-org",
		EntityType:       "accounts",
		Mode:             "preventative",
		Stacks:           []pulumiapi.StackReference{},
		PolicyPacks:      []pulumiapi.PolicyPackMetadata{},
	}

	pm := input.ToPropertyMap()

	// Verify all fields are present
	assert.True(t, pm["name"].HasValue())
	assert.Equal(t, "test-policy-group", pm["name"].StringValue())

	assert.True(t, pm["organizationName"].HasValue())
	assert.Equal(t, "test-org", pm["organizationName"].StringValue())

	assert.True(t, pm["entityType"].HasValue())
	assert.Equal(t, "accounts", pm["entityType"].StringValue())

	assert.True(t, pm["mode"].HasValue())
	assert.Equal(t, "preventative", pm["mode"].StringValue())

	assert.True(t, pm["stacks"].HasValue())
	assert.True(t, pm["stacks"].IsArray())

	assert.True(t, pm["policyPacks"].HasValue())
	assert.True(t, pm["policyPacks"].IsArray())
}

// TestPolicyGroup_ToPulumiServicePolicyGroupInput tests parsing of entityType and mode from PropertyMap
func TestPolicyGroup_ToPulumiServicePolicyGroupInput(t *testing.T) {
	inputMap := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"entityType":       resource.NewStringProperty("accounts"),
		"mode":             resource.NewStringProperty("preventative"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	result := ToPulumiServicePolicyGroupInput(inputMap)

	assert.Equal(t, "test-policy-group", result.Name)
	assert.Equal(t, "test-org", result.OrganizationName)
	assert.Equal(t, "accounts", result.EntityType)
	assert.Equal(t, "preventative", result.Mode)
	assert.Empty(t, result.Stacks)
	assert.Empty(t, result.PolicyPacks)
}

// TestPolicyGroup_ToPulumiServicePolicyGroupInput_MissingFields tests that missing entityType/mode are handled
func TestPolicyGroup_ToPulumiServicePolicyGroupInput_MissingFields(t *testing.T) {
	inputMap := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks":      resource.NewArrayProperty([]resource.PropertyValue{}),
	}

	result := ToPulumiServicePolicyGroupInput(inputMap)

	assert.Equal(t, "test-policy-group", result.Name)
	assert.Equal(t, "test-org", result.OrganizationName)
	// EntityType and Mode should be empty strings if not provided
	assert.Equal(t, "", result.EntityType)
	assert.Equal(t, "", result.Mode)
}

// TestPolicyGroup_ToPulumiServicePolicyGroupInput_OptionalPolicyPackFields tests that optional policy pack fields are handled
func TestPolicyGroup_ToPulumiServicePolicyGroupInput_OptionalPolicyPackFields(t *testing.T) {
	// Test with only required name field
	inputMap := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"entityType":       resource.NewStringProperty("stacks"),
		"mode":             resource.NewStringProperty("audit"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"name": resource.NewStringProperty("test-policy-pack"),
			}),
		}),
	}

	result := ToPulumiServicePolicyGroupInput(inputMap)

	assert.Equal(t, "test-policy-group", result.Name)
	assert.Equal(t, "test-org", result.OrganizationName)
	assert.Equal(t, "stacks", result.EntityType)
	assert.Equal(t, "audit", result.Mode)
	assert.Empty(t, result.Stacks)
	assert.Len(t, result.PolicyPacks, 1)

	// Verify only name is set, others are empty/zero values
	pp := result.PolicyPacks[0]
	assert.Equal(t, "test-policy-pack", pp.Name)
	assert.Equal(t, "", pp.DisplayName)
	assert.Equal(t, 0, pp.Version)
	assert.Equal(t, "", pp.VersionTag)
	assert.Nil(t, pp.Config)
}

// TestPolicyGroup_ToPulumiServicePolicyGroupInput_PartialPolicyPackFields tests mixed optional fields
func TestPolicyGroup_ToPulumiServicePolicyGroupInput_PartialPolicyPackFields(t *testing.T) {
	// Test with some optional fields provided
	inputMap := resource.PropertyMap{
		"name":             resource.NewStringProperty("test-policy-group"),
		"organizationName": resource.NewStringProperty("test-org"),
		"entityType":       resource.NewStringProperty("stacks"),
		"mode":             resource.NewStringProperty("audit"),
		"stacks":           resource.NewArrayProperty([]resource.PropertyValue{}),
		"policyPacks": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"name":        resource.NewStringProperty("test-policy-pack"),
				"displayName": resource.NewStringProperty("Test Policy Pack"),
				// Note: version and versionTag are intentionally omitted
				"config": resource.NewObjectProperty(resource.PropertyMap{
					"enabled": resource.NewBoolProperty(true),
				}),
			}),
		}),
	}

	result := ToPulumiServicePolicyGroupInput(inputMap)

	assert.Len(t, result.PolicyPacks, 1)

	// Verify partial fields are set correctly
	pp := result.PolicyPacks[0]
	assert.Equal(t, "test-policy-pack", pp.Name)
	assert.Equal(t, "Test Policy Pack", pp.DisplayName)
	assert.Equal(t, 0, pp.Version)     // Should be zero value when not provided
	assert.Equal(t, "", pp.VersionTag) // Should be empty string when not provided
	assert.NotNil(t, pp.Config)
	assert.Equal(t, true, pp.Config["enabled"])
}

// TestHasParentAccount tests the hasParentAccount helper function
func TestHasParentAccount(t *testing.T) {
	tests := []struct {
		name     string
		account  string
		accounts []string
		expected bool
	}{
		{
			name:     "child account with parent in list",
			account:  "parent/child",
			accounts: []string{"parent", "other"},
			expected: true,
		},
		{
			name:     "child account without parent in list",
			account:  "parent/child",
			accounts: []string{"other", "another"},
			expected: false,
		},
		{
			name:     "parent account (no parent exists)",
			account:  "parent",
			accounts: []string{"other", "another"},
			expected: false,
		},
		{
			name:     "empty accounts list",
			account:  "parent/child",
			accounts: []string{},
			expected: false,
		},
		{
			name:     "similar prefix but not parent",
			account:  "parent-other/child",
			accounts: []string{"parent", "other"},
			expected: false,
		},
		{
			name:     "exact match is not a parent",
			account:  "parent",
			accounts: []string{"parent"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasParentAccount(tt.account, tt.accounts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPolicyGroup_Read(t *testing.T) {
	t.Run("preserves original inputs when API state matches previous state", func(t *testing.T) {
		// Scenario: User specified ["parent-account"], API returned ["parent-account", "parent-account/child"]
		// After refresh, API still returns the same. We should preserve the original input ["parent-account"]
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return &pulumiapi.PolicyGroup{
						Name:       "test-policy-group",
						EntityType: "accounts",
						Mode:       "audit",
						Accounts:   []string{"parent-account", "parent-account/child"},
					}, nil
				},
			},
		}

		// Previous state (from last refresh) - includes auto-added child
		previousState := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("parent-account"),
				resource.NewStringProperty("parent-account/child"),
			}),
		}
		propsStruct, err := structpb.NewStruct(previousState.Mappable())
		require.NoError(t, err)

		// Previous inputs (what user originally specified)
		previousInputs := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("parent-account"),
			}),
		}
		inputsStruct, err := structpb.NewStruct(previousInputs.Mappable())
		require.NoError(t, err)

		req := &pulumirpc.ReadRequest{
			Id:         "test-org/test-policy-group",
			Urn:        "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
			Properties: propsStruct,
			Inputs:     inputsStruct,
		}

		resp, err := provider.Read(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Properties should reflect full API response (including child)
		propsMap := resource.PropertyMap{}
		for k, v := range resp.Properties.GetFields() {
			propsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
		}
		assert.Len(t, propsMap["accounts"].ArrayValue(), 2, "Properties should have both parent and child accounts")

		// Inputs should preserve original user input (just the parent)
		inputsMap := resource.PropertyMap{}
		for k, v := range resp.Inputs.GetFields() {
			inputsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
		}
		assert.Len(t, inputsMap["accounts"].ArrayValue(), 1, "Inputs should preserve original user input")
		assert.Equal(t, "parent-account", inputsMap["accounts"].ArrayValue()[0].StringValue())
	})

	t.Run("updates inputs when API state has changed externally", func(t *testing.T) {
		// Scenario: API state changed (someone added a new account externally)
		// We should update inputs to reflect the new state
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return &pulumiapi.PolicyGroup{
						Name:       "test-policy-group",
						EntityType: "accounts",
						Mode:       "audit",
						Accounts:   []string{"account-a", "account-b", "account-c"}, // External change
					}, nil
				},
			},
		}

		// Previous state (before external change)
		previousState := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("account-a"),
				resource.NewStringProperty("account-b"),
			}),
		}
		propsStruct, err := structpb.NewStruct(previousState.Mappable())
		require.NoError(t, err)

		// Previous inputs
		previousInputs := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("account-a"),
				resource.NewStringProperty("account-b"),
			}),
		}
		inputsStruct, err := structpb.NewStruct(previousInputs.Mappable())
		require.NoError(t, err)

		req := &pulumirpc.ReadRequest{
			Id:         "test-org/test-policy-group",
			Urn:        "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
			Properties: propsStruct,
			Inputs:     inputsStruct,
		}

		resp, err := provider.Read(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Both properties and inputs should reflect the new API state
		inputsMap := resource.PropertyMap{}
		for k, v := range resp.Inputs.GetFields() {
			inputsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
		}
		assert.Len(t, inputsMap["accounts"].ArrayValue(), 3, "Inputs should be updated to match new API state")
	})

	t.Run("handles read with no previous state or inputs", func(t *testing.T) {
		// Scenario: Import or first read - no previous state available
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return &pulumiapi.PolicyGroup{
						Name:       "test-policy-group",
						EntityType: "accounts",
						Mode:       "audit",
						Accounts:   []string{"account-a"},
					}, nil
				},
			},
		}

		req := &pulumirpc.ReadRequest{
			Id:         "test-org/test-policy-group",
			Urn:        "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
			Properties: nil,
			Inputs:     nil,
		}

		resp, err := provider.Read(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Both should reflect API state
		propsMap := resource.PropertyMap{}
		for k, v := range resp.Properties.GetFields() {
			propsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
		}
		assert.Len(t, propsMap["accounts"].ArrayValue(), 1)

		inputsMap := resource.PropertyMap{}
		for k, v := range resp.Inputs.GetFields() {
			inputsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
		}
		assert.Len(t, inputsMap["accounts"].ArrayValue(), 1)
	})

	t.Run("returns empty response when policy group not found", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return nil, nil // Not found
				},
			},
		}

		req := &pulumirpc.ReadRequest{
			Id:  "test-org/test-policy-group",
			Urn: "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup",
		}

		resp, err := provider.Read(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Id, "Should return empty response for not found")
		assert.Nil(t, resp.Properties)
	})
}

// TestParsePreviousAccounts tests the parsePreviousAccounts helper function
func TestParsePreviousAccounts(t *testing.T) {
	t.Run("parses both properties and inputs", func(t *testing.T) {
		props := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("state-account-1"),
				resource.NewStringProperty("state-account-2"),
			}),
		}
		propsStruct, err := structpb.NewStruct(props.Mappable())
		require.NoError(t, err)

		inputs := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("input-account-1"),
			}),
		}
		inputsStruct, err := structpb.NewStruct(inputs.Mappable())
		require.NoError(t, err)

		stateAccounts, inputAccounts := parsePreviousAccounts(propsStruct, inputsStruct)

		assert.Equal(t, []string{"state-account-1", "state-account-2"}, stateAccounts)
		assert.Equal(t, []string{"input-account-1"}, inputAccounts)
	})

	t.Run("handles nil properties", func(t *testing.T) {
		inputs := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("input-account"),
			}),
		}
		inputsStruct, err := structpb.NewStruct(inputs.Mappable())
		require.NoError(t, err)

		stateAccounts, inputAccounts := parsePreviousAccounts(nil, inputsStruct)

		assert.Nil(t, stateAccounts)
		assert.Equal(t, []string{"input-account"}, inputAccounts)
	})

	t.Run("handles nil inputs", func(t *testing.T) {
		props := resource.PropertyMap{
			"accounts": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("state-account"),
			}),
		}
		propsStruct, err := structpb.NewStruct(props.Mappable())
		require.NoError(t, err)

		stateAccounts, inputAccounts := parsePreviousAccounts(propsStruct, nil)

		assert.Equal(t, []string{"state-account"}, stateAccounts)
		assert.Nil(t, inputAccounts)
	})

	t.Run("handles both nil", func(t *testing.T) {
		stateAccounts, inputAccounts := parsePreviousAccounts(nil, nil)

		assert.Nil(t, stateAccounts)
		assert.Nil(t, inputAccounts)
	})
}

// Test the refactored PolicyGroup serialization with complex policy pack configs
func TestPolicyGroupSerializationConsistency(t *testing.T) {
	// Create a policy group input with complex config
	input := PulumiServicePolicyGroupInput{
		Name:             "test-group",
		OrganizationName: "test-org",
		EntityType:       "stacks",
		Mode:             "audit",
		Stacks: []pulumiapi.StackReference{
			{Name: "stack1", RoutingProject: "project1"},
			{Name: "stack2", RoutingProject: "project2"},
		},
		PolicyPacks: []pulumiapi.PolicyPackMetadata{
			{
				Name:        "aws-compliance",
				DisplayName: "AWS Compliance Pack",
				Version:     1,
				VersionTag:  "v1.0.0",
				Config: map[string]interface{}{
					"approvedAmiIds": []interface{}{"ami-123", "ami-456"},
					"maxInstances":   float64(10),
					"nestedConfig": map[string]interface{}{
						"regions": []interface{}{"us-east-1", "us-west-2"},
						"enabled": true,
					},
				},
			},
		},
	}

	// Convert to PropertyMap using refactored method
	propertyMap := input.ToPropertyMap()

	// Convert back using refactored method
	roundtripInput := ToPulumiServicePolicyGroupInput(propertyMap)

	// Verify all fields are preserved correctly
	assert.Equal(t, input.Name, roundtripInput.Name)
	assert.Equal(t, input.OrganizationName, roundtripInput.OrganizationName)
	assert.Equal(t, input.EntityType, roundtripInput.EntityType)
	assert.Equal(t, input.Mode, roundtripInput.Mode)
	assert.Equal(t, input.Stacks, roundtripInput.Stacks)

	// Verify policy pack details including complex config
	assert.Len(t, roundtripInput.PolicyPacks, 1)
	pp := roundtripInput.PolicyPacks[0]
	assert.Equal(t, input.PolicyPacks[0].Name, pp.Name)
	assert.Equal(t, input.PolicyPacks[0].DisplayName, pp.DisplayName)
	assert.Equal(t, input.PolicyPacks[0].Version, pp.Version)
	assert.Equal(t, input.PolicyPacks[0].VersionTag, pp.VersionTag)

	// Verify complex config is preserved
	originalConfig := input.PolicyPacks[0].Config
	roundtripConfig := pp.Config
	assert.Equal(t, originalConfig, roundtripConfig)
}
