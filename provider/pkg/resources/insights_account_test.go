// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
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
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
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
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
				},
			},
		}

		resp, err := ia.Read(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-account", resp.ID)
		assert.Equal(t, "test-org", resp.Inputs.OrganizationName)
		assert.Equal(t, "test-account", resp.Inputs.AccountName)
		assert.Equal(t, "aws", resp.Inputs.Provider)
		assert.Equal(t, "test-env", resp.Inputs.Environment)
		assert.Equal(t, ScanScheduleDaily, *resp.Inputs.ScanSchedule)
		assert.Equal(t, "test-account-id", resp.State.InsightsAccountId)
		assert.Equal(t, true, resp.State.ScheduledScanEnabled)
		assert.Equal(t, "us-west-2", resp.State.ProviderConfig["region"])
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
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
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
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
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
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
				},
			},
		}

		resp, err := ia.Create(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "none", capturedRequest.ScanSchedule)
		assert.Equal(t, false, resp.Output.ScheduledScanEnabled)
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
					Provider:         "aws",
					Environment:      "updated-env",
					ScanSchedule:     &newSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         "aws",
					Environment:      "old-env",
					ScanSchedule:     &oldSchedule,
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
					Provider:         "aws",
					Environment:      "updated-env",
					ScanSchedule:     &newSchedule,
				},
			},
			State: InsightsAccountState{
				InsightsAccountCore: InsightsAccountCore{
					OrganizationName: "test-org",
					AccountName:      "test-account",
					Provider:         "aws",
					Environment:      "old-env",
					ScanSchedule:     &oldSchedule,
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
					Provider:         "aws",
					Environment:      "test-env",
					ScanSchedule:     &scanSchedule,
				},
				InsightsAccountId: "account-id-123",
			},
		}

		_, err := ia.Delete(ctx, req)

		assert.NoError(t, err)
		assert.True(t, deleteCalled)
	})

	t.Run("Check with valid provider", func(t *testing.T) {
		ia := &InsightsAccount{}

		validProviders := []string{"aws", "azure", "gcp"}
		for _, provider := range validProviders {
			t.Run(provider, func(t *testing.T) {
				inputs := property.NewMap(map[string]property.Value{
					"organizationName": property.New("test-org"),
					"accountName":      property.New("test-account"),
					"provider":         property.New(provider),
					"environment":      property.New("test-env"),
					"scanSchedule":     property.New("daily"),
				})

				req := infer.CheckRequest{
					NewInputs: inputs,
				}

				resp, err := ia.Check(context.Background(), req)

				assert.NoError(t, err)
				assert.Empty(t, resp.Failures)
				assert.Equal(t, provider, resp.Inputs.Provider)
			})
		}
	})

	t.Run("Check with invalid provider", func(t *testing.T) {
		ia := &InsightsAccount{}

		inputs := property.NewMap(map[string]property.Value{
			"organizationName": property.New("test-org"),
			"accountName":      property.New("test-account"),
			"provider":         property.New("invalid-provider"),
			"environment":      property.New("test-env"),
			"scanSchedule":     property.New("daily"),
		})

		req := infer.CheckRequest{
			NewInputs: inputs,
		}

		resp, err := ia.Check(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Failures, 1)
		assert.Equal(t, "provider", resp.Failures[0].Property)
		assert.Contains(t, resp.Failures[0].Reason, "provider must be one of")
		assert.Contains(t, resp.Failures[0].Reason, "invalid-provider")
	})
}
