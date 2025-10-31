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
)

// Mock types for DriftSchedule tests
type getDriftScheduleFunc func() (*pulumiapi.StackScheduleResponse, error)
type createDriftScheduleFunc func() (*string, error)
type updateDriftScheduleFunc func() (*string, error)
type deleteDriftScheduleFunc func() error

type driftScheduleClientMock struct {
	getFunc    getDriftScheduleFunc
	createFunc createDriftScheduleFunc
	updateFunc updateDriftScheduleFunc
	deleteFunc deleteDriftScheduleFunc
}

func (c *driftScheduleClientMock) GetStackSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, scheduleID string) (*pulumiapi.StackScheduleResponse, error) {
	if c.getFunc == nil {
		return nil, nil
	}
	return c.getFunc()
}

func (c *driftScheduleClientMock) CreateDriftSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDriftScheduleRequest) (*string, error) {
	if c.createFunc == nil {
		return nil, nil
	}
	return c.createFunc()
}

func (c *driftScheduleClientMock) UpdateDriftSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDriftScheduleRequest, scheduleID string) (*string, error) {
	if c.updateFunc == nil {
		return nil, nil
	}
	return c.updateFunc()
}

func (c *driftScheduleClientMock) DeleteStackSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, scheduleID string) error {
	if c.deleteFunc == nil {
		return nil
	}
	return c.deleteFunc()
}

// Implement remaining interface methods (not used in these tests but required by interface)
func (c *driftScheduleClientMock) CreateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDeploymentScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *driftScheduleClientMock) CreateTtlSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateTtlScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *driftScheduleClientMock) UpdateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDeploymentScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *driftScheduleClientMock) UpdateTtlSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateTtlScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func buildDriftScheduleClientMock(
	getFunc getDriftScheduleFunc,
	createFunc createDriftScheduleFunc,
	updateFunc updateDriftScheduleFunc,
	deleteFunc deleteDriftScheduleFunc,
) *driftScheduleClientMock {
	return &driftScheduleClientMock{
		getFunc:    getFunc,
		createFunc: createFunc,
		updateFunc: updateFunc,
		deleteFunc: deleteFunc,
	}
}

// Test helper functions for DriftSchedule
func testDriftScheduleInput() PulumiServiceDriftScheduleInput {
	return PulumiServiceDriftScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		ScheduleCron:  "0 0 * * *",
		AutoRemediate: false,
	}
}

func testDriftScheduleResponse() *pulumiapi.StackScheduleResponse {
	cron := "0 0 * * *"
	return &pulumiapi.StackScheduleResponse{
		ID:           "test-schedule-id",
		ScheduleCron: &cron,
		Definition: pulumiapi.StackScheduleDefinition{
			Request: pulumiapi.CreateDeploymentRequest{
				OperationContext: pulumiapi.ScheduleOperationContext{
					Options: pulumiapi.ScheduleOperationContextOptions{
						AutoRemediate: false,
					},
				},
			},
		},
	}
}

// Read Tests
func TestDriftSchedule_Read_NotFound(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		func() (*pulumiapi.StackScheduleResponse, error) { return nil, nil },
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-stack/drift/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

func TestDriftSchedule_Read_Found(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		func() (*pulumiapi.StackScheduleResponse, error) {
			return testDriftScheduleResponse(), nil
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-stack/drift/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-project/test-stack/drift/test-schedule-id", resp.Id)
	assert.NotNil(t, resp.Properties)
}

func TestDriftSchedule_Read_APIError(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		func() (*pulumiapi.StackScheduleResponse, error) {
			return nil, errors.New("API error")
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-stack/drift/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Create Tests
func TestDriftSchedule_Create_Success(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		nil,
		func() (*string, error) {
			id := "test-schedule-id"
			return &id, nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, "", resp.Id)
}

func TestDriftSchedule_Create_WithDefaultAutoRemediate(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		nil,
		func() (*string, error) {
			id := "test-schedule-id"
			return &id, nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	// AutoRemediate not set - should default to false
	input := PulumiServiceDriftScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		ScheduleCron: "0 0 * * *",
		// AutoRemediate not set
	}

	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDriftSchedule_Create_APIError(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		nil,
		func() (*string, error) {
			return nil, errors.New("API error")
		},
		nil,
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Update Tests
func TestDriftSchedule_Update_Success(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		nil,
		nil,
		func() (*string, error) {
			id := "test-schedule-id"
			return &id, nil
		},
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	// Update to new cron schedule
	newInput := PulumiServiceDriftScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		ScheduleCron:  "0 1 * * *", // Changed from "0 0 * * *"
		AutoRemediate: true,
	}

	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	oldProperties, _ := plugin.MarshalProperties(
		func() resource.PropertyMap { i := testDriftScheduleInput(); return i.ToPropertyMap() }(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:    "test-org/test-project/test-stack/drift/test-schedule-id",
		Olds:  oldProperties,
		News:  newProperties,
	}

	resp, err := provider.Update(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDriftSchedule_Update_APIError(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		nil,
		nil,
		func() (*string, error) {
			return nil, errors.New("API error")
		},
		nil,
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:    "test-org/test-project/test-stack/drift/test-schedule-id",
		Olds:  inputProperties,
		News:  inputProperties,
	}

	resp, err := provider.Update(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Delete Tests
func TestDriftSchedule_Delete_Success(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		nil,
		nil,
		nil,
		func() error { return nil },
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-stack/drift/test-schedule-id",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDriftSchedule_Delete_APIError(t *testing.T) {
	mockedClient := buildDriftScheduleClientMock(
		nil,
		nil,
		nil,
		func() error { return errors.New("API error") },
	)

	provider := PulumiServiceDriftScheduleResource{
		Client: mockedClient,
	}

	input := testDriftScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-stack/drift/test-schedule-id",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Diff Tests
func TestDriftSchedule_Diff_NoChanges(t *testing.T) {
	provider := PulumiServiceDriftScheduleResource{}

	input := testDriftScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:        "test-org/test-project/test-stack/drift/test-schedule-id",
		OldInputs: inputProperties,
		News:      inputProperties,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
}

func TestDriftSchedule_Diff_DetectsChanges(t *testing.T) {
	provider := PulumiServiceDriftScheduleResource{}

	oldInput := testDriftScheduleInput()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	// Change scheduleCron
	newInput := PulumiServiceDriftScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		ScheduleCron:  "0 1 * * *", // Changed
		AutoRemediate: true,         // Changed
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:        "test-org/test-project/test-stack/drift/test-schedule-id",
		OldInputs: oldProperties,
		News:      newProperties,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
}

func TestDriftSchedule_Diff_DefaultAutoRemediate(t *testing.T) {
	provider := PulumiServiceDriftScheduleResource{}

	// Old input with autoRemediate not set (should default to false)
	oldInput := PulumiServiceDriftScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		ScheduleCron: "0 0 * * *",
		// AutoRemediate not set
	}
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	// New input explicitly sets false - should be no change
	newInput := PulumiServiceDriftScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		ScheduleCron:  "0 0 * * *",
		AutoRemediate: false,
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:        "test-org/test-project/test-stack/drift/test-schedule-id",
		OldInputs: oldProperties,
		News:      newProperties,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
}

// Check Tests
func TestDriftSchedule_Check_ValidInputs(t *testing.T) {
	provider := PulumiServiceDriftScheduleResource{}

	input := testDriftScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CheckRequest{
		News: inputProperties,
	}

	resp, err := provider.Check(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}

func TestDriftSchedule_Check_MissingScheduleCron(t *testing.T) {
	provider := PulumiServiceDriftScheduleResource{}

	// Missing organization and scheduleCron - should fail validation
	propertyMap := resource.PropertyMap{
		"project":       resource.NewPropertyValue("test-project"),
		"stack":         resource.NewPropertyValue("test-stack"),
		"autoRemediate": resource.NewPropertyValue(false),
	}

	inputProperties, _ := plugin.MarshalProperties(
		propertyMap,
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CheckRequest{
		News: inputProperties,
	}

	resp, err := provider.Check(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Failures)
	// Should have at least 2 failures (organization and scheduleCron)
	assert.True(t, len(resp.Failures) >= 2)
}
