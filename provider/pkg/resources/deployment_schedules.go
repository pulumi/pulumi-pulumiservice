package resources

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type PulumiServiceDeploymentScheduleResource struct {
	Client pulumiapi.StackScheduleClient
}

type PulumiServiceDeploymentScheduleInput struct {
	Stack           pulumiapi.StackIdentifier
	ScheduleCron    *string    `pulumi:"scheduleCron"`
	ScheduleOnce    *time.Time `pulumi:"scheduleOnce"`
	PulumiOperation string     `pulumi:"pulumiOperation"`
}

type PulumiServiceStackScheduleOutput struct {
	Stack      pulumiapi.StackIdentifier
	ScheduleID string `pulumi:"scheduleId"`
}

func StackToPropertyMap(stack pulumiapi.StackIdentifier) resource.PropertyMap {
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

func AddScheduleIDToPropertyMap(scheduleID string, propertyMap resource.PropertyMap) resource.PropertyMap {
	propertyMap["scheduleId"] = resource.NewPropertyValue(scheduleID)
	return propertyMap
}

func ParseStack(inputMap resource.PropertyMap) (*pulumiapi.StackIdentifier, error) {
	var stack pulumiapi.StackIdentifier
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

func ToPulumiServiceDeploymentScheduleInput(
	properties *structpb.Struct,
) (*PulumiServiceDeploymentScheduleInput, error) {
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

func ToPulumiServiceStackScheduleOutput(properties *structpb.Struct) (*PulumiServiceStackScheduleOutput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	stack, err := ParseStack(inputMap)
	if err != nil {
		return nil, err
	}

	output := PulumiServiceStackScheduleOutput{}
	output.Stack = *stack

	if inputMap["scheduleId"].HasValue() && inputMap["scheduleId"].IsString() {
		output.ScheduleID = inputMap["scheduleId"].StringValue()
	}

	return &output, nil
}

func ScheduleSharedDiff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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

	return StackScheduleSharedDiffMaps(olds, news)
}

func StackScheduleSharedDiffMaps(
	olds resource.PropertyMap,
	news resource.PropertyMap,
) (*pulumirpc.DiffResponse, error) {
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
		"timestamp":    true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec // Kind values are bounded by protobuf enum
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

func (st *PulumiServiceDeploymentScheduleResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return ScheduleSharedDiff(req)
}

func StackScheduleSharedDelete(
	req *pulumirpc.DeleteRequest,
	client pulumiapi.StackScheduleClient,
) (*pbempty.Empty, error) {
	output, err := ToPulumiServiceStackScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	err = client.DeleteStackSchedule(context.Background(), output.Stack, output.ScheduleID)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return StackScheduleSharedDelete(req, st.Client)
}

func (st *PulumiServiceDeploymentScheduleResource) Create(
	req *pulumirpc.CreateRequest,
) (*pulumirpc.CreateResponse, error) {
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
	scheduleID, err := st.Client.CreateDeploymentSchedule(context.Background(), input.Stack, scheduleReq)
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
		Id:         path.Join(input.Stack.OrgName, input.Stack.ProjectName, input.Stack.StackName, *scheduleID),
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Check(
	req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
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

func (st *PulumiServiceDeploymentScheduleResource) Update(
	req *pulumirpc.UpdateRequest,
) (*pulumirpc.UpdateResponse, error) {
	previousOutput, err := ToPulumiServiceStackScheduleOutput(req.GetOlds())
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
	scheduleID, err := st.Client.UpdateDeploymentSchedule(
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

func (st *PulumiServiceDeploymentScheduleResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	stack, scheduleID, err := ParseStackScheduleID(req.Id, "")
	if err != nil {
		return nil, err
	}

	scheduleResponse, err := st.Client.GetStackSchedule(context.Background(), *stack, *scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to read DeploymentSchedule (%q): %w", req.Id, err)
	}
	if scheduleResponse == nil {
		// if schedule doesn't exist, then return empty response to delete it from state
		return &pulumirpc.ReadResponse{}, nil
	}

	var scheduleOnce *time.Time
	if scheduleResponse.ScheduleOnce != nil {
		parsed, err := time.Parse(time.DateTime, *scheduleResponse.ScheduleOnce)
		if err != nil {
			return nil, fmt.Errorf("failed to read DeploymentSchedule (%q): %w", req.Id, err)
		}
		scheduleOnce = &parsed
	}
	input := PulumiServiceDeploymentScheduleInput{
		Stack:           *stack,
		ScheduleCron:    scheduleResponse.ScheduleCron,
		ScheduleOnce:    scheduleOnce,
		PulumiOperation: scheduleResponse.Definition.Request.PulumiOperation,
	}

	inputs, err := plugin.MarshalProperties(
		input.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read DeploymentSchedule (%q): %w", req.Id, err)
	}
	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIDToPropertyMap(*scheduleID, input.ToPropertyMap()),
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
		Inputs:     inputs,
	}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Name() string {
	return "pulumiservice:index:DeploymentSchedule"
}

func ParseStackScheduleID(id string, scheduleType string) (*pulumiapi.StackIdentifier, *string, error) {
	splitID := strings.Split(id, "/")
	if len(splitID) < 4 {
		return nil, nil, fmt.Errorf("invalid stack id: %s", id)
	}
	stack := pulumiapi.StackIdentifier{
		OrgName:     splitID[0],
		ProjectName: splitID[1],
		StackName:   splitID[2],
	}
	if scheduleType == "" {
		if len(splitID) != 4 {
			return nil, nil, fmt.Errorf("invalid schedule id: %s", id)
		}
		return &stack, &splitID[3], nil
	}
	if len(splitID) != 5 || splitID[3] != scheduleType {
		return nil, nil, fmt.Errorf("invalid schedule id: %s", id)
	}
	return &stack, &splitID[4], nil
}
