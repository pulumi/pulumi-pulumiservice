// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/stretchr/testify/assert"
)

type InsightsAccountClientMock struct {
	config.Client
	getInsightsAccountFunc    func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error)
	createInsightsAccountFunc func(ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest) error
	updateInsightsAccountFunc func(ctx context.Context, orgName string, accountName string, req pulumiapi.UpdateInsightsAccountRequest) error
	deleteInsightsAccountFunc func(ctx context.Context, orgName string, accountName string) error
}

func (c *InsightsAccountClientMock) GetInsightsAccount(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
	if c.getInsightsAccountFunc == nil {
		return nil, nil
	}
	return c.getInsightsAccountFunc(ctx, orgName, accountName)
}

func (c *InsightsAccountClientMock) CreateInsightsAccount(ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest) error {
	if c.createInsightsAccountFunc == nil {
		return nil
	}
	return c.createInsightsAccountFunc(ctx, orgName, accountName, req)
}

func (c *InsightsAccountClientMock) UpdateInsightsAccount(ctx context.Context, orgName string, accountName string, req pulumiapi.UpdateInsightsAccountRequest) error {
	if c.updateInsightsAccountFunc == nil {
		return nil
	}
	return c.updateInsightsAccountFunc(ctx, orgName, accountName, req)
}

func (c *InsightsAccountClientMock) DeleteInsightsAccount(ctx context.Context, orgName string, accountName string) error {
	if c.deleteInsightsAccountFunc == nil {
		return nil
	}
	return c.deleteInsightsAccountFunc(ctx, orgName, accountName)
}

func (c *InsightsAccountClientMock) TriggerScan(ctx context.Context, orgName string, accountName string) (*pulumiapi.TriggerScanResponse, error) {
	return &pulumiapi.TriggerScanResponse{
		WorkflowRun: pulumiapi.WorkflowRun{
			ID:     "test-scan-id",
			Status: "running",
		},
	}, nil
}

func (c *InsightsAccountClientMock) GetScanStatus(ctx context.Context, orgName string, accountName string) (*pulumiapi.ScanStatusResponse, error) {
	return &pulumiapi.ScanStatusResponse{
		WorkflowRun: pulumiapi.WorkflowRun{
			ID:     "test-scan-id",
			Status: "succeeded",
		},
		ResourceCount: 42,
	}, nil
}

func TestInsightsAccount(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		scanSchedule := ScanScheduleDaily
		req := infer.ReadRequest[InsightsAccountInput, InsightsAccountState]{
			ID: "test-org/test-account",
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
				},
				InsightsAccountId: "test-account-id",
			},
		}

		resp, err := ia.Read(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, InsightsAccountInput{}, resp.Inputs)
		assert.Equal(t, InsightsAccountState{}, resp.State)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "test-account-id",
					Name:                 "test-account",
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: true,
					ProviderConfig: map[string]interface{}{
						"region": "us-west-2",
					},
				}, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		scanSchedule := ScanScheduleDaily
		req := infer.ReadRequest[InsightsAccountInput, InsightsAccountState]{
			ID: "test-org/test-account",
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
				},
			},
		}

		resp, err := ia.Read(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-account", resp.ID)
		assert.Equal(t, "test-org", resp.Inputs.OrganizationName)
		assert.Equal(t, "test-account", resp.Inputs.AccountName)
		assert.Equal(t, CloudProviderAWS, resp.Inputs.Provider)
		assert.Equal(t, "test-env", resp.Inputs.Environment)
		assert.Equal(t, ScanScheduleDaily, resp.Inputs.ScanSchedule)
		assert.Equal(t, "test-account-id", resp.State.InsightsAccountId)
		assert.Equal(t, true, resp.State.ScheduledScanEnabled)
		assert.Equal(t, "us-west-2", resp.State.ProviderConfig["region"])
	})

	t.Run("Read preserves nil providerConfig when API returns empty map", func(t *testing.T) {
		// This test ensures we don't get spurious diffs during refresh
		// when providerConfig was not specified initially
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "test-account-id",
					Name:                 "test-account",
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: false,
					ProviderConfig:       map[string]interface{}{}, // API returns empty map
				}, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		req := infer.ReadRequest[InsightsAccountInput, InsightsAccountState]{
			ID: "test-org/test-account",
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					// providerConfig is nil - not specified
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					// providerConfig is nil
				},
				InsightsAccountId: "test-account-id",
			},
		}

		resp, err := ia.Read(ctx, req)

		assert.NoError(t, err)
		assert.Nil(t, resp.Inputs.ProviderConfig, "providerConfig should remain nil when input was nil and API returned empty map")
		assert.Nil(t, resp.State.ProviderConfig, "providerConfig should remain nil in state too")
	})

	t.Run("Create with DryRun", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		scanSchedule := ScanScheduleDaily
		req := infer.CreateRequest[InsightsAccountInput]{
			DryRun: true,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
					ProviderConfig: map[string]interface{}{
						"regions": []string{"us-west-2"},
					},
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-account", resp.ID)
		assert.Equal(t, "", resp.Output.InsightsAccountId)
		assert.Equal(t, true, resp.Output.ScheduledScanEnabled)
		assert.Equal(t, "test-org", resp.Output.OrganizationName)
		assert.Equal(t, "test-account", resp.Output.AccountName)
	})

	t.Run("Create successfully", func(t *testing.T) {
		var capturedRequest pulumiapi.CreateInsightsAccountRequest
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest) error {
				capturedRequest = req
				return nil
			},
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "account-id-123",
					Name:                 accountName,
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: true,
					ProviderConfig: map[string]interface{}{
						"regions": []interface{}{"us-west-2"},
					},
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		scanSchedule := ScanScheduleDaily
		req := infer.CreateRequest[InsightsAccountInput]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
					ProviderConfig: map[string]interface{}{
						"regions": []string{"us-west-2"},
					},
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-account", resp.ID)
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountId)
		assert.Equal(t, true, resp.Output.ScheduledScanEnabled)
		assert.Equal(t, "aws", capturedRequest.Provider)
		assert.Equal(t, "test-env", capturedRequest.Environment)
		assert.Equal(t, "daily", capturedRequest.ScanSchedule)
	})

	t.Run("Create with scanSchedule none", func(t *testing.T) {
		var capturedRequest pulumiapi.CreateInsightsAccountRequest
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest) error {
				capturedRequest = req
				return nil
			},
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "account-id-123",
					Name:                 accountName,
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: false,
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		scanSchedule := ScanScheduleNone
		req := infer.CreateRequest[InsightsAccountInput]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "none", capturedRequest.ScanSchedule)
		assert.Equal(t, false, resp.Output.ScheduledScanEnabled)
	})

	t.Run("Create fails with API error", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest) error {
				return fmt.Errorf("API error: invalid environment reference")
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		req := infer.CreateRequest[InsightsAccountInput]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "invalid-env",
					ScanSchedule:     ScanScheduleDaily,
				},
			},
		}

		_, err := ia.Create(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error creating insights account 'test-account'")
		assert.Contains(t, err.Error(), "invalid environment reference")
	})

	t.Run("Create succeeds but GET fails after creation", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest) error {
				return nil
			},
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, fmt.Errorf("API error: internal server error")
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		req := infer.CreateRequest[InsightsAccountInput]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     ScanScheduleDaily,
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "internal server error")
		assert.Equal(t, "test-org/test-account", resp.ID)
		assert.Equal(t, "", resp.Output.InsightsAccountId)
	})

	t.Run("Create succeeds but account not found after creation", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest) error {
				return nil
			},
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, nil // 404 - not found
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		req := infer.CreateRequest[InsightsAccountInput]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     ScanScheduleDaily,
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "insights account 'test-account' not found after creation")
		assert.Equal(t, "test-org/test-account", resp.ID)
		assert.Equal(t, "", resp.Output.InsightsAccountId)
	})

	t.Run("Update with DryRun", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		oldSchedule := ScanScheduleDaily
		newSchedule := ScanScheduleNone
		req := infer.UpdateRequest[InsightsAccountInput, InsightsAccountState]{
			DryRun: true,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "updated-env",
					ScanSchedule:     newSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "old-env",
					ScanSchedule:     oldSchedule,
				},
				InsightsAccountId:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "updated-env", resp.Output.Environment)
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountId)
		assert.Equal(t, true, resp.Output.ScheduledScanEnabled) // State value preserved in DryRun
	})

	t.Run("Update successfully", func(t *testing.T) {
		var capturedRequest pulumiapi.UpdateInsightsAccountRequest
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.UpdateInsightsAccountRequest) error {
				capturedRequest = req
				return nil
			},
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "account-id-123",
					Name:                 accountName,
					Provider:             "aws",
					ProviderEnvRef:       "updated-env",
					ScheduledScanEnabled: false,
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		oldSchedule := ScanScheduleDaily
		newSchedule := ScanScheduleNone
		req := infer.UpdateRequest[InsightsAccountInput, InsightsAccountState]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "updated-env",
					ScanSchedule:     newSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "old-env",
					ScanSchedule:     oldSchedule,
				},
				InsightsAccountId:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "updated-env", capturedRequest.Environment)
		assert.Equal(t, "none", capturedRequest.ScanSchedule)
		assert.Equal(t, false, resp.Output.ScheduledScanEnabled)
	})

	t.Run("Update fails with API error", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.UpdateInsightsAccountRequest) error {
				return fmt.Errorf("API error: environment not found")
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		req := infer.UpdateRequest[InsightsAccountInput, InsightsAccountState]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "invalid-env",
					ScanSchedule:     ScanScheduleNone,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "old-env",
					ScanSchedule:     ScanScheduleDaily,
				},
				InsightsAccountId:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		_, err := ia.Update(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error updating insights account 'test-account'")
		assert.Contains(t, err.Error(), "environment not found")
	})

	t.Run("Update succeeds but GET fails after update", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.UpdateInsightsAccountRequest) error {
				return nil
			},
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, fmt.Errorf("API error: internal server error")
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		req := infer.UpdateRequest[InsightsAccountInput, InsightsAccountState]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "updated-env",
					ScanSchedule:     ScanScheduleNone,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "old-env",
					ScanSchedule:     ScanScheduleDaily,
				},
				InsightsAccountId:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "internal server error")
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountId)
	})

	t.Run("Update succeeds but account not found after update", func(t *testing.T) {
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string, req pulumiapi.UpdateInsightsAccountRequest) error {
				return nil
			},
			getInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return nil, nil // 404 - not found
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		req := infer.UpdateRequest[InsightsAccountInput, InsightsAccountState]{
			DryRun: false,
			Inputs: InsightsAccountInput{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "updated-env",
					ScanSchedule:     ScanScheduleNone,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "old-env",
					ScanSchedule:     ScanScheduleDaily,
				},
				InsightsAccountId:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "insights account 'test-account' not found after update")
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountId)
	})

	t.Run("Delete successfully", func(t *testing.T) {
		deleteCalled := false
		mockedClient := &InsightsAccountClientMock{
			deleteInsightsAccountFunc: func(ctx context.Context, orgName string, accountName string) error {
				deleteCalled = true
				assert.Equal(t, "test-org", orgName)
				assert.Equal(t, "test-account", accountName)
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mockedClient)

		ia := &InsightsAccount{}
		scanSchedule := ScanScheduleDaily
		req := infer.DeleteRequest[InsightsAccountState]{
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     scanSchedule,
				},
				InsightsAccountId: "account-id-123",
			},
		}

		_, err := ia.Delete(ctx, req)

		assert.NoError(t, err)
		assert.True(t, deleteCalled)
	})

}
