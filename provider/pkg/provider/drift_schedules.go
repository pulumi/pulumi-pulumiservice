package provider

import (
	"context"
	"fmt"
	"path"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type PulumiServiceDriftScheduleResource struct{}

type PulumiServiceDriftScheduleInput struct {
	Stack         pulumiapi.StackName
	ScheduleCron  string `pulumi:"scheduleCron"`
	AutoRemediate bool   `pulumi:"autoRemediate"`
}

type PulumiServiceDriftScheduleOutput struct {
	Input      PulumiServiceDriftScheduleInput
	ScheduleID string `pulumi:"scheduleId"`
}

func (*PulumiServiceDriftScheduleResource) client(ctx context.Context) pulumiapi.ScheduleClient {
	return GetClient[pulumiapi.ScheduleClient](ctx)
}

func (i *PulumiServiceDriftScheduleInput) ToPropertyMap() resource.PropertyMap {
	propertyMap := StackToPropertyMap(i.Stack)

	propertyMap["scheduleCron"] = resource.NewPropertyValue(i.ScheduleCron)
	propertyMap["autoRemediate"] = resource.NewPropertyValue(i.AutoRemediate)

	return propertyMap
}

func ToPulumiServiceDriftScheduleInput(properties *structpb.Struct) (*PulumiServiceDriftScheduleInput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := PulumiServiceDriftScheduleInput{}
	stack, err := ParseStack(inputMap)
	if err != nil {
		return nil, err
	}
	input.Stack = *stack

	if inputMap["scheduleCron"].HasValue() && inputMap["scheduleCron"].IsString() {
		scheduleCron := inputMap["scheduleCron"].StringValue()
		input.ScheduleCron = scheduleCron
	}

	if inputMap["autoRemediate"].HasValue() && inputMap["autoRemediate"].IsBool() {
		input.AutoRemediate = inputMap["autoRemediate"].BoolValue()
	}

	return &input, nil
}

func (st *PulumiServiceDriftScheduleResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return ScheduleSharedDiff(req)
}

func (st *PulumiServiceDriftScheduleResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return ScheduleSharedDelete(req, st.client(ctx))
}

func (st *PulumiServiceDriftScheduleResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	input, err := ToPulumiServiceDriftScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleReq := pulumiapi.CreateDriftScheduleRequest{
		ScheduleCron:  input.ScheduleCron,
		AutoRemediate: input.AutoRemediate,
	}
	scheduleID, err := st.client(ctx).CreateDriftSchedule(ctx, input.Stack, scheduleReq)
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
		Id:         path.Join(input.Stack.OrgName, input.Stack.ProjectName, input.Stack.StackName, "drift", *scheduleID),
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceDriftScheduleResource) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "stack", "scheduleCron"} {
		if !inputMap[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	if inputMap["autoRemediate"].HasValue() && !inputMap["autoRemediate"].IsBool() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "autoRemediate property is present but can't be parsed as bool",
			Property: "autoRemediate",
		})
	}

	return &pulumirpc.CheckResponse{Inputs: req.GetNews(), Failures: failures}, nil
}

func (st *PulumiServiceDriftScheduleResource) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	previousOutput, err := ToPulumiServiceSharedScheduleOutput(req.GetOlds())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceDriftScheduleInput(req.GetNews())
	if err != nil {
		return nil, err
	}

	updateReq := pulumiapi.CreateDriftScheduleRequest{
		ScheduleCron:  input.ScheduleCron,
		AutoRemediate: input.AutoRemediate,
	}
	scheduleID, err := st.client(ctx).UpdateDriftSchedule(ctx, input.Stack, updateReq, previousOutput.ScheduleID)
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

func (st *PulumiServiceDriftScheduleResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	output, err := ToPulumiServiceSharedScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceDriftScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleID, err := st.client(ctx).GetSchedule(ctx, output.Stack, output.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to read DriftSchedule (%q): %w", req.Id, err)
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
		return nil, fmt.Errorf("failed to read DriftSchedule (%q): %w", req.Id, err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputProperties,
		Inputs:     outputProperties,
	}, nil
}

func (st *PulumiServiceDriftScheduleResource) Name() string {
	return "pulumiservice:index:DriftSchedule"
}
