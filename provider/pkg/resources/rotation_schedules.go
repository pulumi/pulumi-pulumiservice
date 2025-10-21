package resources

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type PulumiServiceEnvironmentRotationScheduleResource struct{}

type PulumiServiceEnvironmentRotationScheduleInput struct {
	Environment  pulumiapi.EnvironmentIdentifier
	ScheduleCron *string    `pulumi:"scheduleCron"`
	ScheduleOnce *time.Time `pulumi:"scheduleOnce"`
}

type PulumiServiceEnvironmentRotationScheduleOutput struct {
	Environment pulumiapi.EnvironmentIdentifier
	ScheduleID  string `pulumi:"scheduleId"`
}

func EnvironmentToPropertyMap(environment pulumiapi.EnvironmentIdentifier) resource.PropertyMap {
	propertyMap := resource.PropertyMap{}
	propertyMap["organization"] = resource.NewPropertyValue(environment.OrgName)
	propertyMap["project"] = resource.NewPropertyValue(environment.ProjectName)
	propertyMap["environment"] = resource.NewPropertyValue(environment.EnvName)
	return propertyMap
}

func (i *PulumiServiceEnvironmentRotationScheduleInput) ToPropertyMap() resource.PropertyMap {
	propertyMap := EnvironmentToPropertyMap(i.Environment)

	if i.ScheduleCron != nil {
		propertyMap["scheduleCron"] = resource.NewPropertyValue(i.ScheduleCron)
	}
	if i.ScheduleOnce != nil {
		propertyMap["timestamp"] = resource.NewPropertyValue(i.ScheduleOnce.Format(time.RFC3339))
	}

	return propertyMap
}

func ParseEnvironment(inputMap resource.PropertyMap) (*pulumiapi.EnvironmentIdentifier, error) {
	var environment pulumiapi.EnvironmentIdentifier
	if inputMap["organization"].HasValue() && inputMap["organization"].IsString() {
		organization := inputMap["organization"].StringValue()
		environment.OrgName = organization
	} else {
		return nil, fmt.Errorf("failed to unmarshal organization value from properties: %s", inputMap)
	}
	if inputMap["project"].HasValue() && inputMap["project"].IsString() {
		project := inputMap["project"].StringValue()
		environment.ProjectName = project
	} else {
		return nil, fmt.Errorf("failed to unmarshal project value from properties: %s", inputMap)
	}
	if inputMap["environment"].HasValue() && inputMap["environment"].IsString() {
		environmentName := inputMap["environment"].StringValue()
		environment.EnvName = environmentName
	} else {
		return nil, fmt.Errorf("failed to unmarshal environmentName value from properties: %s", inputMap)
	}
	return &environment, nil
}

func ToPulumiServiceEnvironmentRotationScheduleInput(properties *structpb.Struct) (*PulumiServiceEnvironmentRotationScheduleInput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	input := PulumiServiceEnvironmentRotationScheduleInput{}
	environmentIdentifier, err := ParseEnvironment(inputMap)
	if err != nil {
		return nil, err
	}
	input.Environment = *environmentIdentifier

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

func ToPulumiServiceEnvironmentRotationScheduleOutput(properties *structpb.Struct) (*PulumiServiceEnvironmentRotationScheduleOutput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}
	environment, err := ParseEnvironment(inputMap)
	if err != nil {
		return nil, err
	}

	output := PulumiServiceEnvironmentRotationScheduleOutput{}
	output.Environment = *environment

	if inputMap["scheduleId"].HasValue() && inputMap["scheduleId"].IsString() {
		output.ScheduleID = inputMap["scheduleId"].StringValue()
	}

	return &output, nil
}

func EnvironmentScheduleSharedDiffMaps(olds resource.PropertyMap, news resource.PropertyMap) (*pulumirpc.DiffResponse, error) {
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

func (st *PulumiServiceEnvironmentRotationScheduleResource) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	return EnvironmentScheduleSharedDiffMaps(olds, news)
}

func EnvironmentScheduleSharedDelete(req *pulumirpc.DeleteRequest, client pulumiapi.EnvironmentScheduleClient) (*pbempty.Empty, error) {
	output, err := ToPulumiServiceEnvironmentRotationScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	err = client.DeleteEnvironmentSchedule(context.Background(), output.Environment, output.ScheduleID)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (st *PulumiServiceEnvironmentRotationScheduleResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	client := config.GetClient[pulumiapi.EnvironmentScheduleClient](ctx)
	return EnvironmentScheduleSharedDelete(req, client)
}

func (st *PulumiServiceEnvironmentRotationScheduleResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	client := config.GetClient[pulumiapi.EnvironmentScheduleClient](ctx)
	input, err := ToPulumiServiceEnvironmentRotationScheduleInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleReq := pulumiapi.CreateEnvironmentRotationScheduleRequest{
		ScheduleCron: input.ScheduleCron,
		ScheduleOnce: input.ScheduleOnce,
	}
	scheduleID, err := client.CreateEnvironmentRotationSchedule(ctx, input.Environment, scheduleReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(*scheduleID, input.ToPropertyMap()),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.Environment.OrgName, input.Environment.ProjectName, input.Environment.EnvName, "rotations", *scheduleID),
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceEnvironmentRotationScheduleResource) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "environment"} {
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

func (st *PulumiServiceEnvironmentRotationScheduleResource) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	client := config.GetClient[pulumiapi.EnvironmentScheduleClient](ctx)
	previousOutput, err := ToPulumiServiceEnvironmentRotationScheduleOutput(req.GetOlds())
	if err != nil {
		return nil, err
	}
	input, err := ToPulumiServiceEnvironmentRotationScheduleInput(req.GetNews())
	if err != nil {
		return nil, err
	}

	updateReq := pulumiapi.CreateEnvironmentRotationScheduleRequest{
		ScheduleCron: input.ScheduleCron,
		ScheduleOnce: input.ScheduleOnce,
	}
	scheduleID, err := client.UpdateEnvironmentRotationSchedule(ctx, input.Environment, updateReq, previousOutput.ScheduleID)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(*scheduleID, input.ToPropertyMap()),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceEnvironmentRotationScheduleResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	client := config.GetClient[pulumiapi.EnvironmentScheduleClient](ctx)
	environment, scheduleID, err := ParseEnvironmentScheduleID(req.Id, "rotations")
	if err != nil {
		return nil, err
	}

	scheduleResponse, err := client.GetEnvironmentSchedule(ctx, *environment, *scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to read EnvironmentRotationSchedule (%q): %w", req.Id, err)
	}
	if scheduleResponse == nil {
		// if schedule doesn't exist, then return empty response to delete it from state
		return &pulumirpc.ReadResponse{}, nil
	}

	var scheduleOnce *time.Time = nil
	if scheduleResponse.ScheduleOnce != nil {
		parsed, err := time.Parse(time.DateTime, *scheduleResponse.ScheduleOnce)
		if err != nil {
			return nil, fmt.Errorf("failed to read EnvironmentRotationSchedule (%q): %w", req.Id, err)
		}
		scheduleOnce = &parsed
	}
	input := PulumiServiceEnvironmentRotationScheduleInput{
		Environment:  *environment,
		ScheduleCron: scheduleResponse.ScheduleCron,
		ScheduleOnce: scheduleOnce,
	}

	inputs, err := plugin.MarshalProperties(
		input.ToPropertyMap(),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read EnvironmentRotationSchedule (%q): %w", req.Id, err)
	}
	outputProperties, err := plugin.MarshalProperties(
		AddScheduleIdToPropertyMap(*scheduleID, input.ToPropertyMap()),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read EnvironmentRotationSchedule (%q): %w", req.Id, err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputProperties,
		Inputs:     inputs,
	}, nil
}

func (st *PulumiServiceEnvironmentRotationScheduleResource) Name() string {
	return "pulumiservice:index:EnvironmentRotationSchedule"
}

func ParseEnvironmentScheduleID(id string, scheduleType string) (*pulumiapi.EnvironmentIdentifier, *string, error) {
	splitID := strings.Split(id, "/")
	if len(splitID) < 4 {
		return nil, nil, fmt.Errorf("invalid environment id: %s", id)
	}
	envId := pulumiapi.EnvironmentIdentifier{
		OrgName:     splitID[0],
		ProjectName: splitID[1],
		EnvName:     splitID[2],
	}
	if scheduleType == "" {
		if len(splitID) != 4 {
			return nil, nil, fmt.Errorf("invalid schedule id: %s", id)
		}
		return &envId, &splitID[3], nil
	}
	if len(splitID) != 5 || splitID[3] != scheduleType {
		return nil, nil, fmt.Errorf("invalid schedule id: %s", id)
	}
	return &envId, &splitID[4], nil
}
