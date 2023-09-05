package provider

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceDeployFunction struct {
	client *pulumiapi.Client
}

func (f *PulumiServiceDeployFunction) Name() string {
	return "pulumiservice:index:Deploy"
}

type PulumiServiceDeployInput struct {
	pulumiapi.CreateDeploymentRequest
	Stack pulumiapi.StackName
}

type PulumiServiceDeployOutput struct {
	ID         string
	Version    int
	ConsoleURL string
	Status     string
}

func (f *PulumiServiceDeployFunction) ToPulumiServiceDeployInput(inputMap resource.PropertyMap) PulumiServiceDeployInput {
	input := PulumiServiceDeployInput{}

	// TODO: Do defaults not work for Invoke?
	if inputMap.HasValue("inheritSettings") {
		input.InheritSettings = inputMap["inheritSettings"].BoolValue()
	} else {
		input.InheritSettings = true
	}
	if inputMap.HasValue("operation") {
		input.Operation = apitype.PulumiOperation(inputMap["operation"].StringValue())
	} else {
		input.Operation = apitype.Update
	}

	input.Stack.StackName = inputMap["stack"].StringValue()
	input.Stack.ProjectName = inputMap["project"].StringValue()
	input.Stack.OrgName = inputMap["organization"].StringValue()

	return input
}

func (f *PulumiServiceDeployFunction) Invoke(ctx context.Context, inputsMap resource.PropertyMap) (*pulumirpc.InvokeResponse, error) {
	inputs := f.ToPulumiServiceDeployInput(inputsMap)

	args := inputs.CreateDeploymentRequest
	resp, err := f.client.CreateDeployment(ctx, inputs.Stack, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment for stack (%s): %w", inputs.Stack.String(), err)
	}

	// Wait for the deployment to complete
	var finalStatus string
	for {
		r, err := f.client.GetDeployment(ctx, inputs.Stack, resp.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment for stack (%s): %w", inputs.Stack.String(), err)
		}

		// Check if the deployment has completed (successfully or not)
		if r.Status == "succeeded" || r.Status == "failed" {
			finalStatus = r.Status
			break // Exit the loop if so
		}

		// If we haven't reached the end of the deployment, sleep for 3 seconds before trying again
		time.Sleep(3 * time.Second)
	}

	pm := resource.PropertyMap{}
	pm["id"] = resource.NewStringProperty(resp.ID)
	pm["version"] = resource.NewNumberProperty(float64(resp.Version))
	pm["consoleUrl"] = resource.NewStringProperty(resp.ConsoleURL)
	pm["status"] = resource.NewStringProperty(finalStatus)

	invokeResult, err := plugin.MarshalProperties(pm, plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %v result: %w", f.Name(), err)
	}

	return &pulumirpc.InvokeResponse{Return: invokeResult}, nil
}

type PulumiServiceCreateDeploymentFunction struct {
	client *pulumiapi.Client
}

func (f *PulumiServiceCreateDeploymentFunction) Name() string {
	return "pulumiservice:index:CreateDeployment"
}

type PulumiServiceCreateDeploymentInput struct {
	pulumiapi.CreateDeploymentRequest
	Stack pulumiapi.StackName
}

func (f *PulumiServiceCreateDeploymentFunction) ToPulumiServiceCreateDeploymentInput(inputMap resource.PropertyMap) PulumiServiceCreateDeploymentInput {
	input := PulumiServiceCreateDeploymentInput{}

	// TODO: Do defaults not work for Invoke?
	if inputMap.HasValue("inheritSettings") {
		input.InheritSettings = inputMap["inheritSettings"].BoolValue()
	} else {
		input.InheritSettings = true
	}
	if inputMap.HasValue("operation") {
		input.Operation = apitype.PulumiOperation(inputMap["operation"].StringValue())
	} else {
		input.Operation = apitype.Update
	}

	input.Stack.StackName = inputMap["stack"].StringValue()
	input.Stack.ProjectName = inputMap["project"].StringValue()
	input.Stack.OrgName = inputMap["organization"].StringValue()

	return input
}

func (f *PulumiServiceCreateDeploymentFunction) Invoke(ctx context.Context, inputsMap resource.PropertyMap) (*pulumirpc.InvokeResponse, error) {
	inputs := f.ToPulumiServiceCreateDeploymentInput(inputsMap)
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

func (f *PulumiServiceGetDeploymentFunction) Invoke(ctx context.Context, inputsMap resource.PropertyMap) (*pulumirpc.InvokeResponse, error) {
	var invokeResponse pulumirpc.InvokeResponse
	var stack pulumiapi.StackName
	stack.StackName = inputsMap["stack"].StringValue()
	stack.ProjectName = inputsMap["project"].StringValue()
	stack.OrgName = inputsMap["organization"].StringValue()

	if (!inputsMap.HasValue("deploymentId") && !inputsMap.HasValue("version")) ||
		(inputsMap.HasValue("deploymentId") && inputsMap.HasValue("version")) {
		invokeResponse.Failures = []*pulumirpc.CheckFailure{
			{
				Reason: "Either deploymentId or version must be specified",
			},
		}
		return &invokeResponse, nil
	}

	var id string
	if inputsMap.HasValue("deploymentId") {
		id = inputsMap["deploymentId"].StringValue()
	} else {
		id = path.Join("version", fmt.Sprintf("%d", int(inputsMap["version"].NumberValue())))
	}

	resp, err := f.client.GetDeployment(ctx, stack, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment for stack (%s): %w", stack.String(), err)
	}

	pm := resource.PropertyMap{}
	pm["id"] = resource.NewStringProperty(resp.ID)
	pm["version"] = resource.NewNumberProperty(float64(resp.Version))
	pm["status"] = resource.NewStringProperty(resp.Status)

	invokeResult, err := plugin.MarshalProperties(pm, plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %v result: %w", f.Name(), err)
	}

	invokeResponse.Return = invokeResult
	return &invokeResponse, nil
}
