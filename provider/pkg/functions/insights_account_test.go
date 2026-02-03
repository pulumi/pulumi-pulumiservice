// Copyright 2016-2025, Pulumi Corporation.

package functions

import (
	"context"
	"errors"
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

func (c *insightsAccountClientMock) GetInsightsAccount(
	ctx context.Context,
	orgName string,
	accountName string,
) (*pulumiapi.InsightsAccount, error) {
	if c.getInsightsAccountFunc == nil {
		return nil, nil
	}
	return c.getInsightsAccountFunc(ctx, orgName, accountName)
}

func (c *insightsAccountClientMock) ListInsightsAccounts(
	ctx context.Context,
	orgName string,
) ([]pulumiapi.InsightsAccount, error) {
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

		ctx := config.WithMockClient(t.Context(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		resp, err := fn.Invoke(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, resources.InsightsAccountStateFromAPI(testOrg, *insightsAccount), resp.Output)
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

		ctx := config.WithMockClient(t.Context(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		_, err := fn.Invoke(ctx, req)

		assert.ErrorContains(t, err, fmt.Sprintf("insights account %q not found", testAccount))
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		t.Parallel()
		testOrg := "test-org"
		testAccount := "test-account"
		apiError := "API error: internal server error"
		mockedClient := &insightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, errors.New(apiError)
			},
		}

		ctx := config.WithMockClient(t.Context(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		_, err := fn.Invoke(ctx, req)

		assert.ErrorContains(t, err, fmt.Sprintf("failed to get insights account: %s", apiError))
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

		ctx := config.WithMockClient(t.Context(), mockedClient)

		fn := GetInsightsAccountFunction{}
		req := infer.FunctionRequest[GetInsightsAccountInput]{
			Input: GetInsightsAccountInput{
				OrganizationName: testOrg,
				AccountName:      testAccount,
			},
		}

		resp, err := fn.Invoke(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, resources.InsightsAccountStateFromAPI(testOrg, *insightsAccount), resp.Output)
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

		ctx := config.WithMockClient(t.Context(), mockedClient)

		fn := GetInsightsAccountsFunction{}
		req := infer.FunctionRequest[GetInsightsAccountsInput]{
			Input: GetInsightsAccountsInput{
				OrganizationName: testOrg,
			},
		}

		resp, err := fn.Invoke(ctx, req)

		require.NoError(t, err)
		require.Len(t, resp.Output.Accounts, len(insightsAccounts))

		assert.Equal(t, resources.InsightsAccountStateFromAPI(testOrg, awsAccount), resp.Output.Accounts[0])
		assert.Equal(t, resources.InsightsAccountStateFromAPI(testOrg, gcpAccount), resp.Output.Accounts[1])
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

		ctx := config.WithMockClient(t.Context(), mockedClient)

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
		apiError := "API error: unauthorized"
		mockedClient := &insightsAccountClientMock{
			listInsightsAccountsFunc: func(ctx context.Context, orgName string) ([]pulumiapi.InsightsAccount, error) {
				return nil, errors.New(apiError)
			},
		}

		ctx := config.WithMockClient(t.Context(), mockedClient)

		fn := GetInsightsAccountsFunction{}
		req := infer.FunctionRequest[GetInsightsAccountsInput]{
			Input: GetInsightsAccountsInput{
				OrganizationName: testOrg,
			},
		}

		_, err := fn.Invoke(ctx, req)

		assert.ErrorContains(t, err, fmt.Sprintf("failed to list insights accounts: %s", apiError))
	})
}
