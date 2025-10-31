// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

// Mock types for TtlSchedule tests
type getTtlScheduleFunc func() (*pulumiapi.StackScheduleResponse, error)
type createTtlScheduleFunc func() (*string, error)
type updateTtlScheduleFunc func() (*string, error)
type deleteTtlScheduleFunc func() error

type ttlScheduleClientMock struct {
	getFunc    getTtlScheduleFunc
	createFunc createTtlScheduleFunc
	updateFunc updateTtlScheduleFunc
	deleteFunc deleteTtlScheduleFunc
}

func (c *ttlScheduleClientMock) GetStackSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, scheduleID string) (*pulumiapi.StackScheduleResponse, error) {
	if c.getFunc == nil {
		return nil, nil
	}
	return c.getFunc()
}

func (c *ttlScheduleClientMock) CreateTtlSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateTtlScheduleRequest) (*string, error) {
	if c.createFunc == nil {
		return nil, nil
	}
	return c.createFunc()
}

func (c *ttlScheduleClientMock) UpdateTtlSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateTtlScheduleRequest, scheduleID string) (*string, error) {
	if c.updateFunc == nil {
		return nil, nil
	}
	return c.updateFunc()
}

func (c *ttlScheduleClientMock) DeleteStackSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, scheduleID string) error {
	if c.deleteFunc == nil {
		return nil
	}
	return c.deleteFunc()
}

// Implement remaining interface methods (not used in these tests but required by interface)
func (c *ttlScheduleClientMock) CreateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDeploymentScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ttlScheduleClientMock) CreateDriftSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDriftScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ttlScheduleClientMock) UpdateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDeploymentScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *ttlScheduleClientMock) UpdateDriftSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDriftScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func buildTtlScheduleClientMock(
	getFunc getTtlScheduleFunc,
	createFunc createTtlScheduleFunc,
	updateFunc updateTtlScheduleFunc,
	deleteFunc deleteTtlScheduleFunc,
) *ttlScheduleClientMock {
	return &ttlScheduleClientMock{
		getFunc:    getFunc,
		createFunc: createFunc,
		updateFunc: updateFunc,
		deleteFunc: deleteFunc,
	}
}

// Test helper functions
func testTtlScheduleInput() PulumiServiceTtlScheduleInput {
	timestamp := time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC)
	return PulumiServiceTtlScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		Timestamp:          timestamp,
		DeleteAfterDestroy: false,
	}
}

func testTtlStackScheduleResponse() *pulumiapi.StackScheduleResponse {
	timeString := "2026-06-06 00:00:00.000"
	return &pulumiapi.StackScheduleResponse{
		ID:           "test-schedule-id",
		ScheduleOnce: &timeString,
		Definition: pulumiapi.StackScheduleDefinition{
			Request: pulumiapi.CreateDeploymentRequest{
				OperationContext: pulumiapi.ScheduleOperationContext{
					Options: pulumiapi.ScheduleOperationContextOptions{
						DeleteAfterDestroy: false,
					},
				},
			},
		},
	}
}

// Read Tests
func TestTtlSchedule_Read_NotFound(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		func() (*pulumiapi.StackScheduleResponse, error) { return nil, nil },
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	input := testTtlScheduleInput()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-stack/ttl/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

func TestTtlSchedule_Read_Found(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		func() (*pulumiapi.StackScheduleResponse, error) {
			return testTtlStackScheduleResponse(), nil
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	input := testTtlScheduleInput()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-stack/ttl/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-project/test-stack/ttl/test-schedule-id", resp.Id)
	assert.NotNil(t, resp.Properties)
}

func TestTtlSchedule_Read_APIError(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		func() (*pulumiapi.StackScheduleResponse, error) {
			return nil, errors.New("API error")
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	input := testTtlScheduleInput()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-stack/ttl/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Create Tests
func TestTtlSchedule_Create_Success(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		nil,
		func() (*string, error) {
			id := "test-schedule-id"
			return &id, nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	input := testTtlScheduleInput()
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

func TestTtlSchedule_Create_WithDefaultDeleteAfterDestroy(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		nil,
		func() (*string, error) {
			id := "test-schedule-id"
			return &id, nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	timestamp := time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC)
	input := PulumiServiceTtlScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		Timestamp: timestamp,
		// DeleteAfterDestroy not set - should default to false
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

// Update Tests
func TestTtlSchedule_Update_Success(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		nil,
		nil,
		func() (*string, error) {
			id := "test-schedule-id"
			return &id, nil
		},
		nil,
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	// Update to new timestamp
	newTimestamp := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	input := PulumiServiceTtlScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		Timestamp:          newTimestamp,
		DeleteAfterDestroy: true,
	}

	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	oldProperties, _ := plugin.MarshalProperties(
		func() resource.PropertyMap { i := testTtlScheduleInput(); return i.ToPropertyMap() }(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:    "test-org/test-project/test-stack/ttl/test-schedule-id",
		Olds:  oldProperties,
		News:  inputProperties,
	}

	resp, err := provider.Update(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestTtlSchedule_Update_APIError(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		nil,
		nil,
		func() (*string, error) {
			return nil, errors.New("API error")
		},
		nil,
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	input := testTtlScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:    "test-org/test-project/test-stack/ttl/test-schedule-id",
		Olds:  inputProperties,
		News:  inputProperties,
	}

	resp, err := provider.Update(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Delete Tests
func TestTtlSchedule_Delete_Success(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		nil,
		nil,
		nil,
		func() error { return nil },
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	input := testTtlScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-stack/ttl/test-schedule-id",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestTtlSchedule_Delete_APIError(t *testing.T) {
	mockedClient := buildTtlScheduleClientMock(
		nil,
		nil,
		nil,
		func() error { return errors.New("API error") },
	)

	provider := PulumiServiceTtlScheduleResource{
		Client: mockedClient,
	}

	input := testTtlScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-stack/ttl/test-schedule-id",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Diff Tests
func TestTtlSchedule_Diff_NoChanges(t *testing.T) {
	provider := PulumiServiceTtlScheduleResource{}

	input := testTtlScheduleInput()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:        "test-org/test-project/test-stack/ttl/test-schedule-id",
		OldInputs: inputProperties,
		News:      inputProperties,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
}

func TestTtlSchedule_Diff_DetectsChanges(t *testing.T) {
	provider := PulumiServiceTtlScheduleResource{}

	oldInput := testTtlScheduleInput()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	// Change timestamp
	newTimestamp := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	newInput := PulumiServiceTtlScheduleInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		Timestamp:          newTimestamp,
		DeleteAfterDestroy: true,
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:        "test-org/test-project/test-stack/ttl/test-schedule-id",
		OldInputs: oldProperties,
		News:      newProperties,
	}

	resp, err := provider.Diff(&req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
}

// Check Tests
func TestTtlSchedule_Check_ValidInputs(t *testing.T) {
	provider := PulumiServiceTtlScheduleResource{}

	input := testTtlScheduleInput()
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

func TestTtlSchedule_Check_MissingTimestamp(t *testing.T) {
	provider := PulumiServiceTtlScheduleResource{}

	// Missing organization and timestamp - should fail validation
	propertyMap := resource.PropertyMap{
		"project":              resource.NewPropertyValue("test-project"),
		"stack":                resource.NewPropertyValue("test-stack"),
		"deleteAfterDestroy":   resource.NewPropertyValue(false),
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
	// Should have at least 2 failures (organization and timestamp)
	assert.True(t, len(resp.Failures) >= 2)
}
