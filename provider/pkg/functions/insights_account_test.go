// Copyright 2016-2025, Pulumi Corporation.

package functions

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type insightsAccountClientMock struct {
	config.Client
	getInsightsAccountFunc   func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error)
	listInsightsAccountsFunc func(ctx context.Context, orgName string) ([]pulumiapi.InsightsAccount, error)
}

func (c *insightsAccountClientMock) GetInsightsAccount(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
	if c.getInsightsAccountFunc == nil {
		return nil, nil
	}
	return c.getInsightsAccountFunc(ctx, orgName, accountName)
}

func (c *insightsAccountClientMock) ListInsightsAccounts(ctx context.Context, orgName string) ([]pulumiapi.InsightsAccount, error) {
	if c.listInsightsAccountsFunc == nil {
		return nil, nil
	}
	return c.listInsightsAccountsFunc(ctx, orgName)
}

func TestGetInsightsAccountFunction(t *testing.T) {
	t.Parallel()

	t.Run("successfully gets an account", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		testAccount := "test-account"
		insightsAccount := &pulumiapi.InsightsAccount{
			ID:                   "account-id-123",
			Name:                 testAccount,
			Provider:             string(resources.CloudProviderAWS),
			ProviderEnvRef:       "test-env",
			ScheduledScanEnabled: true,
			ProviderConfig: map[string]interface{}{
				"regions": []interface{}{"us-west-2"},
			},
		}
		mockedClient := &insightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				assert.Equal(t, testOrg, orgName)
				assert.Equal(t, testAccount, accountName)
				return insightsAccount, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		resp, err := fn.Invoke(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, testOrg, resp.Output.OrganizationName)
		assert.Equal(t, testAccount, resp.Output.AccountName)
		assert.Equal(t, resources.CloudProvider(insightsAccount.Provider), resp.Output.Provider)
		assert.Equal(t, insightsAccount.ProviderEnvRef, resp.Output.Environment)
		assert.Equal(t, resources.ScanScheduleDaily, resp.Output.ScanSchedule)
		assert.Equal(t, insightsAccount.ID, resp.Output.InsightsAccountId)
		assert.True(t, resp.Output.ScheduledScanEnabled)
		assert.Equal(t, insightsAccount.ProviderConfig, resp.Output.ProviderConfig)
	})

	t.Run("returns error when account not found", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		testAccount := "nonexistent-account"
		mockedClient := &insightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, nil // 404 - not found
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		_, err := fn.Invoke(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("insights account %q not found", testAccount))
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		testAccount := "test-account"
		expectedError := "internal server error"
		mockedClient := &insightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, fmt.Errorf("API error: %s", expectedError)
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		_, err := fn.Invoke(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get insights account")
		assert.Contains(t, err.Error(), expectedError)
	})

	t.Run("correctly maps scan schedule none", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		testAccount := "test-account"
		insightsAccount := &pulumiapi.InsightsAccount{
			ID:                   "account-id-123",
			Name:                 testAccount,
			Provider:             string(resources.CloudProviderAzure),
			ProviderEnvRef:       "test-env",
			ScheduledScanEnabled: false,
		}
		mockedClient := &insightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return insightsAccount, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		resp, err := fn.Invoke(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, resources.ScanScheduleNone, resp.Output.ScanSchedule)
		assert.Equal(t, insightsAccount.ScheduledScanEnabled, resp.Output.ScheduledScanEnabled)
		assert.Equal(t, resources.CloudProvider(insightsAccount.Provider), resp.Output.Provider)
	})
}

func TestGetInsightsAccountsFunction(t *testing.T) {
	t.Parallel()

	t.Run("successfully lists accounts", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		awsAccount := pulumiapi.InsightsAccount{
			ID:                   "account-1",
			Name:                 "aws-account",
			Provider:             string(resources.CloudProviderAWS),
			ProviderEnvRef:       "aws-env",
			ScheduledScanEnabled: true,
			ProviderConfig: map[string]interface{}{
				"regions": []interface{}{"us-west-2", "us-east-1"},
			},
		}
		gcpAccount := pulumiapi.InsightsAccount{
			ID:                   "account-2",
			Name:                 "gcp-account",
			Provider:             string(resources.CloudProviderGCP),
			ProviderEnvRef:       "gcp-env",
			ScheduledScanEnabled: false,
		}
		insightsAccounts := []pulumiapi.InsightsAccount{awsAccount, gcpAccount}
		mockedClient := &insightsAccountClientMock{
			listInsightsAccountsFunc: func(ctx context.Context, orgName string) ([]pulumiapi.InsightsAccount, error) {
				assert.Equal(t, testOrg, orgName)
				return insightsAccounts, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		fn := GetInsightsAccountsFunction{}
		req := infer.FunctionRequest[GetInsightsAccountsInput]{
			Input: GetInsightsAccountsInput{
				OrganizationName: testOrg,
			},
		}

		resp, err := fn.Invoke(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp.Output.Accounts, len(insightsAccounts))

		// Check first account (AWS)
		firstAccount := resp.Output.Accounts[0]
		assert.Equal(t, testOrg, firstAccount.OrganizationName)
		assert.Equal(t, awsAccount.Name, firstAccount.AccountName)
		assert.Equal(t, resources.CloudProvider(awsAccount.Provider), firstAccount.Provider)
		assert.Equal(t, awsAccount.ProviderEnvRef, firstAccount.Environment)
		assert.Equal(t, resources.ScanScheduleDaily, firstAccount.ScanSchedule)
		assert.Equal(t, awsAccount.ID, firstAccount.InsightsAccountId)
		assert.Equal(t, awsAccount.ScheduledScanEnabled, firstAccount.ScheduledScanEnabled)

		// Check second account (GCP)
		secondAccount := resp.Output.Accounts[1]
		assert.Equal(t, testOrg, secondAccount.OrganizationName)
		assert.Equal(t, gcpAccount.Name, secondAccount.AccountName)
		assert.Equal(t, resources.CloudProvider(gcpAccount.Provider), secondAccount.Provider)
		assert.Equal(t, gcpAccount.ProviderEnvRef, secondAccount.Environment)
		assert.Equal(t, resources.ScanScheduleNone, secondAccount.ScanSchedule)
		assert.Equal(t, gcpAccount.ID, secondAccount.InsightsAccountId)
		assert.Equal(t, gcpAccount.ScheduledScanEnabled, secondAccount.ScheduledScanEnabled)
	})

	t.Run("returns empty list when no accounts exist", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		emptyAccounts := []pulumiapi.InsightsAccount{}
		mockedClient := &insightsAccountClientMock{
			listInsightsAccountsFunc: func(ctx context.Context, orgName string) ([]pulumiapi.InsightsAccount, error) {
				return emptyAccounts, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		fn := GetInsightsAccountsFunction{}
		req := infer.FunctionRequest[GetInsightsAccountsInput]{
			Input: GetInsightsAccountsInput{
				OrganizationName: testOrg,
			},
		}

		resp, err := fn.Invoke(ctx, req)

		require.NoError(t, err)
		assert.Len(t, resp.Output.Accounts, len(emptyAccounts))
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		expectedError := "unauthorized"
		mockedClient := &insightsAccountClientMock{
			listInsightsAccountsFunc: func(ctx context.Context, orgName string) ([]pulumiapi.InsightsAccount, error) {
				return nil, fmt.Errorf("API error: %s", expectedError)
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		fn := GetInsightsAccountsFunction{}
		req := infer.FunctionRequest[GetInsightsAccountsInput]{
			Input: GetInsightsAccountsInput{
				OrganizationName: testOrg,
			},
		}

		_, err := fn.Invoke(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list insights accounts")
		assert.Contains(t, err.Error(), expectedError)
	})
}
