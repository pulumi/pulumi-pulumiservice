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

type PulumiServiceDeploymentScheduleResource struct {
	client pulumiapi.ScheduleClient
}

type PulumiServiceDeploymentScheduleInput struct {
	Stack           pulumiapi.StackName
	ScheduleCron    *string    `pulumi:"scheduleCron"`
	ScheduleOnce    *time.Time `pulumi:"scheduleOnce"`
	PulumiOperation string     `pulumi:"pulumiOperation"`
}

type PulumiServiceDeploymentScheduleOutput struct {
	Input      PulumiServiceDeploymentScheduleInput
	ScheduleID string `pulumi:"scheduleID"`
}

func (i *PulumiServiceDeploymentScheduleInput) ToPropertyMap() resource.PropertyMap {
	propertyMap := resource.PropertyMap{}
	propertyMap["organization"] = resource.NewPropertyValue(i.Stack.OrgName)
	propertyMap["project"] = resource.NewPropertyValue(i.Stack.ProjectName)
	propertyMap["stack"] = resource.NewPropertyValue(i.Stack.StackName)

	if i.ScheduleCron != nil {
		propertyMap["scheduleCron"] = resource.NewPropertyValue(i.ScheduleCron)
	}
	if i.ScheduleOnce != nil {
		propertyMap["timestamp"] = resource.NewPropertyValue(i.ScheduleOnce.Format(time.RFC3339))
	}
	propertyMap["pulumiOperation"] = resource.NewPropertyValue(i.PulumiOperation)

	return propertyMap
}

func (i *PulumiServiceDeploymentScheduleInput) ToOutputPropertyMap(scheduleID string) resource.PropertyMap {
	propertyMap := i.ToPropertyMap()
	propertyMap["scheduleID"] = resource.NewPropertyValue(scheduleID)
	return propertyMap
}

func (st *PulumiServiceDeploymentScheduleResource) ToPulumiServiceDeploymentScheduleInput(properties *structpb.Struct) (*PulumiServiceDeploymentScheduleInput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := PulumiServiceDeploymentScheduleInput{}

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
	input.Stack = stack

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

func (st *PulumiServiceDeploymentScheduleResource) ToPulumiServiceDeploymentScheduleOutput(properties *structpb.Struct) (*PulumiServiceDeploymentScheduleOutput, error) {
	input, err := st.ToPulumiServiceDeploymentScheduleInput(properties)
	if err != nil {
		return nil, err
	}

	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	output := PulumiServiceDeploymentScheduleOutput{}
	output.Input = *input

	if inputMap["scheduleID"].HasValue() && inputMap["scheduleID"].IsString() {
		output.ScheduleID = inputMap["scheduleID"].StringValue()
	}

	return &output, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	// preprocess olds to remove the `scheduleID` property since it's only an output and shouldn't cause a diff
	if olds["scheduleID"].HasValue() {
		delete(olds, "scheduleID")
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

func (st *PulumiServiceDeploymentScheduleResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	output, err := st.ToPulumiServiceDeploymentScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	err = st.client.DeleteSchedule(context.Background(), output.Input.Stack, output.ScheduleID)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (st *PulumiServiceDeploymentScheduleResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	input, err := st.ToPulumiServiceDeploymentScheduleInput(req.GetProperties())
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
	scheduleID, err := st.client.CreateDeploymentSchedule(context.Background(), input.Stack, scheduleReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		input.ToOutputPropertyMap(*scheduleID),
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

func (st *PulumiServiceDeploymentScheduleResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
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

func (st *PulumiServiceDeploymentScheduleResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	previousOutput, err := st.ToPulumiServiceDeploymentScheduleOutput(req.GetOlds())
	if err != nil {
		return nil, err
	}
	input, err := st.ToPulumiServiceDeploymentScheduleInput(req.GetNews())
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
	scheduleID, err := st.client.UpdateDeploymentSchedule(context.Background(), input.Stack, updateReq, previousOutput.ScheduleID)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		input.ToOutputPropertyMap(*scheduleID),
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
	output, err := st.ToPulumiServiceDeploymentScheduleOutput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	scheduleID, err := st.client.GetSchedule(context.Background(), output.Input.Stack, output.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to read DeploymentSchedule (%q): %w", req.Id, err)
	}
	if scheduleID == nil {
		// if the tag doesn't exist, then return empty response
		return &pulumirpc.ReadResponse{}, nil
	}

	outputProperties, err := plugin.MarshalProperties(
		output.Input.ToPropertyMap(),
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

func (st *PulumiServiceDeploymentScheduleResource) Configure(_ PulumiServiceConfig) {
}
