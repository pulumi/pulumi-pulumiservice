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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// Mock client for PolicyGroup tests
type PolicyGroupClientMock struct {
	getPolicyGroupFunc         func() (*pulumiapi.PolicyGroup, error)
	createPolicyGroupFunc      func(ctx context.Context, orgName, policyGroupName, entityType, mode string) error
	batchUpdatePolicyGroupFunc func(
		ctx context.Context, orgName, policyGroupName string, reqs []pulumiapi.UpdatePolicyGroupRequest,
	) error
}

func (c *PolicyGroupClientMock) ListPolicyGroups(
	_ context.Context,
	_ string,
) ([]pulumiapi.PolicyGroupSummary, error) {
	return nil, nil
}

func (c *PolicyGroupClientMock) GetPolicyGroup(
	_ context.Context,
	_ string,
	_ string,
) (*pulumiapi.PolicyGroup, error) {
	if c.getPolicyGroupFunc != nil {
		return c.getPolicyGroupFunc()
	}
	return nil, nil
}

func (c *PolicyGroupClientMock) CreatePolicyGroup(
	ctx context.Context,
	orgName, policyGroupName, entityType, mode string,
) error {
	if c.createPolicyGroupFunc != nil {
		return c.createPolicyGroupFunc(ctx, orgName, policyGroupName, entityType, mode)
	}
	return nil
}

func (c *PolicyGroupClientMock) UpdatePolicyGroup(
	_ context.Context,
	_, _ string,
	_ pulumiapi.UpdatePolicyGroupRequest,
) error {
	return nil
}

func (c *PolicyGroupClientMock) BatchUpdatePolicyGroup(
	ctx context.Context,
	orgName, policyGroupName string,
	reqs []pulumiapi.UpdatePolicyGroupRequest,
) error {
	if c.batchUpdatePolicyGroupFunc != nil {
		return c.batchUpdatePolicyGroupFunc(ctx, orgName, policyGroupName, reqs)
	}
	return nil
}

func (c *PolicyGroupClientMock) DeletePolicyGroup(_ context.Context, _, _ string) error {
	return nil
}

// Test helper types and functions

// policyGroupInputBuilder provides a fluent API for building policy group inputs
type policyGroupInputBuilder struct {
	name        string
	org         string
	entityType  string
	mode        string
	stacks      []stackRef
	policyPacks []policyPackRef
	accounts    []string
}

type stackRef struct {
	name, project string
}

type policyPackRef struct {
	name    string
	version int
}

func newPolicyGroupInput() *policyGroupInputBuilder {
	return &policyGroupInputBuilder{
		name:       "test-policy-group",
		org:        "test-org",
		entityType: "stacks",
		mode:       "audit",
	}
}

func (b *policyGroupInputBuilder) withEntityType(t string) *policyGroupInputBuilder {
	b.entityType = t
	return b
}

func (b *policyGroupInputBuilder) withStacks(stacks ...stackRef) *policyGroupInputBuilder {
	b.stacks = stacks
	return b
}

func (b *policyGroupInputBuilder) withPolicyPacks(packs ...policyPackRef) *policyGroupInputBuilder {
	b.policyPacks = packs
	return b
}

func (b *policyGroupInputBuilder) withAccounts(accounts ...string) *policyGroupInputBuilder {
	b.accounts = accounts
	return b
}

func (b *policyGroupInputBuilder) build() resource.PropertyMap {
	pm := resource.PropertyMap{
		"name":             resource.NewStringProperty(b.name),
		"organizationName": resource.NewStringProperty(b.org),
		"entityType":       resource.NewStringProperty(b.entityType),
		"mode":             resource.NewStringProperty(b.mode),
	}

	// Build stacks array
	stackValues := make([]resource.PropertyValue, len(b.stacks))
	for i, s := range b.stacks {
		stackValues[i] = resource.NewObjectProperty(resource.PropertyMap{
			"name":           resource.NewStringProperty(s.name),
			"routingProject": resource.NewStringProperty(s.project),
		})
	}
	pm["stacks"] = resource.NewArrayProperty(stackValues)

	// Build policy packs array
	ppValues := make([]resource.PropertyValue, len(b.policyPacks))
	for i, pp := range b.policyPacks {
		ppValues[i] = resource.NewObjectProperty(resource.PropertyMap{
			"name":    resource.NewStringProperty(pp.name),
			"version": resource.NewNumberProperty(float64(pp.version)),
		})
	}
	pm["policyPacks"] = resource.NewArrayProperty(ppValues)

	// Build accounts array
	accountValues := make([]resource.PropertyValue, len(b.accounts))
	for i, a := range b.accounts {
		accountValues[i] = resource.NewStringProperty(a)
	}
	pm["accounts"] = resource.NewArrayProperty(accountValues)

	return pm
}

func (b *policyGroupInputBuilder) buildStruct(t *testing.T) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(b.build().Mappable())
	require.NoError(t, err)
	return s
}

// mockPolicyGroupBuilder provides a fluent API for building mock PolicyGroup responses
type mockPolicyGroupBuilder struct {
	pg *pulumiapi.PolicyGroup
}

func newMockPolicyGroup() *mockPolicyGroupBuilder {
	return &mockPolicyGroupBuilder{
		pg: &pulumiapi.PolicyGroup{
			Name:       "test-policy-group",
			EntityType: "stacks",
			Mode:       "audit",
		},
	}
}

func (b *mockPolicyGroupBuilder) withStacks(stacks ...stackRef) *mockPolicyGroupBuilder {
	b.pg.Stacks = make([]pulumiapi.StackReference, len(stacks))
	for i, s := range stacks {
		b.pg.Stacks[i] = pulumiapi.StackReference{Name: s.name, RoutingProject: s.project}
	}
	return b
}

func (b *mockPolicyGroupBuilder) withPolicyPacks(packs ...policyPackRef) *mockPolicyGroupBuilder {
	b.pg.AppliedPolicyPacks = make([]pulumiapi.PolicyPackMetadata, len(packs))
	for i, p := range packs {
		b.pg.AppliedPolicyPacks[i] = pulumiapi.PolicyPackMetadata{Name: p.name, Version: p.version}
	}
	return b
}

func (b *mockPolicyGroupBuilder) withAccounts(accounts ...string) *mockPolicyGroupBuilder {
	b.pg.EntityType = "accounts"
	b.pg.Accounts = accounts
	return b
}

func (b *mockPolicyGroupBuilder) build() *pulumiapi.PolicyGroup {
	return b.pg
}

const testPolicyGroupURN = "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::testPolicyGroup"
const testPolicyGroupID = "test-org/test-policy-group"

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

			assert.Equal(
				t,
				tc.entityType,
				outputMap["entityType"].StringValue(),
				"entityType should preserve explicit value",
			)
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

// TestPolicyGroup_Diff_NullVsEmptyArray tests that null arrays in inputs don't
// cause diffs against empty arrays in state
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

// TestPolicyGroup_ToPulumiServicePolicyGroupInput_OptionalPolicyPackFields
// tests that optional policy pack fields are handled
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
	parentAccount := "parent-account"
	childAccount := "parent-account/child"
	accountA := "account-a"
	accountB := "account-b"
	accountC := "account-c"

	t.Run("preserves original inputs when API state matches previous state", func(t *testing.T) {
		// Scenario: User specified ["parent-account"], API returned ["parent-account", "parent-account/child"]
		// After refresh, API still returns the same. We should preserve the original input ["parent-account"]
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withAccounts(parentAccount, childAccount).build(), nil
				},
			},
		}

		req := &pulumirpc.ReadRequest{
			Id:  testPolicyGroupID,
			Urn: testPolicyGroupURN,
			Properties: newPolicyGroupInput().withEntityType("accounts").
				withAccounts(parentAccount, childAccount).
				buildStruct(t),
			Inputs: newPolicyGroupInput().withEntityType("accounts").withAccounts(parentAccount).buildStruct(t),
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
		assert.Equal(t, parentAccount, inputsMap["accounts"].ArrayValue()[0].StringValue())
	})

	t.Run("updates inputs when API state has changed externally", func(t *testing.T) {
		// Scenario: API state changed (someone added a new account externally)
		// We should update inputs to reflect the new state
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withAccounts(accountA, accountB, accountC).build(), nil
				},
			},
		}

		req := &pulumirpc.ReadRequest{
			Id:  testPolicyGroupID,
			Urn: testPolicyGroupURN,
			Properties: newPolicyGroupInput().withEntityType("accounts").
				withAccounts(accountA, accountB).
				buildStruct(t),
			Inputs: newPolicyGroupInput().withEntityType("accounts").
				withAccounts(accountA, accountB).
				buildStruct(t),
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
					return newMockPolicyGroup().withAccounts(accountA).build(), nil
				},
			},
		}

		req := &pulumirpc.ReadRequest{
			Id:         testPolicyGroupID,
			Urn:        testPolicyGroupURN,
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
			Id:  testPolicyGroupID,
			Urn: testPolicyGroupURN,
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

// TestPolicyGroup_Create tests the Create function
func TestPolicyGroup_Create(t *testing.T) {
	stack1 := stackRef{name: "stack-1", project: "project-1"}
	stack2 := stackRef{name: "stack-2", project: "project-2"}
	pp1 := policyPackRef{name: "policy-pack-1", version: 1}
	account1 := "account-1"
	account2 := "account-2"

	t.Run("creates empty policy group", func(t *testing.T) {
		var createCalled bool
		var batchUpdateCalled bool
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, orgName, policyGroupName, entityType, mode string) error {
					createCalled = true
					assert.Equal(t, "test-org", orgName)
					assert.Equal(t, "test-policy-group", policyGroupName)
					assert.Equal(t, "stacks", entityType)
					assert.Equal(t, "audit", mode)
					return nil
				},
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					batchUpdateCalled = true
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().build(), nil
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn:        testPolicyGroupURN,
			Properties: newPolicyGroupInput().buildStruct(t),
		}

		resp, err := provider.Create(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.True(t, createCalled, "CreatePolicyGroup should be called")
		assert.False(t, batchUpdateCalled, "BatchUpdatePolicyGroup should not be called for empty policy group")
		assert.Equal(t, testPolicyGroupID, resp.Id)
	})

	t.Run("creates policy group with stacks", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, _, _, _, _ string) error {
					return nil
				},
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withStacks(stack1, stack2).build(), nil
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn:        testPolicyGroupURN,
			Properties: newPolicyGroupInput().withStacks(stack1, stack2).buildStruct(t),
		}

		resp, err := provider.Create(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 2)
		require.NotNil(t, capturedReqs[0].AddStack)
		assert.Equal(t, stack1.name, capturedReqs[0].AddStack.Name)
		require.NotNil(t, capturedReqs[1].AddStack)
		assert.Equal(t, stack2.name, capturedReqs[1].AddStack.Name)
	})

	t.Run("creates policy group with policy packs", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, _, _, _, _ string) error {
					return nil
				},
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withPolicyPacks(pp1).build(), nil
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn:        testPolicyGroupURN,
			Properties: newPolicyGroupInput().withPolicyPacks(pp1).buildStruct(t),
		}

		resp, err := provider.Create(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 1)
		require.NotNil(t, capturedReqs[0].AddPolicyPack)
		assert.Equal(t, pp1.name, capturedReqs[0].AddPolicyPack.Name)
	})

	t.Run("creates policy group with accounts", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, _, _, _, _ string) error {
					return nil
				},
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withAccounts(account1, account2).build(), nil
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn: testPolicyGroupURN,
			Properties: newPolicyGroupInput().withEntityType("accounts").
				withAccounts(account1, account2).
				buildStruct(t),
		}

		resp, err := provider.Create(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 2)
		require.NotNil(t, capturedReqs[0].AddInsightsAccount)
		assert.Equal(t, account1, capturedReqs[0].AddInsightsAccount.Name)
		require.NotNil(t, capturedReqs[1].AddInsightsAccount)
		assert.Equal(t, account2, capturedReqs[1].AddInsightsAccount.Name)
	})

	t.Run("returns error from CreatePolicyGroup", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, _, _, _, _ string) error {
					return fmt.Errorf("create error")
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn:        testPolicyGroupURN,
			Properties: newPolicyGroupInput().buildStruct(t),
		}

		_, err := provider.Create(req)
		assert.ErrorContains(t, err, "error creating policy group")
		assert.ErrorContains(t, err, "create error")
	})

	t.Run("returns error from BatchUpdatePolicyGroup", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, _, _, _, _ string) error {
					return nil
				},
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					return fmt.Errorf("batch update error")
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn:        testPolicyGroupURN,
			Properties: newPolicyGroupInput().withStacks(stack1).buildStruct(t),
		}

		_, err := provider.Create(req)
		assert.ErrorContains(t, err, "failed to add items to policy group")
		assert.ErrorContains(t, err, "batch update error")
	})

	t.Run("returns error from GetPolicyGroup after create", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, _, _, _, _ string) error {
					return nil
				},
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return nil, fmt.Errorf("read error")
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn:        testPolicyGroupURN,
			Properties: newPolicyGroupInput().withStacks(stack1).buildStruct(t),
		}

		_, err := provider.Create(req)
		assert.ErrorContains(t, err, "read error")
	})

	t.Run("returns state from API response", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				createPolicyGroupFunc: func(_ context.Context, _, _, _, _ string) error {
					return nil
				},
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					// API returns more than what was requested (e.g., auto-added child accounts)
					return newMockPolicyGroup().withAccounts(account1, account1+"/child").build(), nil
				},
			},
		}

		req := &pulumirpc.CreateRequest{
			Urn:        testPolicyGroupURN,
			Properties: newPolicyGroupInput().withEntityType("accounts").withAccounts(account1).buildStruct(t),
		}

		resp, err := provider.Create(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Properties)

		propsMap := resource.PropertyMap{}
		for k, v := range resp.Properties.GetFields() {
			propsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
		}

		accounts := propsMap["accounts"].ArrayValue()
		require.Len(t, accounts, 2, "Should return full state from API including auto-added accounts")
	})
}

// TestPolicyGroup_Update tests the Update function
func TestPolicyGroup_Update(t *testing.T) {
	stack1 := stackRef{name: "stack-1", project: "project-1"}
	stack2 := stackRef{name: "stack-2", project: "project-2"}
	pp1 := policyPackRef{name: "policy-pack-1", version: 1}
	pp2 := policyPackRef{name: "policy-pack-2", version: 1}
	account1 := "account-1"
	account2 := "account-2"
	parentAccount := "parent-account"
	childAccount := "parent-account/child"

	t.Run("adds new stacks", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withStacks(stack1, stack2).build(), nil
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: newPolicyGroupInput().withStacks(stack1).buildStruct(t),
			News: newPolicyGroupInput().withStacks(stack1, stack2).buildStruct(t),
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 1)
		require.NotNil(t, capturedReqs[0].AddStack)
		assert.Equal(t, stack2.name, capturedReqs[0].AddStack.Name)
		assert.Equal(t, stack2.project, capturedReqs[0].AddStack.RoutingProject)
	})

	t.Run("removes stacks", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withStacks(stack1).build(), nil
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: newPolicyGroupInput().withStacks(stack1, stack2).buildStruct(t),
			News: newPolicyGroupInput().withStacks(stack1).buildStruct(t),
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 1)
		require.NotNil(t, capturedReqs[0].RemoveStack)
		assert.Equal(t, stack2.name, capturedReqs[0].RemoveStack.Name)
		assert.Equal(t, stack2.project, capturedReqs[0].RemoveStack.RoutingProject)
	})

	t.Run("adds and removes policy packs", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withPolicyPacks(pp2).build(), nil
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: newPolicyGroupInput().withPolicyPacks(pp1).buildStruct(t),
			News: newPolicyGroupInput().withPolicyPacks(pp2).buildStruct(t),
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 2)

		var removeReq, addReq *pulumiapi.UpdatePolicyGroupRequest
		for i := range capturedReqs {
			if capturedReqs[i].RemovePolicyPack != nil {
				removeReq = &capturedReqs[i]
			}
			if capturedReqs[i].AddPolicyPack != nil {
				addReq = &capturedReqs[i]
			}
		}

		require.NotNil(t, removeReq)
		assert.Equal(t, pp1.name, removeReq.RemovePolicyPack.Name)
		require.NotNil(t, addReq)
		assert.Equal(t, pp2.name, addReq.AddPolicyPack.Name)
	})

	t.Run("adds accounts", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withAccounts(account1, account2).build(), nil
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: newPolicyGroupInput().withEntityType("accounts").withAccounts(account1).buildStruct(t),
			News: newPolicyGroupInput().withEntityType("accounts").withAccounts(account1, account2).buildStruct(t),
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 1)
		require.NotNil(t, capturedReqs[0].AddInsightsAccount)
		assert.Equal(t, account2, capturedReqs[0].AddInsightsAccount.Name)
	})

	t.Run("removes accounts but skips child accounts with parent still present", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withAccounts(parentAccount, childAccount).build(), nil
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:  testPolicyGroupID,
			Urn: testPolicyGroupURN,
			Olds: newPolicyGroupInput().withEntityType("accounts").
				withAccounts(parentAccount, childAccount).
				buildStruct(t),
			News: newPolicyGroupInput().withEntityType("accounts").withAccounts(parentAccount).buildStruct(t),
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Child should NOT be removed because parent is still present
		assert.Empty(t, capturedReqs, "Should not remove child account when parent is still present")
	})

	t.Run("removes child account when parent is also removed", func(t *testing.T) {
		var capturedReqs []pulumiapi.UpdatePolicyGroupRequest
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, reqs []pulumiapi.UpdatePolicyGroupRequest) error {
					capturedReqs = reqs
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withAccounts().build(), nil
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:  testPolicyGroupID,
			Urn: testPolicyGroupURN,
			Olds: newPolicyGroupInput().withEntityType("accounts").
				withAccounts(parentAccount, childAccount).
				buildStruct(t),
			News: newPolicyGroupInput().withEntityType("accounts").buildStruct(t),
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Len(t, capturedReqs, 2)

		removedAccounts := make([]string, 0, 2)
		for _, req := range capturedReqs {
			require.NotNil(t, req.RemoveInsightsAccount)
			removedAccounts = append(removedAccounts, req.RemoveInsightsAccount.Name)
		}
		assert.Contains(t, removedAccounts, parentAccount)
		assert.Contains(t, removedAccounts, childAccount)
	})

	t.Run("no changes when inputs are identical", func(t *testing.T) {
		var batchUpdateCalled bool
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					batchUpdateCalled = true
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return newMockPolicyGroup().withStacks(stack1).build(), nil
				},
			},
		}

		inputs := newPolicyGroupInput().withStacks(stack1).buildStruct(t)
		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: inputs,
			News: inputs,
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.False(t, batchUpdateCalled, "BatchUpdatePolicyGroup should not be called when there are no changes")
	})

	t.Run("returns error from BatchUpdatePolicyGroup", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					return fmt.Errorf("API error")
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: newPolicyGroupInput().buildStruct(t),
			News: newPolicyGroupInput().withStacks(stack1).buildStruct(t),
		}

		_, err := provider.Update(req)
		assert.ErrorContains(t, err, "failed to update policy group: API error")
	})

	t.Run("returns error from GetPolicyGroup after update", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					return nil, fmt.Errorf("read error")
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: newPolicyGroupInput().buildStruct(t),
			News: newPolicyGroupInput().withStacks(stack1).buildStruct(t),
		}

		_, err := provider.Update(req)
		assert.ErrorContains(t, err, "failed to read policy group after update: read error")
	})

	t.Run("returns updated state from API response", func(t *testing.T) {
		provider := PulumiServicePolicyGroupResource{
			Client: &PolicyGroupClientMock{
				batchUpdatePolicyGroupFunc: func(_ context.Context, _, _ string, _ []pulumiapi.UpdatePolicyGroupRequest) error {
					return nil
				},
				getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
					// API includes auto-added child account
					return newMockPolicyGroup().withAccounts(parentAccount, childAccount).build(), nil
				},
			},
		}

		req := &pulumirpc.UpdateRequest{
			Id:   testPolicyGroupID,
			Urn:  testPolicyGroupURN,
			Olds: newPolicyGroupInput().withEntityType("accounts").buildStruct(t),
			News: newPolicyGroupInput().withEntityType("accounts").withAccounts(parentAccount).buildStruct(t),
		}

		resp, err := provider.Update(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Properties)

		propsMap := resource.PropertyMap{}
		for k, v := range resp.Properties.GetFields() {
			propsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.AsInterface())
		}

		accounts := propsMap["accounts"].ArrayValue()
		require.Len(t, accounts, 2)

		accountNames := make([]string, len(accounts))
		for i, a := range accounts {
			accountNames[i] = a.StringValue()
		}
		assert.Contains(t, accountNames, parentAccount)
		assert.Contains(t, accountNames, childAccount)
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
