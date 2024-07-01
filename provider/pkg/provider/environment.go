package provider

import (
	"context"
	"fmt"
	"path"
	"strings"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/asset"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type PulumiServiceEnvironmentResource struct {
	client esc_client.Client
}

type PulumiServiceEnvironmentInput struct {
	OrgName string
	EnvName string
	Yaml    []byte
}

type PulumiServiceEnvironmentOutput struct {
	input    PulumiServiceEnvironmentInput
	revision int
}

func (i *PulumiServiceEnvironmentInput) ToPropertyMap() (resource.PropertyMap, error) {
	propertyMap := resource.PropertyMap{}
	propertyMap["organization"] = resource.NewPropertyValue(i.OrgName)
	propertyMap["name"] = resource.NewPropertyValue(i.EnvName)

	yamlAsset, err := asset.FromText(strings.TrimSuffix(string(i.Yaml), "\n"))
	if err != nil {
		return nil, err
	}
	propertyMap["yaml"] = resource.MakeSecret(resource.NewAssetProperty(yamlAsset))

	return propertyMap, nil
}

func (i *PulumiServiceEnvironmentOutput) ToPropertyMap() (resource.PropertyMap, error) {
	propertyMap, err := i.input.ToPropertyMap()
	if err != nil {
		return nil, err
	}

	propertyMap["revision"] = resource.NewPropertyValue(i.revision)

	return propertyMap, nil
}

func ToPulumiServiceEnvironmentInput(properties *structpb.Struct) (*PulumiServiceEnvironmentInput, error) {
	inputMap, err := plugin.UnmarshalProperties(properties, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := PulumiServiceEnvironmentInput{}
	if inputMap["organization"].HasValue() && inputMap["organization"].IsString() {
		input.OrgName = inputMap["organization"].StringValue()
	} else {
		return nil, fmt.Errorf("failed to unmarshal organization value from properties: %s", inputMap)
	}
	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.EnvName = inputMap["name"].StringValue()
	} else {
		return nil, fmt.Errorf("failed to unmarshal environment name value from properties: %s", inputMap)
	}
	if inputMap["yaml"].HasValue() && inputMap["yaml"].IsAsset() {
		input.Yaml = []byte(inputMap["yaml"].AssetValue().Text)
	} else {
		return nil, fmt.Errorf("failed to unmarshal yaml value from properties: %s", inputMap)
	}

	return &input, nil
}

func (st *PulumiServiceEnvironmentResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
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
		"name":         true,
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

func (st *PulumiServiceEnvironmentResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	input, err := ToPulumiServiceEnvironmentInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	err = st.client.DeleteEnvironment(context.Background(), input.OrgName, input.EnvName)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (st *PulumiServiceEnvironmentResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	input, err := ToPulumiServiceEnvironmentInput(req.GetProperties())
	if err != nil {
		return nil, err
	}

	// First check if yaml is valid
	_, diagnostics, err := st.client.CheckYAMLEnvironment(context.Background(), input.OrgName, input.Yaml)
	if err != nil {
		return nil, err
	}
	if diagnostics != nil {
		return nil, fmt.Errorf("failed to create environment, yaml code failed following checks: %+v", diagnostics)
	}

	// Then create environment, and update it with yaml provided. ESC API architecture doesn't let you do it in one call
	err = st.client.CreateEnvironment(context.Background(), input.OrgName, input.EnvName)
	if err != nil {
		return nil, err
	}
	diagnostics, revision, err := st.client.UpdateEnvironmentWithRevision(context.Background(), input.OrgName, input.EnvName, input.Yaml, "")
	if err != nil {
		return nil, err
	}
	if diagnostics != nil {
		return nil, fmt.Errorf("failed to update brand new environment with pre-checked yaml, due to failing the following checks: %+v \n"+
			"This should never happen, if you're seeing this message there's likely a bug in ESC APIs", diagnostics)
	}

	output := PulumiServiceEnvironmentOutput{
		input:    *input,
		revision: revision,
	}

	propertyMap, err := output.ToPropertyMap()
	if err != nil {
		return nil, err
	}
	outputProperties, err := plugin.MarshalProperties(
		propertyMap,
		plugin.MarshalOptions{
			KeepSecrets: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.OrgName, input.EnvName),
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceEnvironmentResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "name", "yaml"} {
		if !inputMap[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.GetNews(), Failures: failures}, nil
}

func (st *PulumiServiceEnvironmentResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	input, err := ToPulumiServiceEnvironmentInput(req.GetNews())
	if err != nil {
		return nil, err
	}

	diagnostics, revision, err := st.client.UpdateEnvironmentWithRevision(context.Background(), input.OrgName, input.EnvName, input.Yaml, "")
	if err != nil {
		return nil, err
	}
	if diagnostics != nil {
		return nil, fmt.Errorf("failed to update environment, yaml code failed following checks: %+v", diagnostics)
	}

	output := PulumiServiceEnvironmentOutput{
		input:    *input,
		revision: revision,
	}

	propertyMap, err := output.ToPropertyMap()
	if err != nil {
		return nil, err
	}
	outputProperties, err := plugin.MarshalProperties(
		propertyMap,
		plugin.MarshalOptions{
			KeepSecrets: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceEnvironmentResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	splitID := strings.Split(req.Id, "/")
	if len(splitID) < 2 {
		return nil, fmt.Errorf("invalid environment id: %s", req.Id)
	}
	orgName := splitID[0]
	envName := splitID[1]

	retrievedYaml, _, revision, err := st.client.GetEnvironment(context.Background(), orgName, envName, "", false)
	if err != nil {
		return &pulumirpc.ReadResponse{Id: "", Properties: nil}, nil
	}

	input := PulumiServiceEnvironmentInput{
		OrgName: orgName,
		EnvName: envName,
		Yaml:    retrievedYaml,
	}

	result := PulumiServiceEnvironmentOutput{
		input:    input,
		revision: revision,
	}

	inputMap, err := input.ToPropertyMap()
	if err != nil {
		return nil, err
	}
	inputs, err := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{
			KeepSecrets: true,
		},
	)
	if err != nil {
		return nil, err
	}

	propertyMap, err := result.ToPropertyMap()
	if err != nil {
		return nil, err
	}
	properties, err := plugin.MarshalProperties(
		propertyMap,
		plugin.MarshalOptions{
			KeepSecrets: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     inputs,
	}, nil
}

func (st *PulumiServiceEnvironmentResource) Name() string {
	return "pulumiservice:index:Environment"
}

func (st *PulumiServiceEnvironmentResource) Configure(_ PulumiServiceConfig) {
}
