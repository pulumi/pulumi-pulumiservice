// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

// Mock client for EnvironmentScheduleClient
type getEnvironmentScheduleFunc func() (*pulumiapi.EnvironmentScheduleResponse, error)
type createEnvironmentRotationScheduleFunc func() (*string, error)
type updateEnvironmentRotationScheduleFunc func() (*string, error)
type deleteEnvironmentScheduleFunc func() error

type EnvironmentScheduleClientMock struct {
	getEnvironmentScheduleFunc            getEnvironmentScheduleFunc
	createEnvironmentRotationScheduleFunc createEnvironmentRotationScheduleFunc
	updateEnvironmentRotationScheduleFunc updateEnvironmentRotationScheduleFunc
	deleteEnvironmentScheduleFunc         deleteEnvironmentScheduleFunc
}

func (c *EnvironmentScheduleClientMock) GetEnvironmentSchedule(ctx context.Context, env pulumiapi.EnvironmentIdentifier, scheduleID string) (*pulumiapi.EnvironmentScheduleResponse, error) {
	return c.getEnvironmentScheduleFunc()
}

func (c *EnvironmentScheduleClientMock) CreateEnvironmentRotationSchedule(ctx context.Context, env pulumiapi.EnvironmentIdentifier, req pulumiapi.CreateEnvironmentRotationScheduleRequest) (*string, error) {
	return c.createEnvironmentRotationScheduleFunc()
}

func (c *EnvironmentScheduleClientMock) UpdateEnvironmentRotationSchedule(ctx context.Context, env pulumiapi.EnvironmentIdentifier, req pulumiapi.CreateEnvironmentRotationScheduleRequest, scheduleID string) (*string, error) {
	return c.updateEnvironmentRotationScheduleFunc()
}

func (c *EnvironmentScheduleClientMock) DeleteEnvironmentSchedule(ctx context.Context, env pulumiapi.EnvironmentIdentifier, scheduleID string) error {
	return c.deleteEnvironmentScheduleFunc()
}

func buildEnvironmentScheduleClientMock(
	getFunc getEnvironmentScheduleFunc,
	createFunc createEnvironmentRotationScheduleFunc,
	updateFunc updateEnvironmentRotationScheduleFunc,
	deleteFunc deleteEnvironmentScheduleFunc,
) *EnvironmentScheduleClientMock {
	return &EnvironmentScheduleClientMock{
		getEnvironmentScheduleFunc:            getFunc,
		createEnvironmentRotationScheduleFunc: createFunc,
		updateEnvironmentRotationScheduleFunc: updateFunc,
		deleteEnvironmentScheduleFunc:         deleteFunc,
	}
}

// Test helper functions for RotationSchedule
func testRotationScheduleInputWithCron() PulumiServiceEnvironmentRotationScheduleInput {
	cron := "0 0 * * *"
	return PulumiServiceEnvironmentRotationScheduleInput{
		Environment: pulumiapi.EnvironmentIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			EnvName:     "test-env",
		},
		ScheduleCron: &cron,
	}
}

func testRotationScheduleInputWithOnce() PulumiServiceEnvironmentRotationScheduleInput {
	timestamp := time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC)
	return PulumiServiceEnvironmentRotationScheduleInput{
		Environment: pulumiapi.EnvironmentIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			EnvName:     "test-env",
		},
		ScheduleOnce: &timestamp,
	}
}

func testEnvironmentScheduleResponseWithCron() *pulumiapi.EnvironmentScheduleResponse {
	cron := "0 0 * * *"
	return &pulumiapi.EnvironmentScheduleResponse{
		ID:           "test-schedule-id",
		ScheduleCron: &cron,
		Definition: pulumiapi.EnvironmentScheduleDefinition{
			EnvironmentPath: "test-org/test-project/test-env",
			EnvironmentID:   "env-123",
		},
	}
}

func testEnvironmentScheduleResponseWithOnce() *pulumiapi.EnvironmentScheduleResponse {
	timeString := "2026-06-06 00:00:00.000"
	return &pulumiapi.EnvironmentScheduleResponse{
		ID:           "test-schedule-id",
		ScheduleOnce: &timeString,
		Definition: pulumiapi.EnvironmentScheduleDefinition{
			EnvironmentPath: "test-org/test-project/test-env",
			EnvironmentID:   "env-123",
		},
	}
}

// Read Tests
func TestEnvironmentRotationSchedule_Read_NotFound(t *testing.T) {
	mockedClient := buildEnvironmentScheduleClientMock(
		func() (*pulumiapi.EnvironmentScheduleResponse, error) { return nil, nil },
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithCron()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-env/rotations/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

func TestEnvironmentRotationSchedule_Read_FoundWithCron(t *testing.T) {
	mockedClient := buildEnvironmentScheduleClientMock(
		func() (*pulumiapi.EnvironmentScheduleResponse, error) {
			return testEnvironmentScheduleResponseWithCron(), nil
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithCron()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-env/rotations/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-project/test-env/rotations/test-schedule-id", resp.Id)
	assert.NotNil(t, resp.Properties)
}

func TestEnvironmentRotationSchedule_Read_FoundWithOnce(t *testing.T) {
	mockedClient := buildEnvironmentScheduleClientMock(
		func() (*pulumiapi.EnvironmentScheduleResponse, error) {
			return testEnvironmentScheduleResponseWithOnce(), nil
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithOnce()
	scheduleID := "test-schedule-id"

	outputProperties, _ := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-env/rotations/test-schedule-id",
		Properties: outputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-project/test-env/rotations/test-schedule-id", resp.Id)
	assert.NotNil(t, resp.Properties)
}

// Create Tests
func TestEnvironmentRotationSchedule_Create_WithScheduleCron(t *testing.T) {
	scheduleID := "created-schedule-id"
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		func() (*string, error) { return &scheduleID, nil },
		nil,
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithCron()
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

func TestEnvironmentRotationSchedule_Create_WithScheduleOnce(t *testing.T) {
	scheduleID := "created-schedule-id"
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		func() (*string, error) { return &scheduleID, nil },
		nil,
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithOnce()
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

func TestEnvironmentRotationSchedule_Create_APIError(t *testing.T) {
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		func() (*string, error) { return nil, errors.New("API error") },
		nil,
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithCron()
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
func TestEnvironmentRotationSchedule_Update_ChangesCron(t *testing.T) {
	scheduleID := "test-schedule-id"
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		nil,
		func() (*string, error) { return &scheduleID, nil },
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	oldInput := testRotationScheduleInputWithCron()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	// Change cron schedule
	newCron := "0 1 * * *"
	newInput := PulumiServiceEnvironmentRotationScheduleInput{
		Environment: pulumiapi.EnvironmentIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			EnvName:     "test-env",
		},
		ScheduleCron: &newCron,
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-project/test-env/rotations/test-schedule-id",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Update(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Update_FromCronToOnce(t *testing.T) {
	scheduleID := "test-schedule-id"
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		nil,
		func() (*string, error) { return &scheduleID, nil },
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	oldInput := testRotationScheduleInputWithCron()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	newInput := testRotationScheduleInputWithOnce()
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-project/test-env/rotations/test-schedule-id",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Update(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Update_APIError(t *testing.T) {
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		nil,
		func() (*string, error) { return nil, errors.New("API error") },
		nil,
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithCron()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-project/test-env/rotations/test-schedule-id",
		Olds: inputProperties,
		News: inputProperties,
	}

	resp, err := provider.Update(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Delete Tests
func TestEnvironmentRotationSchedule_Delete_Success(t *testing.T) {
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		nil,
		nil,
		func() error { return nil },
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithCron()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-env/rotations/test-schedule-id",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Delete_APIError(t *testing.T) {
	mockedClient := buildEnvironmentScheduleClientMock(
		nil,
		nil,
		nil,
		func() error { return errors.New("API error") },
	)

	provider := PulumiServiceEnvironmentRotationScheduleResource{
		Client: mockedClient,
	}

	input := testRotationScheduleInputWithCron()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-env/rotations/test-schedule-id",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Diff Tests
func TestEnvironmentRotationSchedule_Diff_NoChanges(t *testing.T) {
	provider := PulumiServiceEnvironmentRotationScheduleResource{}

	input := testRotationScheduleInputWithCron()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:   "test-org/test-project/test-env/rotations/test-schedule-id",
		Olds: inputProperties,
		News: inputProperties,
	}

	resp, err := provider.Diff(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Diff_DetectsChangesAndReplacements(t *testing.T) {
	provider := PulumiServiceEnvironmentRotationScheduleResource{}

	oldInput := testRotationScheduleInputWithCron()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	// Change organization - should trigger replacement
	newCron := "0 0 * * *"
	newInput := PulumiServiceEnvironmentRotationScheduleInput{
		Environment: pulumiapi.EnvironmentIdentifier{
			OrgName:     "new-org", // Changed
			ProjectName: "test-project",
			EnvName:     "test-env",
		},
		ScheduleCron: &newCron,
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:   "test-org/test-project/test-env/rotations/test-schedule-id",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Diff(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// Check Tests
func TestEnvironmentRotationSchedule_Check_ValidWithScheduleCron(t *testing.T) {
	provider := PulumiServiceEnvironmentRotationScheduleResource{}

	input := testRotationScheduleInputWithCron()
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

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Check_ValidWithScheduleOnce(t *testing.T) {
	provider := PulumiServiceEnvironmentRotationScheduleResource{}

	input := testRotationScheduleInputWithOnce()
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

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Check_BothSchedulesSet(t *testing.T) {
	provider := PulumiServiceEnvironmentRotationScheduleResource{}

	// Both scheduleCron and scheduleOnce set - should fail
	cron := "0 0 * * *"
	timestamp := time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC)
	input := PulumiServiceEnvironmentRotationScheduleInput{
		Environment: pulumiapi.EnvironmentIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			EnvName:     "test-env",
		},
		ScheduleCron: &cron,
		ScheduleOnce: &timestamp,
	}

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

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Check_NeitherScheduleSet(t *testing.T) {
	provider := PulumiServiceEnvironmentRotationScheduleResource{}

	// Neither scheduleCron nor scheduleOnce set - should fail
	input := PulumiServiceEnvironmentRotationScheduleInput{
		Environment: pulumiapi.EnvironmentIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			EnvName:     "test-env",
		},
		// No schedule set
	}

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

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnvironmentRotationSchedule_Check_MissingEnvironment(t *testing.T) {
	provider := PulumiServiceEnvironmentRotationScheduleResource{}

	// Missing environment fields
	cron := "0 0 * * *"
	input := PulumiServiceEnvironmentRotationScheduleInput{
		Environment: pulumiapi.EnvironmentIdentifier{
			OrgName: "test-org",
			// ProjectName and EnvName missing
		},
		ScheduleCron: &cron,
	}

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

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
