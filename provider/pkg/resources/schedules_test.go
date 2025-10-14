package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type getDeploymentScheduleFunc func() (*pulumiapi.StackScheduleResponse, error)

type ScheduleClientMock struct {
	getDeploymentScheduleFunc getDeploymentScheduleFunc
}

func (c *ScheduleClientMock) GetStackSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* scheduleID */ string,
) (*pulumiapi.StackScheduleResponse, error) {
	return c.getDeploymentScheduleFunc()
}

func (c *ScheduleClientMock) CreateDeploymentSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* req */ pulumiapi.CreateDeploymentScheduleRequest,
) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) CreateDriftSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* req */ pulumiapi.CreateDriftScheduleRequest,
) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) CreateTTLSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* req */ pulumiapi.CreateTTLScheduleRequest,
) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateDeploymentSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* req */ pulumiapi.CreateDeploymentScheduleRequest,
	_ /* scheduleID */ string,
) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateDriftSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* req */ pulumiapi.CreateDriftScheduleRequest,
	_ /* scheduleID */ string,
) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) UpdateTTLSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* req */ pulumiapi.CreateTTLScheduleRequest,
	_ /* scheduleID */ string,
) (*string, error) {
	return nil, nil
}

func (c *ScheduleClientMock) DeleteStackSchedule(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* scheduleID */ string,
) error {
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
			func() (*pulumiapi.StackScheduleResponse, error) { return nil, nil },
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
			AddScheduleIDToPropertyMap(scheduleID, input.ToPropertyMap()),
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
			func() (*pulumiapi.StackScheduleResponse, error) {
				timeString := "2026-06-06 00:00:00.000"
				return &pulumiapi.StackScheduleResponse{
					ID:           "fake-id",
					ScheduleOnce: &timeString,
					ScheduleCron: nil,
					Definition: pulumiapi.StackScheduleDefinition{
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
			AddScheduleIDToPropertyMap(scheduleID, input.ToPropertyMap()),
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
