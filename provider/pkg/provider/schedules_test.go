package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type getDeploymentScheduleFunc func() (*string, error)

type ScheduleClientMock struct {
	getDeploymentScheduleFunc getDeploymentScheduleFunc
}

func (c *ScheduleClientMock) GetSchedule(ctx context.Context, stack pulumiapi.StackName, scheduleID string) (*string, error) {
	return c.getDeploymentScheduleFunc()
}

func (c *ScheduleClientMock) CreateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackName, req pulumiapi.CreateDeploymentScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) CreateDriftSchedule(ctx context.Context, stack pulumiapi.StackName, req pulumiapi.CreateDriftScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) CreateTtlSchedule(ctx context.Context, stack pulumiapi.StackName, req pulumiapi.CreateTtlScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackName, req pulumiapi.CreateDeploymentScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateDriftSchedule(ctx context.Context, stack pulumiapi.StackName, req pulumiapi.CreateDriftScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateTtlSchedule(ctx context.Context, stack pulumiapi.StackName, req pulumiapi.CreateTtlScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) DeleteSchedule(ctx context.Context, stack pulumiapi.StackName, scheduleID string) error {
	return nil
}

func buildScheduleClientMock(getDeploymentScheduleFunc getDeploymentScheduleFunc) *ScheduleClientMock {
	return &ScheduleClientMock{
		getDeploymentScheduleFunc,
	}
}

func TestDeploymentSchedule(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildScheduleClientMock(
			func() (*string, error) { return nil, nil },
		)

		provider := PulumiServiceDeploymentScheduleResource{}

		input := PulumiServiceDeploymentScheduleInput{
			Stack: pulumiapi.StackName{
				OrgName:     "org",
				ProjectName: "project",
				StackName:   "stack",
			},
			ScheduleCron:    nil,
			ScheduleOnce:    nil,
			PulumiOperation: "update",
		}
		scheduleID := "fake-schedule-id"

		outputProperties, _ := plugin.MarshalProperties(
			AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.ReadRequest{
			Id:         "org/project/stack/fake-schedule-id",
			Properties: outputProperties,
		}

		resp, err := provider.Read(WithClient(mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildScheduleClientMock(
			func() (*string, error) {
				scheduleID := "fake-schedule-id"
				return &scheduleID, nil
			},
		)

		provider := PulumiServiceDeploymentScheduleResource{}

		input := PulumiServiceDeploymentScheduleInput{
			Stack: pulumiapi.StackName{
				OrgName:     "org",
				ProjectName: "project",
				StackName:   "stack",
			},
			ScheduleCron:    nil,
			ScheduleOnce:    nil,
			PulumiOperation: "update",
		}
		scheduleID := "fake-schedule-id"

		outputProperties, _ := plugin.MarshalProperties(
			AddScheduleIdToPropertyMap(scheduleID, input.ToPropertyMap()),
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.ReadRequest{
			Id:         "org/project/stack/fake-schedule-id",
			Properties: outputProperties,
		}

		resp, err := provider.Read(WithClient(mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "org/project/stack/fake-schedule-id")
	})
}
