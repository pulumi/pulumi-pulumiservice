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

type PulumiServiceDeploymentScheduleResource struct{}

type PulumiServiceDeploymentScheduleInput struct {
	Stack           pulumiapi.StackName
	ScheduleCron    *string    `pulumi:"scheduleCron"`
	ScheduleOnce    *time.Time `pulumi:"scheduleOnce"`
	PulumiOperation string     `pulumi:"pulumiOperation"`
}

type PulumiServiceSharedScheduleOutput struct {
	Stack      pulumiapi.StackName
	ScheduleID string `pulumi:"scheduleId"`
}

func (*PulumiServiceDeploymentScheduleResource) client(ctx context.Context) pulumiapi.ScheduleClient {
	return GetClient[pulumiapi.ScheduleClient](ctx)
}

func StackToPropertyMap(stack pulumiapi.StackName) resource.PropertyMap {
	propertyMap := resource.PropertyMap{}
	propertyMap["organization"] = resource.NewPropertyValue(stack.OrgName)
	propertyMap["project"] = resource.NewPropertyValue(stack.ProjectName)
	propertyMap["stack"] = resource.NewPropertyValue(stack.StackName)
	return propertyMap
}

func (i *PulumiServiceDeploymentScheduleInput) ToPropertyMap() resource.PropertyMap {
	propertyMap := StackToPropertyMap(i.Stack)

	if i.ScheduleCron != nil {
		propertyMap["scheduleCron"] = resource.NewPropertyValue(i.ScheduleCron)
	}
	if i.ScheduleOnce != nil {
		propertyMap["timestamp"] = resource.NewPropertyValue(i.ScheduleOnce.Format(time.RFC3339))
	}
	propertyMap["pulumiOperation"] = resource.NewPropertyValue(i.PulumiOperation)

	return propertyMap
}

func AddScheduleIdToPropertyMap(scheduleID string, propertyMap resource.PropertyMap) resource.PropertyMap {
	propertyMap["scheduleId"] = resource.NewPropertyValue(scheduleID)
	return propertyMap
}

func ParseStack(inputMap resource.PropertyMap) (*pulumiapi.StackName, error) {
	var stack pulumiapi.StackName
	if inputMap["organization"].HasValue() && inputMap["organization"].IsString() {
		organization := inputMap["organization"].StringValue()
		stack.OrgName = organization
	} else {
		return nil, fmt.Errorf("failed to unmarshal organization value from properties: %s", inputMap)
	}
	if inputMap["project"].HasValue() && inputMap["project"].IsString() {
		project := inputMap["project"].StringValue()
		stack.ProjectName = project
	} else {
		return nil, fmt.Errorf("failed to unmarshal project value from properties: %s", inputMap)
	}
	if inputMap["stack"].HasValue() && inputMap["stack"].IsString() {
		stackName := inputMap["stack"].StringValue()
		stack.StackName = stackName
	} else {
		return nil, fmt.Errorf("failed to unmarshal stackName value from properties: %s", inputMap)
	}
	return &stack, nil
}

func ToPulumiServiceDeploymentScheduleInput(properties *structpb.Struct) (*PulumiServiceDeploymentScheduleInput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := PulumiServiceDeploymentScheduleInput{}
	stack, err := ParseStack(inputMap)
	if err != nil {
		return nil, err
	}
	input.Stack = *stack

	if inputMap["pulumiOperation"].HasValue() && inputMap["pulumiOperation"].IsString() {
		pulumiOperation := inputMap["pulumiOperation"].StringValue()
		input.PulumiOperation = pulumiOperation
	} else {
		return nil, fmt.Errorf("failed to unmarshal pulumiOperation value from properties: %s", inputMap)
	}

	if inputMap["scheduleCron"].HasValue() && inputMap["scheduleCron"].IsString() {
		scheduleCron := inputMap["scheduleCron"].StringValue()
		input.ScheduleCron = &scheduleCron
	}

	if inputMap["timestamp"].HasValue() && inputMap["timestamp"].IsString() {
		timestamp, err := time.Parse(time.RFC3339, inputMap["timestamp"].StringValue())
		if err != nil {
			return nil, err
		}
		input.ScheduleOnce = &timestamp
	}

	return &input, nil
}

func ToPulumiServiceSharedScheduleOutput(properties *structpb.Struct) (*PulumiServiceSharedScheduleOutput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	stack, err := ParseStack(inputMap)
	if err != nil {
		return nil, err
	}

	output := PulumiServiceSharedScheduleOutput{}
	output.Stack = *stack

	if inputMap["scheduleId"].HasValue() && inputMap["scheduleId"].IsString() {
		output.ScheduleID = inputMap["scheduleId"].StringValue()
	}

	return &output, nil
}

func ScheduleSharedDiff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	// preprocess olds to remove the `scheduleId` property since it's only an output and shouldn't cause a diff
	if olds["scheduleId"].HasValue() {
		delete(olds, "scheduleId")
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaces := []string(nil)
	replaceProperties := map[string]bool{
		"organization": true,
		"project":      true,
		"stack":        true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(detailedDiffs) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            replaces,
		DetailedDiff:        detailedDiffs,
		HasDetailedDiff:     true,
		DeleteBeforeReplace: len(replaces) > 0,
	}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return ScheduleSharedDiff(req)
}

func ScheduleSharedDelete(req *pulumirpc.DeleteRequest, client pulumiapi.ScheduleClient) (*pbempty.Empty, error) {
	output, err := ToPulumiServiceSharedScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	err = client.DeleteSchedule(context.Background(), output.Stack, output.ScheduleID)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return ScheduleSharedDelete(req, st.client(ctx))
}

func (st *PulumiServiceDeploymentScheduleResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	input, err := ToPulumiServiceDeploymentScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleReq := pulumiapi.CreateDeploymentScheduleRequest{
		ScheduleCron: input.ScheduleCron,
		ScheduleOnce: input.ScheduleOnce,
		Request: pulumiapi.CreateDeploymentRequest{
			PulumiOperation: input.PulumiOperation,
		},
	}
	scheduleID, err := st.client(ctx).CreateDeploymentSchedule(ctx, input.Stack, scheduleReq)
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
		Id:         path.Join(input.Stack.OrgName, input.Stack.ProjectName, input.Stack.StackName, *scheduleID),
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "stack", "pulumiOperation"} {
		if !inputMap[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	if (inputMap["scheduleCron"].HasValue() && inputMap["timestamp"].HasValue()) ||
		(!inputMap["scheduleCron"].HasValue() && !inputMap["timestamp"].HasValue()) {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "One of scheduleCron or timestamp must be specified but not both",
			Property: "scheduleCron",
		})
	}

	if inputMap["timestamp"].HasValue() && inputMap["timestamp"].IsString() {
		_, err := time.Parse(time.RFC3339, inputMap["timestamp"].StringValue())
		if err != nil {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("timestamp failed to parse due to: %s", err),
				Property: "timestamp",
			})
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.GetNews(), Failures: failures}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	previousOutput, err := ToPulumiServiceSharedScheduleOutput(req.GetOlds())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceDeploymentScheduleInput(req.GetNews())
	if err != nil {
		return nil, err
	}

	updateReq := pulumiapi.CreateDeploymentScheduleRequest{
		ScheduleCron: input.ScheduleCron,
		ScheduleOnce: input.ScheduleOnce,
		Request: pulumiapi.CreateDeploymentRequest{
			PulumiOperation: input.PulumiOperation,
		},
	}
	scheduleID, err := st.client(ctx).UpdateDeploymentSchedule(ctx, input.Stack, updateReq, previousOutput.ScheduleID)
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

func (st *PulumiServiceDeploymentScheduleResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	output, err := ToPulumiServiceSharedScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceDeploymentScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleID, err := st.client(ctx).GetSchedule(ctx, output.Stack, output.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to read DeploymentSchedule (%q): %w", req.Id, err)
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
		return nil, fmt.Errorf("failed to read DeploymentSchedule (%q): %w", req.Id, err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputProperties,
		Inputs:     outputProperties,
	}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Name() string {
	return "pulumiservice:index:DeploymentSchedule"
}
