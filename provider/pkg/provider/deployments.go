package provider

import (
	"context"
	"fmt"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceRunDeploymentFunction struct {
	client *pulumiapi.Client
}

func (f *PulumiServiceRunDeploymentFunction) Name() string {
	return "pulumiservice:index:RunDeployment"
}

type PulumiServiceRunDeploymentInput struct {
	pulumiapi.CreateDeploymentRequest
	Stack pulumiapi.StackName
}

func (f *PulumiServiceRunDeploymentFunction) ToPulumiServiceRunDeploymentInput(inputMap resource.PropertyMap) PulumiServiceRunDeploymentInput {
	input := PulumiServiceRunDeploymentInput{}

	input.InheritSettings = inputMap["inheritSettings"].BoolValue()
	input.Operation = apitype.PulumiOperation(inputMap["operation"].StringValue())
	input.Stack.StackName = inputMap["stack"].StringValue()
	input.Stack.ProjectName = inputMap["project"].StringValue()
	input.Stack.OrgName = inputMap["organization"].StringValue()

	return input
}

func (f *PulumiServiceRunDeploymentFunction) Invoke(req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	ctx := context.Background()
	label := fmt.Sprintf("Invoke(%s)", f.Name())
	inputsMap, err := plugin.UnmarshalProperties(
		req.GetArgs(), plugin.MarshalOptions{Label: label, KeepUnknowns: true})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %v args during an Invoke call: %w", f.Name(), err)
	}

	inputs := f.ToPulumiServiceRunDeploymentInput(inputsMap)
	args := inputs.CreateDeploymentRequest
	resp, err := f.client.CreateDeployment(ctx, inputs.Stack, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment for stack (%s): %w", inputs.Stack.String(), err)
	}

	pm := resource.PropertyMap{}
	pm["id"] = resource.NewStringProperty(resp.ID)
	pm["version"] = resource.NewNumberProperty(float64(resp.Version))
	pm["consoleUrl"] = resource.NewStringProperty(resp.ConsoleURL)

	invokeResult, err := plugin.MarshalProperties(pm, plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %v result: %w", f.Name(), err)
	}

	return &pulumirpc.InvokeResponse{Return: invokeResult}, nil
}

type PulumiServiceGetDeploymentFunction struct {
	client *pulumiapi.Client
}

func (f *PulumiServiceGetDeploymentFunction) Name() string {
	return "pulumiservice:index:GetDeployment"
}

func (f *PulumiServiceGetDeploymentFunction) Invoke(req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, nil
}
