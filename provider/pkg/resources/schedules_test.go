package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type getDeploymentScheduleFunc func() (*pulumiapi.ScheduleResponse, error)

type ScheduleClientMock struct {
	getDeploymentScheduleFunc getDeploymentScheduleFunc
}

func (c *ScheduleClientMock) GetSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, scheduleID string) (*pulumiapi.ScheduleResponse, error) {
	return c.getDeploymentScheduleFunc()
}

func (c *ScheduleClientMock) CreateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDeploymentScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) CreateDriftSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDriftScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) CreateTtlSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateTtlScheduleRequest) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateDeploymentSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDeploymentScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateDriftSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateDriftScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateTtlSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, req pulumiapi.CreateTtlScheduleRequest, scheduleID string) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) DeleteSchedule(ctx context.Context, stack pulumiapi.StackIdentifier, scheduleID string) error {
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
			func() (*pulumiapi.ScheduleResponse, error) { return nil, nil },
		)

		provider := PulumiServiceDeploymentScheduleResource{
			Client: mockedClient,
		}

		input := PulumiServiceDeploymentScheduleInput{
			Stack: pulumiapi.StackIdentifier{
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

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildScheduleClientMock(
			func() (*pulumiapi.ScheduleResponse, error) {
				timeString := "2026-06-06 00:00:00.000"
				return &pulumiapi.ScheduleResponse{
					ID:           "fake-id",
					ScheduleOnce: &timeString,
					ScheduleCron: nil,
					Definition: pulumiapi.ScheduleDefinition{
						Request: pulumiapi.CreateDeploymentRequest{
							PulumiOperation: "update",
							OperationContext: pulumiapi.ScheduleOperationContext{
								Options: pulumiapi.ScheduleOperationContextOptions{
									AutoRemediate:      true,
									DeleteAfterDestroy: false,
								},
							},
						},
					},
				}, nil
			},
		)

		provider := PulumiServiceDeploymentScheduleResource{
			Client: mockedClient,
		}

		input := PulumiServiceDeploymentScheduleInput{
			Stack: pulumiapi.StackIdentifier{
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

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "org/project/stack/fake-schedule-id")
	})
}
