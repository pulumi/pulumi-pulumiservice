package provider

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type PulumiServiceTtlScheduleResource struct {
	client pulumiapi.ScheduleClient
}

type PulumiServiceTtlScheduleInput struct {
	Stack              pulumiapi.StackName
	Timestamp          *time.Time `pulumi:"timestamp"`
	DeleteAfterDestroy bool       `pulumi:"deleteAfterDestroy"`
}

type PulumiServiceTtlScheduleOutput struct {
	Input      PulumiServiceTtlScheduleInput
	ScheduleID string `pulumi:"scheduleID"`
}

func (i *PulumiServiceTtlScheduleInput) ToPropertyMap() resource.PropertyMap {
	propertyMap := StackToPropertyMap(i.Stack)

	if i.Timestamp != nil {
		propertyMap["timestamp"] = resource.NewPropertyValue(i.Timestamp.Format(time.RFC3339))
	}
	propertyMap["deleteAfterDestroy"] = resource.NewPropertyValue(i.DeleteAfterDestroy)

	return propertyMap
}

func ToPulumiServiceTtlScheduleInput(properties *structpb.Struct) (*PulumiServiceTtlScheduleInput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := PulumiServiceTtlScheduleInput{}
	stack, err := ParseStack(inputMap)
	if err != nil {
		return nil, err
	}
	input.Stack = *stack

	if inputMap["timestamp"].HasValue() && inputMap["timestamp"].IsString() {
		timestamp, err := time.Parse(time.RFC3339, inputMap["timestamp"].StringValue())
		if err != nil {
			return nil, err
		}
		input.Timestamp = &timestamp
	}

	if inputMap["deleteAfterDestroy"].HasValue() && inputMap["deleteAfterDestroy"].IsBool() {
		input.DeleteAfterDestroy = inputMap["deleteAfterDestroy"].BoolValue()
	}

	return &input, nil
}

func (st *PulumiServiceTtlScheduleResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return ScheduleSharedDiff(req)
}

func (st *PulumiServiceTtlScheduleResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return ScheduleSharedDelete(req, st.client)
}

func (st *PulumiServiceTtlScheduleResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	input, err := ToPulumiServiceTtlScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleReq := pulumiapi.CreateTtlScheduleRequest{
		Timestamp:          input.Timestamp,
		DeleteAfterDestroy: input.DeleteAfterDestroy,
	}
	scheduleID, err := st.client.CreateTtlSchedule(context.Background(), input.Stack, scheduleReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(*scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.Stack.OrgName, input.Stack.ProjectName, input.Stack.StackName, "ttl", *scheduleID),
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceTtlScheduleResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "stack", "timestamp"} {
		if !inputMap[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	if inputMap["deleteAfterDestroy"].HasValue() && !inputMap["deleteAfterDestroy"].IsBool() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "deleteAfterDestroy property is present but can't be parsed as bool",
			Property: "deleteAfterDestroy",
		})
	}

	return &pulumirpc.CheckResponse{Inputs: req.GetNews(), Failures: failures}, nil
}

func (st *PulumiServiceTtlScheduleResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	previousOutput, err := ToPulumiServiceSharedScheduleOutput(req.GetOlds())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceTtlScheduleInput(req.GetNews())
	if err != nil {
		return nil, err
	}

	updateReq := pulumiapi.CreateTtlScheduleRequest{
		Timestamp:          input.Timestamp,
		DeleteAfterDestroy: input.DeleteAfterDestroy,
	}
	scheduleID, err := st.client.UpdateTtlSchedule(context.Background(), input.Stack, updateReq, previousOutput.ScheduleID)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(*scheduleID, input.ToPropertyMap()),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceTtlScheduleResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	output, err := ToPulumiServiceSharedScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceTtlScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleID, err := st.client.GetSchedule(context.Background(), output.Stack, output.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to read TtlSchedule (%q): %w", req.Id, err)
	}
	if scheduleID == nil {
		// if the tag doesn't exist, then return empty response
		return &pulumirpc.ReadResponse{}, nil
	}

	outputProperties, err := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read TtlSchedule (%q): %w", req.Id, err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputProperties,
		Inputs:     outputProperties,
	}, nil
}

func (st *PulumiServiceTtlScheduleResource) Name() string {
	return "pulumiservice:index:TtlSchedule"
}

func (st *PulumiServiceTtlScheduleResource) Configure(_ PulumiServiceConfig) {
}
