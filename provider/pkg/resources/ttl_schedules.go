package resources

import (
	"context"
	"fmt"
	"path"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTTLScheduleResource struct {
	Client pulumiapi.StackScheduleClient
}

type PulumiServiceTTLScheduleInput struct {
	Stack              pulumiapi.StackIdentifier
	Timestamp          time.Time `pulumi:"timestamp"`
	DeleteAfterDestroy bool      `pulumi:"deleteAfterDestroy"`
}

type PulumiServiceTTLScheduleOutput struct {
	Input      PulumiServiceTTLScheduleInput
	ScheduleID string `pulumi:"scheduleId"`
}

func (i *PulumiServiceTTLScheduleInput) ToPropertyMap() resource.PropertyMap {
	propertyMap := StackToPropertyMap(i.Stack)

	propertyMap["timestamp"] = resource.NewPropertyValue(i.Timestamp.Format(time.RFC3339))
	propertyMap["deleteAfterDestroy"] = resource.NewPropertyValue(i.DeleteAfterDestroy)

	return propertyMap
}

func ToPulumiServiceTTLScheduleInput(properties *structpb.Struct) (*PulumiServiceTTLScheduleInput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := PulumiServiceTTLScheduleInput{}
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

func (st *PulumiServiceTTLScheduleResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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

func (st *PulumiServiceTTLScheduleResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return StackScheduleSharedDelete(req, st.Client)
}

func (st *PulumiServiceTTLScheduleResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	input, err := ToPulumiServiceTTLScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleReq := pulumiapi.CreateTTLScheduleRequest{
		Timestamp:          input.Timestamp,
		DeleteAfterDestroy: input.DeleteAfterDestroy,
	}
	scheduleID, err := st.Client.CreateTTLSchedule(context.Background(), input.Stack, scheduleReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIDToPropertyMap(*scheduleID, input.ToPropertyMap()),
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

func (st *PulumiServiceTTLScheduleResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
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

func (st *PulumiServiceTTLScheduleResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	previousOutput, err := ToPulumiServiceStackScheduleOutput(req.GetOlds())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceTTLScheduleInput(req.GetNews())
	if err != nil {
		return nil, err
	}

	updateReq := pulumiapi.CreateTTLScheduleRequest{
		Timestamp:          input.Timestamp,
		DeleteAfterDestroy: input.DeleteAfterDestroy,
	}
	scheduleID, err := st.Client.UpdateTTLSchedule(
		context.Background(),
		input.Stack,
		updateReq,
		previousOutput.ScheduleID,
	)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIDToPropertyMap(*scheduleID, input.ToPropertyMap()),
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

func (st *PulumiServiceTTLScheduleResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
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
	input := PulumiServiceTTLScheduleInput{
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
		AddScheduleIDToPropertyMap(*scheduleID, input.ToPropertyMap()),
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

func (st *PulumiServiceTTLScheduleResource) Name() string {
	return "pulumiservice:index:TtlSchedule"
}
