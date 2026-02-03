package resources

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/asset"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

const defaultProject = "default"

type PulumiServiceEnvironmentResource struct {
	Client esc_client.Client
}

type PulumiServiceEnvironmentInput struct {
	OrgName     string
	ProjectName string
	EnvName     string
	Yaml        string
}

type PulumiServiceEnvironmentOutput struct {
	input    PulumiServiceEnvironmentInput
	revision int
}

func (i *PulumiServiceEnvironmentInput) ToPropertyMap() (resource.PropertyMap, error) {
	propertyMap := resource.PropertyMap{}
	propertyMap["organization"] = resource.NewPropertyValue(i.OrgName)
	propertyMap["project"] = resource.NewPropertyValue(i.ProjectName)
	propertyMap["name"] = resource.NewPropertyValue(i.EnvName)
	propertyMap["yaml"] = resource.MakeSecret(resource.NewStringProperty(i.Yaml))

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
	input.OrgName = inputMap["organization"].StringValue()
	input.EnvName = inputMap["name"].StringValue()

	inputYaml := inputMap["yaml"]
	if inputYaml.IsAsset() {
		input.Yaml = inputYaml.AssetValue().Text
	} else {
		input.Yaml = inputYaml.StringValue()
	}

	// Set project to "default" if not in input
	input.ProjectName = defaultProject

	inputProject := inputMap["project"]
	if inputProject.HasValue() && inputProject.IsString() {
		input.ProjectName = inputProject.StringValue()
	}

	return &input, nil
}

func getBytesFromAsset(asset *asset.Asset) ([]byte, error) {
	reader, err := asset.Read()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = reader.Close()
		if err != nil {
			fmt.Println("failed to close reading asset: %w", err)
		}
	}()
	return io.ReadAll(reader)
}

func (st *PulumiServiceEnvironmentResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
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
		"project":      true,
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

	err = st.Client.DeleteEnvironment(context.Background(), input.OrgName, input.ProjectName, input.EnvName)
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
	_, diagnostics, err := st.Client.CheckYAMLEnvironment(
		context.Background(),
		input.OrgName,
		[]byte(input.Yaml),
		esc_client.CheckYAMLOption{},
	)
	if diagnostics != nil {
		return nil, fmt.Errorf("failed to check environment, yaml code failed following checks: %+v", diagnostics)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check environment due to error: %+v", err)
	}

	// Then create environment, and update it with yaml provided. ESC API architecture doesn't let you do it in one call
	err = st.Client.CreateEnvironmentWithProject(context.Background(), input.OrgName, input.ProjectName, input.EnvName)
	if err != nil {
		return nil, fmt.Errorf("failed to create new environment due to error: %+v", err)
	}
	diagnostics, revision, err := st.Client.UpdateEnvironmentWithRevision(
		context.Background(),
		input.OrgName,
		input.ProjectName,
		input.EnvName,
		[]byte(input.Yaml),
		"",
	)
	if diagnostics != nil {
		return nil, fmt.Errorf(
			"failed to update brand new environment with pre-checked yaml, due to failing the following checks: %+v \n"+
				"This should never happen, if you're seeing this message there's likely a bug in ESC APIs",
			diagnostics,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to push yaml into environment due to error: %+v", err)
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
		Id:         path.Join(input.OrgName, input.ProjectName, input.EnvName),
		Properties: outputProperties,
	}, nil
}

func (st *PulumiServiceEnvironmentResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "name", "yaml"} {
		input := inputMap[(p)]

		if !input.HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		} else if p != "yaml" && !input.IsComputed() && strings.Contains(util.GetSecretOrStringValue(input), "/") {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("'%s' property contains `/` illegal character", p),
				Property: string(p),
			})
		}
	}

	var stringYaml string
	inputYaml := inputMap["yaml"]
	if !inputYaml.IsComputed() {
		if inputYaml.IsSecret() {
			inputYaml = inputYaml.SecretValue().Element
		}

		// After unwrapping secret, check again if the inner value is computed
		if !inputYaml.IsComputed() {
			if inputYaml.IsAsset() {
				yamlBytes, err := getBytesFromAsset(inputYaml.AssetValue())
				if err != nil {
					return nil, err
				}
				stringYaml = string(yamlBytes)
			} else {
				stringYaml = inputYaml.StringValue()
			}
		}
	}

	trimmedYaml := strings.TrimSpace(stringYaml)
	inputMap["yaml"] = resource.MakeSecret(resource.NewStringProperty(trimmedYaml))

	inputs, err := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: inputs, Failures: failures}, nil
}

func (st *PulumiServiceEnvironmentResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	input, err := ToPulumiServiceEnvironmentInput(req.GetNews())
	if err != nil {
		return nil, err
	}

	diagnostics, revision, _ := st.Client.UpdateEnvironmentWithRevision(
		context.Background(),
		input.OrgName,
		input.ProjectName,
		input.EnvName,
		[]byte(input.Yaml),
		"",
	)
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
	// Split Id into either:
	//   <org>/<project>/<env> or
	//   <org>/<env> (legacy pattern)
	var orgName, projectName, envName string

	splitID := strings.Split(req.Id, "/")
	switch len(splitID) {
	case 3:
		orgName = splitID[0]
		projectName = splitID[1]
		envName = splitID[2]
	case 2:
		// Legacy pattern. Assume "default" project
		orgName = splitID[0]
		projectName = defaultProject
		envName = splitID[1]
	default:
		return nil, fmt.Errorf("invalid environment id: %s", req.Id)
	}

	retrievedYaml, _, revision, err := st.Client.GetEnvironment(
		context.Background(),
		orgName,
		projectName,
		envName,
		"",
		false,
	)
	if err != nil {
		return &pulumirpc.ReadResponse{Id: "", Properties: nil}, nil
	}

	stringYaml := string(retrievedYaml)
	trimmedYaml := strings.TrimSpace(stringYaml)

	input := PulumiServiceEnvironmentInput{
		OrgName:     orgName,
		ProjectName: projectName,
		EnvName:     envName,
		Yaml:        trimmedYaml,
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
