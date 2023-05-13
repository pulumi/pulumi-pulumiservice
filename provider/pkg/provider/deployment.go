package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil/rpcerror"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/structpb"
)

type PulumiServiceDeploymentInput struct {
	Stack    pulumiapi.StackName
	Config   resource.PropertyMap
	Settings pulumiapi.DeploymentSettings
}

func (d *PulumiServiceDeploymentInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["stack"] = resource.NewPropertyValue(d.Stack.String())
	pm["config"] = resource.NewObjectProperty(d.Config.Copy())
	pm["settings"] = resource.NewObjectProperty(deploymentSettingsToPropertyMap(d.Settings))
	return pm
}

type PulumiServiceDeploymentResource struct {
	cancelCtx context.Context
	host      *provider.HostClient
	client    *pulumiapi.Client
}

func (d *PulumiServiceDeploymentResource) ToPulumiServiceDeploymentInput(inputMap resource.PropertyMap) (*PulumiServiceDeploymentInput, error) {
	input := PulumiServiceDeploymentInput{}

	if err := input.Stack.FromID(getSecretOrStringValue(inputMap["stack"])); err != nil {
		return nil, err
	}

	if config := inputMap["config"]; config.HasValue() {
		input.Config = mustValue[resource.PropertyMap](config)
	}
	if settings := inputMap["settings"]; settings.HasValue() {
		input.Settings = toDeploymentSettings(mustValue[resource.PropertyMap](settings))
	}

	return &input, nil
}

func (d *PulumiServiceDeploymentResource) Name() string {
	return "pulumiservice:index:Deployment"
}

func (d *PulumiServiceDeploymentResource) Configure(config PulumiServiceConfig) {}

func (d *PulumiServiceDeploymentResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure

	if stack := news["stack"]; !stack.HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "missing required property 'stack'",
			Property: "stack",
		})
	} else if stackID, ok := tryValue[string](stack); ok {
		var name pulumiapi.StackName
		if err := name.FromID(stackID); err != nil {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   err.Error(),
				Property: "stack",
			})
		}
	} else {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   fmt.Sprintf("expected a string, not a %v", stack.TypeString()),
			Property: "stack",
		})
	}

	if config := news["config"]; config.HasValue() && !isType[resource.PropertyMap](config) {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   fmt.Sprintf("expected an object, not a %v", config.TypeString()),
			Property: "config",
		})
	}

	if settings := news["settings"]; settings.HasValue() && !isType[resource.PropertyMap](settings) {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   fmt.Sprintf("expected an object, not a %v", settings.TypeString()),
			Property: "settings",
		})
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (d *PulumiServiceDeploymentResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	dd := plugin.NewDetailedDiffFromObjectDiff(diffs)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	for k, v := range dd {
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	// We always consider the version to differ, as we're always going to run a deployment.
	detailedDiffs["version"] = &pulumirpc.PropertyDiff{Kind: pulumirpc.PropertyDiff_UPDATE}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		DetailedDiff:        detailedDiffs,
		DeleteBeforeReplace: true,
		HasDetailedDiff:     true,
	}, nil
}

func mustQuotePropertyName(name resource.PropertyKey) bool {
	for i, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || i > 0 && r >= '0' && r <= '9') {
			return true
		}
	}
	return false
}

func flattenPrimitive[T bool | float64 | string](m map[string]apitype.SecretValue, path string, v T) {
	m[path] = apitype.SecretValue{Value: fmt.Sprintf("%v", v)}
}

func flattenConfigValue(m map[string]apitype.SecretValue, path string, v resource.PropertyValue) {
	switch {
	case v.IsBool():
		flattenPrimitive(m, path, v.BoolValue())
	case v.IsNumber():
		flattenPrimitive(m, path, v.NumberValue())
	case v.IsString():
		flattenPrimitive(m, path, v.StringValue())
	case v.IsSecret():
		elems := map[string]apitype.SecretValue{}
		flattenConfigValue(elems, path, v.SecretValue().Element)
		for k, v := range elems {
			m[k] = apitype.SecretValue{Value: v.Value, Secret: true}
		}
	case v.IsArray():
		for i, v := range v.ArrayValue() {
			flattenConfigValue(m, fmt.Sprintf("%s[%d]", path, i), v)
		}
	case v.IsObject():
		for k, v := range v.ObjectValue() {
			if !mustQuotePropertyName(k) {
				flattenConfigValue(m, fmt.Sprintf("%s.%s", path, k), v)
			} else {
				flattenConfigValue(m, fmt.Sprintf("%s[%q]", path, k), v)
			}
		}
	}
}

func sortedKeys[M ~map[K]V, K constraints.Ordered, V any](m M) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}

func makeConfigCommands(config resource.PropertyMap) ([]string, map[string]apitype.SecretValue) {
	m := map[string]apitype.SecretValue{}
	for k, v := range config {
		flattenConfigValue(m, string(k), v)
	}
	commands, secrets := make([]string, len(m)), map[string]apitype.SecretValue{}
	for i, path := range sortedKeys(m) {
		v := m[path]

		secretFlag, configValue := "", v.Value
		if v.Secret {
			env := fmt.Sprintf("PULUMI_SECRET_%v", i)
			secretFlag, secrets[env], configValue = " --secret", v, env
		}

		commands[i] = fmt.Sprintf("pulumi config set%v --path \"%q\" %q", secretFlag, path, configValue)
	}
	return commands, secrets
}

func (d *PulumiServiceDeploymentResource) prepareDeployment(ctx context.Context, inputsMap resource.PropertyMap) (*PulumiServiceDeploymentInput, error) {
	inputs, err := d.ToPulumiServiceDeploymentInput(inputsMap)
	if err != nil {
		return nil, err
	}

	// apply config as part of precommands.
	if len(inputs.Config) != 0 {
		configCommands, additionalSecrets := makeConfigCommands(inputs.Config)

		// fetch any existing pre-run commands
		var preRunCommands []string
		if inputs.Settings.OperationContext != nil {
			preRunCommands = inputs.Settings.OperationContext.PreRunCommands
		}
		if len(preRunCommands) == 0 {
			settings, err := d.client.GetDeploymentSettings(ctx, inputs.Stack)
			if err != nil {
				return nil, err
			}
			if settings.OperationContext != nil {
				preRunCommands = settings.OperationContext.PreRunCommands
			}
		}

		if inputs.Settings.OperationContext == nil {
			inputs.Settings.OperationContext = &pulumiapi.OperationContext{}
		}
		inputs.Settings.OperationContext.PreRunCommands = append(configCommands, preRunCommands...)

		if inputs.Settings.OperationContext.EnvironmentVariables == nil {
			inputs.Settings.OperationContext.EnvironmentVariables = map[string]apitype.SecretValue{}
		}
		for k, v := range additionalSecrets {
			inputs.Settings.OperationContext.EnvironmentVariables[k] = v
		}
	}
	return inputs, nil
}

func (d *PulumiServiceDeploymentResource) runDeployment(ctx context.Context, urn resource.URN, operation string, inputsMap resource.PropertyMap, timeout float64) (string, resource.PropertyMap, error) {
	if timeout != 0 {
		withTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout))
		defer cancel()
		ctx = withTimeout
	}

	inputs, err := d.prepareDeployment(ctx, inputsMap)
	if err != nil {
		return "", nil, err
	}
	id, state := inputs.Stack.String(), inputsMap.Copy()

	createResp, err := d.client.CreateDeployment(ctx, inputs.Stack, pulumiapi.CreateDeploymentRequest{
		DeploymentSettings: inputs.Settings,
		Op:                 operation,
		InheritSettings:    true,
	})
	if err != nil {
		return id, state, err
	}
	deploymentID := createResp.ID
	state["version"] = resource.NewNumberProperty(float64(createResp.Version))

	var continuationToken string
	for {
		logs, err := d.client.GetDeploymentLogs(ctx, inputs.Stack, deploymentID, continuationToken)
		if err != nil {
			return id, state, err
		}
		for _, line := range logs.Lines {
			if line.Header != "" {
				d.host.LogStatus(ctx, diag.Info, urn, line.Header)
			}
			d.host.LogStatus(ctx, diag.Info, urn, line.Line)
		}
		if logs.NextToken == "" {
			break
		}
		continuationToken = logs.NextToken
	}

	getResp, err := d.client.GetDeployment(ctx, inputs.Stack, deploymentID)
	if err != nil {
		return id, state, err
	}
	switch getResp.Status {
	case "succeeded":
		return id, state, nil
	case "failed":
		return id, state, fmt.Errorf("deployment failed")
	default:
		return id, state, fmt.Errorf("unexpected deployment status %v" + getResp.Status)
	}
}

func (d *PulumiServiceDeploymentResource) runUpdateDeployment(ctx context.Context, urn resource.URN, rpcInputs *structpb.Struct, timeout float64) (string, *structpb.Struct, error) {
	inputsMap, err := plugin.UnmarshalProperties(rpcInputs, plugin.MarshalOptions{
		KeepUnknowns: true,
		SkipNulls:    true,
		KeepSecrets:  true,
		RejectAssets: true,
	})
	if err != nil {
		return "", nil, err
	}

	id, stateMap, err := d.runDeployment(ctx, urn, "update", inputsMap, timeout)
	if stateMap == nil {
		return id, nil, err
	}

	state, merr := plugin.MarshalProperties(stateMap, plugin.MarshalOptions{
		KeepUnknowns: true,
		SkipNulls:    true,
		KeepSecrets:  true,
	})
	if merr != nil {
		return id, nil, errors.Join(err, fmt.Errorf("marshaling state: %w", merr))
	}
	if err != nil {
		detail := pulumirpc.ErrorResourceInitFailed{
			Id:         id,
			Properties: state,
			Inputs:     rpcInputs,
		}
		return id, nil, rpcerror.WithDetails(rpcerror.New(codes.Unknown, err.Error()), &detail)
	}
	return id, state, nil
}

func (d *PulumiServiceDeploymentResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := d.cancelCtx

	id, state, err := d.runUpdateDeployment(ctx, resource.URN(req.GetUrn()), req.GetProperties(), req.GetTimeout())
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{
		Id:         id,
		Properties: state,
	}, nil
}

func (d *PulumiServiceDeploymentResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := d.cancelCtx

	_, state, err := d.runUpdateDeployment(ctx, resource.URN(req.GetUrn()), req.GetNews(), req.GetTimeout())
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: state,
	}, nil
}

func (d *PulumiServiceDeploymentResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: req.GetProperties(),
		Inputs:     req.GetInputs(),
	}, nil
}

func (d *PulumiServiceDeploymentResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := d.cancelCtx

	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	_, _, err = d.runDeployment(ctx, resource.URN(req.GetUrn()), "destroy", inputsMap, req.GetTimeout())
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}
