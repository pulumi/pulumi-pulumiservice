package provider

import (
	"context"
	"path"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/ryboe/q"
)

type PulumiServiceDeploymentSettingsInput struct {
	pulumiapi.DeploymentSettings
	Stack pulumiapi.StackName
}

func (i *PulumiServiceDeploymentSettingsInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organization"] = resource.NewPropertyValue(i.Stack.OrgName)
	pm["project"] = resource.NewPropertyValue(i.Stack.ProjectName)
	pm["stack"] = resource.NewPropertyValue(i.Stack.StackName)
	return pm
}

type PulumiServiceDeploymentSettingsResource struct {
	client *pulumiapi.Client
}

func (ds *PulumiServiceDeploymentSettingsResource) ToPulumiServiceDeploymentSettingsInput(inputMap resource.PropertyMap) PulumiServiceDeploymentSettingsInput {
	input := PulumiServiceDeploymentSettingsInput{}

	if inputMap["organization"].HasValue() && inputMap["organization"].IsString() {
		input.Stack.OrgName = inputMap["organization"].StringValue()
	}
	if inputMap["project"].HasValue() && inputMap["project"].IsString() {
		input.Stack.ProjectName = inputMap["project"].StringValue()
	}
	if inputMap["stack"].HasValue() && inputMap["stack"].IsString() {
		input.Stack.StackName = inputMap["stack"].StringValue()
	}

	if inputMap["sourceContext"].HasValue() && inputMap["sourceContext"].IsObject() {
		scInput := inputMap["sourceContext"].ObjectValue()
		var sc apitype.SourceContext

		if scInput["git"].HasValue() && scInput["git"].IsObject() {
			gitInput := scInput["git"].ObjectValue()
			var g apitype.SourceContextGit

			if gitInput["repoUrl"].HasValue() && gitInput["repoUrl"].IsString() {
				g.RepoURL = gitInput["repoUrl"].StringValue()
			}
			if gitInput["branch"].HasValue() && gitInput["branch"].IsString() {
				g.Branch = gitInput["branch"].StringValue()
			}
			if gitInput["repoDir"].HasValue() && gitInput["repoDir"].IsString() {
				g.RepoDir = gitInput["repoDir"].StringValue()
			}

			sc.Git = &g
		}

		input.SourceContext = &sc
	}

	if inputMap["operationContext"].HasValue() && inputMap["operationContext"].IsObject() {
		ocInput := inputMap["operationContext"].ObjectValue()
		var oc pulumiapi.OperationContext

		if ocInput["environmentVariables"].HasValue() && ocInput["environmentVariables"].IsObject() {
			ev := map[string]apitype.SecretValue{}
			evInput := ocInput["environmentVariables"].ObjectValue()

			// TODO: Fix secrets
			for k, v := range evInput {
				if v.IsSecret() {
					q.Q("Found a secret: %s", k)
					ev[string(k)] = apitype.SecretValue{Secret: true, Value: v.StringValue()}
				} else {
					q.Q("Not a secret: %s", k)
					ev[string(k)] = apitype.SecretValue{Secret: false, Value: v.StringValue()}
				}
			}

			oc.EnvironmentVariables = ev
		}

		if ocInput["preRunCommands"].HasValue() && ocInput["preRunCommands"].IsArray() {
			pcInput := ocInput["preRunCommands"].ArrayValue()
			pc := make([]string, len(pcInput))

			for i, v := range pcInput {
				if v.IsString() {
					pc[i] = v.StringValue()
				}
			}

			oc.PreRunCommands = pc
		}

		if ocInput["options"].HasValue() && ocInput["options"].IsObject() {
			oInput := ocInput["options"].ObjectValue()
			var o pulumiapi.OperationContextOptions

			if oInput["skipInstallDependencies"].HasValue() && oInput["skipInstallDependencies"].IsBool() {
				o.SkipInstallDependencies = oInput["skipInstallDependencies"].BoolValue()
			}

			if oInput["Shell"].HasValue() && oInput["Shell"].IsString() {
				o.Shell = oInput["Shell"].StringValue()
			}

			oc.Options = &o
		}

		input.OperationContext = &oc
	}

	return input
}

func (ds *PulumiServiceDeploymentSettingsResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	q.Q(olds, news)

	diffs := olds.Diff(news)
	q.Q(diffs)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	replaces := make([]string, len(diffs.ChangedKeys()))
	for i, k := range diffs.ChangedKeys() {
		replaces[i] = string(k)
	}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		Replaces:            replaces,
		DeleteBeforeReplace: true,
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Configure(config PulumiServiceConfig) {}

func (ds *PulumiServiceDeploymentSettingsResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return &pulumirpc.ReadResponse{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputs := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap)
	if err != nil {
		return nil, err
	}
	err = ds.client.DeleteDeploymentSettings(ctx, inputs.Stack)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	q.Q(req.GetProperties())
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	q.Q(inputsMap)
	inputs := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap)
	q.Q(inputs)
	settings := inputs.DeploymentSettings
	err = ds.client.CreateDeploymentSettings(ctx, inputs.Stack, settings)
	if err != nil {
		return nil, err
	}
	q.Q(req.GetProperties(), inputs)
	return &pulumirpc.CreateResponse{
		Id:         path.Join(inputs.Stack.OrgName, inputs.Stack.ProjectName, inputs.Stack.StackName),
		Properties: req.GetProperties(),
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Name() string {
	return "pulumiservice:index:DeploymentSettings"
}
