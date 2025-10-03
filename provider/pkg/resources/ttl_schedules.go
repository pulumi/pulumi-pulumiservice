package resources

import (
	"context"
	"fmt"
	"path"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type PulumiServiceTtlScheduleResource struct {
	Client pulumiapi.StackScheduleClient
}

type PulumiServiceTtlScheduleInput struct {
	Stack              pulumiapi.StackIdentifier
	Timestamp          time.Time `pulumi:"timestamp"`
	DeleteAfterDestroy bool      `pulumi:"deleteAfterDestroy"`
}

type PulumiServiceTtlScheduleOutput struct {
	Input      PulumiServiceTtlScheduleInput
	ScheduleID string `pulumi:"scheduleId"`
}

func (i *PulumiServiceTtlScheduleInput) ToPropertyMap() resource.PropertyMap {
	propertyMap := StackToPropertyMap(i.Stack)

	propertyMap["timestamp"] = resource.NewPropertyValue(i.Timestamp.Format(time.RFC3339))
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
		input.Timestamp = timestamp
	}

	if inputMap["deleteAfterDestroy"].HasValue() && inputMap["deleteAfterDestroy"].IsBool() {
		input.DeleteAfterDestroy = inputMap["deleteAfterDestroy"].BoolValue()
	}

	return &input, nil
}

func (st *PulumiServiceTtlScheduleResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	if !news["deleteAfterDestroy"].HasValue() {
		news["deleteAfterDestroy"] = resource.NewBoolProperty(false)
	}

	return StackScheduleSharedDiffMaps(olds, news)
}

func (st *PulumiServiceTtlScheduleResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return StackScheduleSharedDelete(req, st.Client)
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
	scheduleID, err := st.Client.CreateTtlSchedule(context.Background(), input.Stack, scheduleReq)
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
	inputMap, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
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
	previousOutput, err := ToPulumiServiceStackScheduleOutput(req.GetOlds())
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
	scheduleID, err := st.Client.UpdateTtlSchedule(
		context.Background(),
		input.Stack,
		updateReq,
		previousOutput.ScheduleID,
	)
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
	stack, scheduleID, err := ParseStackScheduleID(req.Id, "ttl")
	if err != nil {
		return nil, err
	}

	scheduleResponse, err := st.Client.GetStackSchedule(context.Background(), *stack, *scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to read TtlSchedule (%q): %w", req.Id, err)
	}
	if scheduleResponse == nil {
		// if schedule doesn't exist, then return empty response to delete it from state
		return &pulumirpc.ReadResponse{}, nil
	}

	timestamp, err := time.Parse(time.DateTime, *scheduleResponse.ScheduleOnce)
	if err != nil {
		return nil, fmt.Errorf("failed to read DeploymentSchedule (%q): %w", req.Id, err)
	}
	input := PulumiServiceTtlScheduleInput{
		Stack:              *stack,
		Timestamp:          timestamp,
		DeleteAfterDestroy: scheduleResponse.Definition.Request.OperationContext.Options.DeleteAfterDestroy,
	}

	inputs, err := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read TtlSchedule (%q): %w", req.Id, err)
	}
	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(*scheduleID, input.ToPropertyMap()),
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
		Inputs:     inputs,
	}, nil
}

func (st *PulumiServiceTtlScheduleResource) Name() string {
	return "pulumiservice:index:TtlSchedule"
}
