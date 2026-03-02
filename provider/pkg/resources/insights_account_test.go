// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type InsightsAccountClientMock struct {
	config.Client
	getInsightsAccountFunc func(
		ctx context.Context, orgName string, accountName string,
	) (*pulumiapi.InsightsAccount, error)
	createInsightsAccountFunc func(
		ctx context.Context, orgName string, accountName string, req pulumiapi.CreateInsightsAccountRequest,
	) error
	updateInsightsAccountFunc func(
		ctx context.Context, orgName string, accountName string, req pulumiapi.UpdateInsightsAccountRequest,
	) error
	deleteInsightsAccountFunc  func(ctx context.Context, orgName string, accountName string) error
	getInsightsAccountTagsFunc func(ctx context.Context, orgName string, accountName string) (map[string]string, error)
	setInsightsAccountTagsFunc func(ctx context.Context, orgName string, accountName string, tags map[string]string) error
}

func (c *InsightsAccountClientMock) GetInsightsAccount(
	ctx context.Context,
	orgName string,
	accountName string,
) (*pulumiapi.InsightsAccount, error) {
	if c.getInsightsAccountFunc == nil {
		return nil, nil
	}
	return c.getInsightsAccountFunc(ctx, orgName, accountName)
}

func (c *InsightsAccountClientMock) CreateInsightsAccount(
	ctx context.Context,
	orgName string,
	accountName string,
	req pulumiapi.CreateInsightsAccountRequest,
) error {
	if c.createInsightsAccountFunc == nil {
		return nil
	}
	return c.createInsightsAccountFunc(ctx, orgName, accountName, req)
}

func (c *InsightsAccountClientMock) UpdateInsightsAccount(
	ctx context.Context,
	orgName string,
	accountName string,
	req pulumiapi.UpdateInsightsAccountRequest,
) error {
	if c.updateInsightsAccountFunc == nil {
		return nil
	}
	return c.updateInsightsAccountFunc(ctx, orgName, accountName, req)
}

func (c *InsightsAccountClientMock) DeleteInsightsAccount(
	ctx context.Context,
	orgName string,
	accountName string,
) error {
	if c.deleteInsightsAccountFunc == nil {
		return nil
	}
	return c.deleteInsightsAccountFunc(ctx, orgName, accountName)
}

func (c *InsightsAccountClientMock) TriggerScan(
	_ context.Context,
	_ string,
	_ string,
) (*pulumiapi.TriggerScanResponse, error) {
	return &pulumiapi.TriggerScanResponse{
		WorkflowRun: pulumiapi.WorkflowRun{
			ID:     "test-scan-id",
			Status: "running",
		},
	}, nil
}

func (c *InsightsAccountClientMock) GetScanStatus(
	_ context.Context,
	_ string,
	_ string,
) (*pulumiapi.ScanStatusResponse, error) {
	return &pulumiapi.ScanStatusResponse{
		WorkflowRun: pulumiapi.WorkflowRun{
			ID:     "test-scan-id",
			Status: "succeeded",
		},
		ResourceCount: 42,
	}, nil
}

func (c *InsightsAccountClientMock) GetInsightsAccountTags(
	ctx context.Context,
	orgName string,
	accountName string,
) (map[string]string, error) {
	if c.getInsightsAccountTagsFunc == nil {
		return map[string]string{}, nil
	}
	return c.getInsightsAccountTagsFunc(ctx, orgName, accountName)
}

func (c *InsightsAccountClientMock) SetInsightsAccountTags(
	ctx context.Context,
	orgName string,
	accountName string,
	tags map[string]string,
) error {
	if c.setInsightsAccountTagsFunc == nil {
		return nil
	}
	return c.setInsightsAccountTagsFunc(ctx, orgName, accountName, tags)
}

func TestInsightsAccount(t *testing.T) {
	t.Parallel()
	t.Run("Read when the resource is not found", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
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
				InsightsAccountID: "test-account-id",
			},
		}

		resp, err := ia.Read(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, infer.ReadResponse[InsightsAccountInput, InsightsAccountState]{}, resp)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
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

		require.NoError(t, err)
		expectedCore := InsightsAccountCore{
			OrganizationName: "test-org",
			AccountName:      "test-account",
			Provider:         CloudProviderAWS,
			Environment:      "test-env",
			ScanSchedule:     ScanScheduleDaily,
			ProviderConfig: map[string]interface{}{
				"region": "us-west-2",
			},
			Tags: map[string]string{},
		}
		assert.Equal(t, infer.ReadResponse[InsightsAccountInput, InsightsAccountState]{
			ID:     "test-org/test-account",
			Inputs: InsightsAccountInput{InsightsAccountCore: expectedCore},
			State: InsightsAccountState{
				InsightsAccountCore:  expectedCore,
				InsightsAccountID:    "test-account-id",
				ScheduledScanEnabled: true,
			},
		}, resp)
	})

	t.Run("Read preserves nil providerConfig when API returns empty map", func(t *testing.T) {
		t.Parallel()
		// This test ensures we don't get spurious diffs during refresh
		// when providerConfig was not specified initially
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
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
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
				},
				InsightsAccountID: "test-account-id",
			},
		}

		resp, err := ia.Read(ctx, req)

		require.NoError(t, err)
		assert.Nil(
			t,
			resp.Inputs.ProviderConfig,
			"providerConfig should remain nil when input was nil and API returned empty map",
		)
		assert.Nil(t, resp.State.ProviderConfig, "providerConfig should remain nil in state too")
	})

	t.Run("Create with DryRun", func(t *testing.T) {
		t.Parallel()
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

		require.NoError(t, err)
		assert.Equal(t, infer.CreateResponse[InsightsAccountState]{
			ID: "test-org/test-account",
			Output: InsightsAccountState{
				InsightsAccountCore:  req.Inputs.InsightsAccountCore,
				InsightsAccountID:    "",
				ScheduledScanEnabled: true,
			},
		}, resp)
	})

	t.Run("Create successfully", func(t *testing.T) {
		t.Parallel()
		var capturedRequest pulumiapi.CreateInsightsAccountRequest
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, req pulumiapi.CreateInsightsAccountRequest,
			) error {
				capturedRequest = req
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, accountName string) (*pulumiapi.InsightsAccount, error) {
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

		require.NoError(t, err)
		assert.Equal(t, "test-org/test-account", resp.ID)
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountID)
		assert.Equal(t, true, resp.Output.ScheduledScanEnabled)
		assert.Equal(t, "aws", capturedRequest.Provider)
		assert.Equal(t, "test-env", capturedRequest.Environment)
		assert.Equal(t, "daily", capturedRequest.ScanSchedule)
	})

	t.Run("Create with scanSchedule none", func(t *testing.T) {
		t.Parallel()
		var capturedRequest pulumiapi.CreateInsightsAccountRequest
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, req pulumiapi.CreateInsightsAccountRequest,
			) error {
				capturedRequest = req
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, accountName string) (*pulumiapi.InsightsAccount, error) {
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

		require.NoError(t, err)
		assert.Equal(t, "none", capturedRequest.ScanSchedule)
		assert.Equal(t, false, resp.Output.ScheduledScanEnabled)
	})

	t.Run("Create fails with API error", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.CreateInsightsAccountRequest,
			) error {
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
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.CreateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
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
		assert.Equal(t, "", resp.Output.InsightsAccountID)
	})

	t.Run("Create succeeds but account not found after creation", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.CreateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
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
		assert.Equal(t, "", resp.Output.InsightsAccountID)
	})

	t.Run("Update with DryRun", func(t *testing.T) {
		t.Parallel()
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
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, infer.UpdateResponse[InsightsAccountState]{
			Output: InsightsAccountState{
				InsightsAccountCore:  req.Inputs.InsightsAccountCore,
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true, // State value preserved in DryRun
			},
		}, resp)
	})

	t.Run("Update successfully", func(t *testing.T) {
		t.Parallel()
		var capturedRequest pulumiapi.UpdateInsightsAccountRequest
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, req pulumiapi.UpdateInsightsAccountRequest,
			) error {
				capturedRequest = req
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, accountName string) (*pulumiapi.InsightsAccount, error) {
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
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "updated-env", capturedRequest.Environment)
		assert.Equal(t, "none", capturedRequest.ScanSchedule)
		assert.Equal(t, false, resp.Output.ScheduledScanEnabled)
	})

	t.Run("Update fails with API error", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.UpdateInsightsAccountRequest,
			) error {
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
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		_, err := ia.Update(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error updating insights account 'test-account'")
		assert.Contains(t, err.Error(), "environment not found")
	})

	t.Run("Update succeeds but GET fails after update", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.UpdateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
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
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "internal server error")
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountID)
	})

	t.Run("Update succeeds but account not found after update", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.UpdateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
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
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "insights account 'test-account' not found after update")
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountID)
	})

	t.Run("Delete successfully", func(t *testing.T) {
		t.Parallel()
		deleteCalled := false
		mockedClient := &InsightsAccountClientMock{
			deleteInsightsAccountFunc: func(_ context.Context, orgName string, accountName string) error {
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
				InsightsAccountID: "account-id-123",
			},
		}

		_, err := ia.Delete(ctx, req)

		require.NoError(t, err)
		assert.True(t, deleteCalled)
	})

	t.Run("Read returns tags from API", func(t *testing.T) {
		t.Parallel()
		expectedTags := map[string]string{
			"environment": "production",
			"team":        "platform",
		}
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "test-account-id",
					Name:                 "test-account",
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: true,
				}, nil
			},
			getInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string) (map[string]string, error) {
				return expectedTags, nil
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
					ScanSchedule:     ScanScheduleDaily,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "test-env",
					ScanSchedule:     ScanScheduleDaily,
				},
			},
		}

		resp, err := ia.Read(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, expectedTags, resp.State.Tags)
		assert.Equal(t, expectedTags, resp.Inputs.Tags)
	})

	t.Run("Read fails when GetInsightsAccountTags fails", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			getInsightsAccountFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "test-account-id",
					Name:                 "test-account",
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: true,
				}, nil
			},
			getInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string) (map[string]string, error) {
				return nil, fmt.Errorf("API error: failed to fetch tags")
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
				},
			},
			State: InsightsAccountState{},
		}

		_, err := ia.Read(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get tags for InsightsAccount")
	})

	t.Run("Create without tags does not call SetInsightsAccountTags", func(t *testing.T) {
		t.Parallel()
		setTagsCalled := false
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.CreateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "account-id-123",
					Name:                 accountName,
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: true,
				}, nil
			},
			setInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string, _ map[string]string) error {
				setTagsCalled = true
				return nil
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

		require.NoError(t, err)
		assert.False(t, setTagsCalled, "SetInsightsAccountTags should not be called when Tags is nil/empty")
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountID)
	})

	t.Run("Create with tags", func(t *testing.T) {
		t.Parallel()
		expectedTags := map[string]string{
			"environment": "staging",
			"cost-center": "engineering",
		}
		var capturedTags map[string]string
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.CreateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "account-id-123",
					Name:                 accountName,
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: true,
				}, nil
			},
			setInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string, tags map[string]string) error {
				capturedTags = tags
				return nil
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
					Tags:             expectedTags,
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, expectedTags, capturedTags)
		assert.Equal(t, expectedTags, resp.Output.Tags)
	})

	t.Run("Create fails when SetInsightsAccountTags fails", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			createInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.CreateInsightsAccountRequest,
			) error {
				return nil
			},
			setInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string, _ map[string]string) error {
				return fmt.Errorf("API error: failed to set tags")
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
					Tags: map[string]string{
						"environment": "staging",
					},
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "failed to set tags")
		assert.Equal(t, "test-org/test-account", resp.ID)
	})

	t.Run("Update with empty tags calls SetInsightsAccountTags to clear them", func(t *testing.T) {
		t.Parallel()
		var capturedTags map[string]string
		setTagsCalled := false
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.UpdateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "account-id-123",
					Name:                 accountName,
					Provider:             "aws",
					ProviderEnvRef:       "updated-env",
					ScheduledScanEnabled: true,
				}, nil
			},
			setInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string, tags map[string]string) error {
				setTagsCalled = true
				capturedTags = tags
				return nil
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
					ScanSchedule:     ScanScheduleDaily,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "old-env",
					ScanSchedule:     ScanScheduleDaily,
					Tags: map[string]string{
						"environment": "staging",
					},
				},
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		require.NoError(t, err)
		assert.True(t, setTagsCalled, "SetInsightsAccountTags should be called even when Tags is nil/empty")
		assert.Nil(t, capturedTags, "Tags should be nil when clearing")
		assert.Nil(t, resp.Output.Tags)
	})

	t.Run("Update with tags", func(t *testing.T) {
		t.Parallel()
		expectedTags := map[string]string{
			"environment": "production",
			"team":        "platform",
		}
		var capturedTags map[string]string
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.UpdateInsightsAccountRequest,
			) error {
				return nil
			},
			getInsightsAccountFunc: func(_ context.Context, _ string, accountName string) (*pulumiapi.InsightsAccount, error) {
				return &pulumiapi.InsightsAccount{
					ID:                   "account-id-123",
					Name:                 accountName,
					Provider:             "aws",
					ProviderEnvRef:       "updated-env",
					ScheduledScanEnabled: true,
				}, nil
			},
			setInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string, tags map[string]string) error {
				capturedTags = tags
				return nil
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
					ScanSchedule:     ScanScheduleDaily,
					Tags:             expectedTags,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         CloudProviderAWS,
					Environment:      "old-env",
					ScanSchedule:     ScanScheduleDaily,
					Tags: map[string]string{
						"environment": "staging",
					},
				},
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, expectedTags, capturedTags)
		assert.Equal(t, expectedTags, resp.Output.Tags)
	})

	t.Run("Update fails when SetInsightsAccountTags fails", func(t *testing.T) {
		t.Parallel()
		mockedClient := &InsightsAccountClientMock{
			updateInsightsAccountFunc: func(
				_ context.Context, _ string, _ string, _ pulumiapi.UpdateInsightsAccountRequest,
			) error {
				return nil
			},
			setInsightsAccountTagsFunc: func(_ context.Context, _ string, _ string, _ map[string]string) error {
				return fmt.Errorf("API error: failed to set tags")
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
					ScanSchedule:     ScanScheduleDaily,
					Tags: map[string]string{
						"environment": "production",
					},
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
				InsightsAccountID:    "account-id-123",
				ScheduledScanEnabled: true,
			},
		}

		resp, err := ia.Update(ctx, req)

		assert.Error(t, err)
		var initErr infer.ResourceInitFailedError
		assert.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "failed to set tags")
		assert.Equal(t, "account-id-123", resp.Output.InsightsAccountID)
	})

}
