// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock types for InsightsAccount tests
type getInsightsAccountFunc func() (*pulumiapi.InsightsAccount, error)
type createInsightsAccountFunc func() error
type updateInsightsAccountFunc func() error
type deleteInsightsAccountFunc func() error

type insightsAccountClientMock struct {
	getFunc    getInsightsAccountFunc
	createFunc createInsightsAccountFunc
	updateFunc updateInsightsAccountFunc
	deleteFunc deleteInsightsAccountFunc
}

func (c *insightsAccountClientMock) GetInsightsAccount(ctx context.Context, orgName, accountName string) (*pulumiapi.InsightsAccount, error) {
	if c.getFunc == nil {
		return nil, nil
	}
	return c.getFunc()
}

func (c *insightsAccountClientMock) CreateInsightsAccount(ctx context.Context, orgName, accountName string, req pulumiapi.CreateInsightsAccountRequest) error {
	if c.createFunc == nil {
		return nil
	}
	return c.createFunc()
}

func (c *insightsAccountClientMock) UpdateInsightsAccount(ctx context.Context, orgName, accountName string, req pulumiapi.UpdateInsightsAccountRequest) error {
	if c.updateFunc == nil {
		return nil
	}
	return c.updateFunc()
}

func (c *insightsAccountClientMock) DeleteInsightsAccount(ctx context.Context, orgName, accountName string) error {
	if c.deleteFunc == nil {
		return nil
	}
	return c.deleteFunc()
}

func (c *insightsAccountClientMock) TriggerScan(ctx context.Context, orgName, accountName string) (*pulumiapi.TriggerScanResponse, error) {
	return &pulumiapi.TriggerScanResponse{
		WorkflowRun: pulumiapi.WorkflowRun{
			ID:        "test-scan-123",
			Status:    "running",
			StartedAt: "2025-11-12T15:30:00Z",
		},
	}, nil
}

func (c *insightsAccountClientMock) GetScanStatus(ctx context.Context, orgName, accountName string) (*pulumiapi.ScanStatusResponse, error) {
	return &pulumiapi.ScanStatusResponse{
		WorkflowRun: pulumiapi.WorkflowRun{
			ID:         "scan-456",
			Status:     "succeeded",
			FinishedAt: "2025-11-12T14:00:00Z",
		},
		NextScan:      "2025-11-13T02:00:00Z",
		ResourceCount: 42,
	}, nil
}

func buildInsightsAccountClientMock(
	getFunc getInsightsAccountFunc,
	createFunc createInsightsAccountFunc,
	updateFunc updateInsightsAccountFunc,
	deleteFunc deleteInsightsAccountFunc,
) *insightsAccountClientMock {
	return &insightsAccountClientMock{
		getFunc:    getFunc,
		createFunc: createFunc,
		updateFunc: updateFunc,
		deleteFunc: deleteFunc,
	}
}

// Test helper functions
func testInsightsAccountInput() PulumiServiceInsightsAccountInput {
	return PulumiServiceInsightsAccountInput{
		OrgName:     "test-org",
		AccountName: "test-account",
		Provider:    "aws",
		Environment: "test-env",
		Cron:        "0 0 * * *",
		ProviderConfig: map[string]interface{}{
			"region": "us-west-2",
		},
	}
}

func testInsightsAccountResponse() *pulumiapi.InsightsAccount {
	return &pulumiapi.InsightsAccount{
		ID:                   "test-account-id",
		Name:                 "test-account",
		Provider:             "aws",
		ProviderVersion:      "6.0.0",
		ProviderEnvRef:       "test-env",
		ScheduledScanEnabled: true,
		ProviderConfig: map[string]interface{}{
			"region": "us-west-2",
		},
	}
}

// Read Tests
func TestInsightsAccount_Read_NotFound(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		func() (*pulumiapi.InsightsAccount, error) { return nil, nil },
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
	}

	inputProperties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-account",
		Properties: inputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Empty(t, resp.Id)
	assert.Nil(t, resp.Properties)
}

func TestInsightsAccount_Read_Found(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		func() (*pulumiapi.InsightsAccount, error) {
			return testInsightsAccountResponse(), nil
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
		"cron":             resource.NewPropertyValue(input.Cron),
		"providerConfig":   resource.NewPropertyValue(input.ProviderConfig),
	}

	inputProperties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-account",
		Properties: inputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-account", resp.Id)
	assert.NotNil(t, resp.Properties)
	assert.NotNil(t, resp.Inputs)
}

func TestInsightsAccount_Read_InvalidId(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(nil, nil, nil, nil)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	inputMap := resource.PropertyMap{}
	inputProperties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "invalid-id-format",
		Properties: inputProperties,
	}

	resp, err := provider.Read(&req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid, must be in the format")
	assert.Nil(t, resp)
}

// Create Tests
func TestInsightsAccount_Create_Success(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		func() (*pulumiapi.InsightsAccount, error) {
			return testInsightsAccountResponse(), nil
		},
		func() error { return nil },
		nil,
		nil,
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
		"cron":             resource.NewPropertyValue(input.Cron),
		"providerConfig":   resource.NewPropertyValue(input.ProviderConfig),
	}

	properties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: properties,
	}

	resp, err := provider.Create(&req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-account", resp.Id)
	assert.NotNil(t, resp.Properties)
}

func TestInsightsAccount_Create_APIError(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		nil,
		func() error { return errors.New("API error") },
		nil,
		nil,
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
	}

	properties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: properties,
	}

	resp, err := provider.Create(&req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating insights account")
	assert.Nil(t, resp)
}

func TestInsightsAccount_Create_NotFoundAfterCreate(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		func() (*pulumiapi.InsightsAccount, error) {
			return nil, nil
		},
		func() error { return nil },
		nil,
		nil,
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
	}

	properties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: properties,
	}

	resp, err := provider.Create(&req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found after creation")
	assert.Nil(t, resp)
}

// Update Tests
func TestInsightsAccount_Update_Success(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		func() (*pulumiapi.InsightsAccount, error) {
			account := testInsightsAccountResponse()
			account.ProviderEnvRef = "new-env"
			return account, nil
		},
		nil,
		func() error { return nil },
		nil,
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	input := testInsightsAccountInput()
	input.Environment = "new-env"

	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
		"cron":             resource.NewPropertyValue(input.Cron),
		"providerConfig":   resource.NewPropertyValue(input.ProviderConfig),
	}

	properties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-account",
		News: properties,
	}

	resp, err := provider.Update(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp.Properties)
}

func TestInsightsAccount_Update_NotFoundAfterUpdate(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		func() (*pulumiapi.InsightsAccount, error) {
			return nil, nil
		},
		nil,
		func() error { return nil },
		nil,
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
	}

	properties, _ := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-account",
		News: properties,
	}

	resp, err := provider.Update(&req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found after update")
	assert.Nil(t, resp)
}

// Delete Tests
func TestInsightsAccount_Delete_Success(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(
		nil,
		nil,
		nil,
		func() error { return nil },
	)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	req := pulumirpc.DeleteRequest{
		Id: "test-org/test-account",
	}

	resp, err := provider.Delete(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestInsightsAccount_Delete_InvalidId(t *testing.T) {
	mockedClient := buildInsightsAccountClientMock(nil, nil, nil, nil)

	provider := PulumiServiceInsightsAccountResource{
		Client: mockedClient,
	}

	req := pulumirpc.DeleteRequest{
		Id: "invalid-id",
	}

	_, err := provider.Delete(&req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid, must be in the format")
}

// Diff Tests - Replacement Properties
func TestInsightsAccount_Diff_OrganizationNameChange_RequiresReplacement(t *testing.T) {
	provider := PulumiServiceInsightsAccountResource{}

	oldInput := testInsightsAccountInput()
	oldInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(oldInput.OrgName),
		"accountName":      resource.NewPropertyValue(oldInput.AccountName),
		"provider":         resource.NewPropertyValue(oldInput.Provider),
		"environment":      resource.NewPropertyValue(oldInput.Environment),
	}

	newInput := testInsightsAccountInput()
	newInput.OrgName = "new-org"
	newInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(newInput.OrgName),
		"accountName":      resource.NewPropertyValue(newInput.AccountName),
		"provider":         resource.NewPropertyValue(newInput.Provider),
		"environment":      resource.NewPropertyValue(newInput.Environment),
	}

	oldInputs, _ := plugin.MarshalProperties(oldInputMap, plugin.MarshalOptions{})
	news, _ := plugin.MarshalProperties(newInputMap, plugin.MarshalOptions{})

	req := pulumirpc.DiffRequest{
		OldInputs: oldInputs,
		News:      news,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
	assert.True(t, resp.HasDetailedDiff)
	assert.Contains(t, resp.DetailedDiff, "organizationName")
	assert.Equal(t, pulumirpc.PropertyDiff_UPDATE_REPLACE, resp.DetailedDiff["organizationName"].Kind)
}

func TestInsightsAccount_Diff_AccountNameChange_RequiresReplacement(t *testing.T) {
	provider := PulumiServiceInsightsAccountResource{}

	oldInput := testInsightsAccountInput()
	oldInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(oldInput.OrgName),
		"accountName":      resource.NewPropertyValue(oldInput.AccountName),
		"provider":         resource.NewPropertyValue(oldInput.Provider),
		"environment":      resource.NewPropertyValue(oldInput.Environment),
	}

	newInput := testInsightsAccountInput()
	newInput.AccountName = "new-account"
	newInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(newInput.OrgName),
		"accountName":      resource.NewPropertyValue(newInput.AccountName),
		"provider":         resource.NewPropertyValue(newInput.Provider),
		"environment":      resource.NewPropertyValue(newInput.Environment),
	}

	oldInputs, _ := plugin.MarshalProperties(oldInputMap, plugin.MarshalOptions{})
	news, _ := plugin.MarshalProperties(newInputMap, plugin.MarshalOptions{})

	req := pulumirpc.DiffRequest{
		OldInputs: oldInputs,
		News:      news,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
	assert.True(t, resp.HasDetailedDiff)
	assert.Contains(t, resp.DetailedDiff, "accountName")
	assert.Equal(t, pulumirpc.PropertyDiff_UPDATE_REPLACE, resp.DetailedDiff["accountName"].Kind)
}

func TestInsightsAccount_Diff_ProviderChange_RequiresReplacement(t *testing.T) {
	provider := PulumiServiceInsightsAccountResource{}

	oldInput := testInsightsAccountInput()
	oldInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(oldInput.OrgName),
		"accountName":      resource.NewPropertyValue(oldInput.AccountName),
		"provider":         resource.NewPropertyValue(oldInput.Provider),
		"environment":      resource.NewPropertyValue(oldInput.Environment),
	}

	newInput := testInsightsAccountInput()
	newInput.Provider = "azure"
	newInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(newInput.OrgName),
		"accountName":      resource.NewPropertyValue(newInput.AccountName),
		"provider":         resource.NewPropertyValue(newInput.Provider),
		"environment":      resource.NewPropertyValue(newInput.Environment),
	}

	oldInputs, _ := plugin.MarshalProperties(oldInputMap, plugin.MarshalOptions{})
	news, _ := plugin.MarshalProperties(newInputMap, plugin.MarshalOptions{})

	req := pulumirpc.DiffRequest{
		OldInputs: oldInputs,
		News:      news,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
	assert.True(t, resp.HasDetailedDiff)
	assert.Contains(t, resp.DetailedDiff, "provider")
	assert.Equal(t, pulumirpc.PropertyDiff_UPDATE_REPLACE, resp.DetailedDiff["provider"].Kind)
}

// Diff Tests - Update-Only Properties
func TestInsightsAccount_Diff_EnvironmentChange_UpdateOnly(t *testing.T) {
	provider := PulumiServiceInsightsAccountResource{}

	oldInput := testInsightsAccountInput()
	oldInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(oldInput.OrgName),
		"accountName":      resource.NewPropertyValue(oldInput.AccountName),
		"provider":         resource.NewPropertyValue(oldInput.Provider),
		"environment":      resource.NewPropertyValue(oldInput.Environment),
		"cron":             resource.NewPropertyValue(oldInput.Cron),
		"providerConfig":   resource.NewPropertyValue(oldInput.ProviderConfig),
	}

	newInput := testInsightsAccountInput()
	newInput.Environment = "new-env"
	newInputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(newInput.OrgName),
		"accountName":      resource.NewPropertyValue(newInput.AccountName),
		"provider":         resource.NewPropertyValue(newInput.Provider),
		"environment":      resource.NewPropertyValue(newInput.Environment),
		"cron":             resource.NewPropertyValue(newInput.Cron),
		"providerConfig":   resource.NewPropertyValue(newInput.ProviderConfig),
	}

	oldInputs, _ := plugin.MarshalProperties(oldInputMap, plugin.MarshalOptions{})
	news, _ := plugin.MarshalProperties(newInputMap, plugin.MarshalOptions{})

	req := pulumirpc.DiffRequest{
		OldInputs: oldInputs,
		News:      news,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
	assert.True(t, resp.HasDetailedDiff)
	assert.Contains(t, resp.DetailedDiff, "environment")
	// Environment change should be UPDATE, not REPLACE
	assert.Equal(t, pulumirpc.PropertyDiff_UPDATE, resp.DetailedDiff["environment"].Kind)
}

func TestInsightsAccount_Diff_NoChanges(t *testing.T) {
	provider := PulumiServiceInsightsAccountResource{}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
		"cron":             resource.NewPropertyValue(input.Cron),
		"providerConfig":   resource.NewPropertyValue(input.ProviderConfig),
	}

	oldInputs, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})
	news, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

	req := pulumirpc.DiffRequest{
		OldInputs: oldInputs,
		News:      news,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
	assert.False(t, resp.HasDetailedDiff)
}

// Check Tests
func TestInsightsAccount_Check_Success(t *testing.T) {
	provider := PulumiServiceInsightsAccountResource{}

	input := testInsightsAccountInput()
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue(input.OrgName),
		"accountName":      resource.NewPropertyValue(input.AccountName),
		"provider":         resource.NewPropertyValue(input.Provider),
		"environment":      resource.NewPropertyValue(input.Environment),
		"cron":             resource.NewPropertyValue(input.Cron),
		"providerConfig":   resource.NewPropertyValue(input.ProviderConfig),
	}

	news, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

	req := pulumirpc.CheckRequest{
		News: news,
	}

	resp, err := provider.Check(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Failures)
}

// Helper function tests
func TestSplitInsightsAccountId_Valid(t *testing.T) {
	orgName, accountName, err := splitInsightsAccountId("test-org/test-account")

	assert.NoError(t, err)
	assert.Equal(t, "test-org", orgName)
	assert.Equal(t, "test-account", accountName)
}

func TestSplitInsightsAccountId_InvalidFormat(t *testing.T) {
	_, _, err := splitInsightsAccountId("invalid-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid, must be in the format")
}

// GenerateInsightsAccountProperties Tests
func TestGenerateInsightsAccountProperties_Success(t *testing.T) {
	input := testInsightsAccountInput()
	account := testInsightsAccountResponse()

	outputs, inputs, err := GenerateInsightsAccountProperties(input, *account)

	require.NoError(t, err)
	require.NotNil(t, outputs)
	require.NotNil(t, inputs)
}

// Name Test
func TestInsightsAccount_Name(t *testing.T) {
	provider := PulumiServiceInsightsAccountResource{}

	name := provider.Name()

	assert.Equal(t, "pulumiservice:index:InsightsAccount", name)
}

