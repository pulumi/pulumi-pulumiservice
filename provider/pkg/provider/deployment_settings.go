package provider

import (
	"context"
	"fmt"
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

func (ds *PulumiServiceDeploymentSettingsInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organization"] = resource.NewPropertyValue(ds.Stack.OrgName)
	pm["project"] = resource.NewPropertyValue(ds.Stack.ProjectName)
	pm["stack"] = resource.NewPropertyValue(ds.Stack.StackName)
	return pm
}

type PulumiServiceDeploymentSettingsResource struct {
	client *pulumiapi.Client
}

func (ds *PulumiServiceDeploymentSettingsResource) ToPulumiServiceDeploymentSettingsInput(inputMap resource.PropertyMap) PulumiServiceDeploymentSettingsInput {
	input := PulumiServiceDeploymentSettingsInput{}

	input.Stack.OrgName = getSecretOrStringValue(inputMap["organization"])
	input.Stack.ProjectName = getSecretOrStringValue(inputMap["project"])
	input.Stack.StackName = getSecretOrStringValue(inputMap["stack"])

	input.ExecutorContext = toExecutorContext(inputMap)
	input.GitHub = toGitHubConfig(inputMap)
	input.SourceContext = toSourceContext(inputMap)
	input.OperationContext = toOperationContext(inputMap)

	return input
}

func toExecutorContext(inputMap resource.PropertyMap) *apitype.ExecutorContext {
	if !inputMap["executorContext"].HasValue() || !inputMap["executorContext"].IsObject() {
		return nil
	}

	ecInput := inputMap["executorContext"].ObjectValue()
	var ec apitype.ExecutorContext

	if ecInput["executorImage"].HasValue() {
		ec.ExecutorImage = getSecretOrStringValue(ecInput["executorImage"])
	}

	return &ec
}

func toGitHubConfig(inputMap resource.PropertyMap) *pulumiapi.GitHubConfiguration {
	if !inputMap["github"].HasValue() || !inputMap["github"].IsObject() {
		return nil
	}

	githubInput := inputMap["github"].ObjectValue()
	var github pulumiapi.GitHubConfiguration

	if githubInput["repository"].HasValue() {
		github.Repository = getSecretOrStringValue(githubInput["repository"])
	}

	if githubInput["deployCommits"].HasValue() && githubInput["deployCommits"].IsBool() {
		github.DeployCommits = githubInput["deployCommits"].BoolValue()
	}
	if githubInput["previewPullRequests"].HasValue() && githubInput["previewPullRequests"].IsBool() {
		github.PreviewPullRequests = githubInput["deployPullRequests"].BoolValue()
	}
	if githubInput["paths"].HasValue() && githubInput["paths"].IsArray() {
		pathsInput := githubInput["paths"].ArrayValue()
		paths := make([]string, len(pathsInput))

		for i, v := range pathsInput {
			paths[i] = getSecretOrStringValue(v)
		}

		github.Paths = paths
	}

	return &github
}

func toSourceContext(inputMap resource.PropertyMap) *apitype.SourceContext {
	if !inputMap["sourceContext"].HasValue() || !inputMap["sourceContext"].IsObject() {
		return nil
	}

	scInput := inputMap["sourceContext"].ObjectValue()
	var sc apitype.SourceContext

	if scInput["git"].HasValue() && scInput["git"].IsObject() {
		gitInput := scInput["git"].ObjectValue()
		var g apitype.SourceContextGit

		if gitInput["repoUrl"].HasValue() {
			g.RepoURL = getSecretOrStringValue(gitInput["repoUrl"])
		}
		if gitInput["branch"].HasValue() {
			g.Branch = getSecretOrStringValue(gitInput["branch"])
		}
		if gitInput["repoDir"].HasValue() {
			g.RepoDir = getSecretOrStringValue(gitInput["repoDir"])
		}

		if gitInput["gitAuth"].HasValue() && gitInput["gitAuth"].IsObject() {
			authInput := gitInput["gitAuth"].ObjectValue()
			var a apitype.GitAuthConfig

			if authInput["sshAuth"].HasValue() && authInput["sshAuth"].IsObject() {
				sshInput := authInput["sshAuth"].ObjectValue()
				var s apitype.SSHAuth

				if sshInput["sshPrivateKey"].HasValue() {
					s.SSHPrivateKey = apitype.SecretValue{
						Secret: true,
						Value:  getSecretOrStringValue(sshInput["sshPrivateKey"]),
					}
				}
				if sshInput["password"].HasValue() {
					s.Password = &apitype.SecretValue{
						Secret: true,
						Value:  getSecretOrStringValue(sshInput["password"]),
					}
				}

				a.SSHAuth = &s
			}

			if authInput["basicAuth"].HasValue() && authInput["basicAuth"].IsObject() {
				basicInput := authInput["basicAuth"].ObjectValue()
				var b apitype.BasicAuth

				if basicInput["username"].HasValue() {
					b.UserName = apitype.SecretValue{
						Value:  getSecretOrStringValue(basicInput["username"]),
						Secret: false,
					}
				}
				if basicInput["password"].HasValue() {
					b.Password = apitype.SecretValue{
						Value:  getSecretOrStringValue(basicInput["password"]),
						Secret: true,
					}
				}

				a.BasicAuth = &b
			}

			g.GitAuth = &a
		}

		sc.Git = &g
	}

	return &sc
}

func toOperationContext(inputMap resource.PropertyMap) *pulumiapi.OperationContext {
	if !inputMap["operationContext"].HasValue() || !inputMap["operationContext"].IsObject() {
		return nil
	}

	ocInput := inputMap["operationContext"].ObjectValue()
	var oc pulumiapi.OperationContext

	if ocInput["environmentVariables"].HasValue() && ocInput["environmentVariables"].IsObject() {
		ev := map[string]apitype.SecretValue{}
		evInput := ocInput["environmentVariables"].ObjectValue()

		for k, v := range evInput {
			if v.IsSecret() {
				ev[string(k)] = apitype.SecretValue{Secret: true, Value: v.SecretValue().Element.StringValue()}
			} else {
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

	if ocInput["oidc"].HasValue() && ocInput["oidc"].IsObject() {
		oidcInput := ocInput["oidc"].ObjectValue()
		var oidc pulumiapi.OIDCConfiguration

		if oidcInput["aws"].HasValue() && oidcInput["aws"].IsObject() {
			awsInput := oidcInput["aws"].ObjectValue()
			var aws pulumiapi.AWSOIDCConfiguration

			if awsInput["roleARN"].HasValue() {
				aws.RoleARN = getSecretOrStringValue(awsInput["roleARN"])
			}
			if awsInput["duration"].HasValue() {
				aws.Duration = getSecretOrStringValue(awsInput["duration"])
			}
			if awsInput["sessionName"].HasValue() {
				aws.SessionName = getSecretOrStringValue(awsInput["sessionName"])
			}
			if awsInput["policyARNs"].HasValue() && awsInput["policyARNs"].IsArray() {
				policyARNsInput := awsInput["policyARNs"].ArrayValue()
				policyARNs := make([]string, len(policyARNsInput))

				for i, v := range policyARNsInput {
					policyARNs[i] = getSecretOrStringValue(v)
				}

				aws.PolicyARNs = policyARNs
			}

			oidc.AWS = &aws
		}

		if oidcInput["gcp"].HasValue() && oidcInput["gcp"].IsObject() {
			gcpInput := oidcInput["gcp"].ObjectValue()
			var gcp pulumiapi.GCPOIDCConfiguration

			if gcpInput["projectId"].HasValue() {
				gcp.ProjectID = getSecretOrStringValue(gcpInput["projectId"])
			}
			if gcpInput["region"].HasValue() {
				gcp.Region = getSecretOrStringValue(gcpInput["region"])
			}
			if gcpInput["workloadPoolId"].HasValue() {
				gcp.WorkloadPoolID = getSecretOrStringValue(gcpInput["workloadPoolId"])
			}
			if gcpInput["providerId"].HasValue() {
				gcp.ProviderID = getSecretOrStringValue(gcpInput["providerId"])
			}
			if gcpInput["serviceAccount"].HasValue() {
				gcp.ServiceAccount = getSecretOrStringValue(gcpInput["serviceAccount"])
			}
			if gcpInput["tokenLifetime"].HasValue() {
				gcp.TokenLifetime = getSecretOrStringValue(gcpInput["tokenLifetime"])
			}

			oidc.GCP = &gcp
		}

		if oidcInput["azure"].HasValue() && oidcInput["azure"].IsObject() {
			azureInput := oidcInput["azure"].ObjectValue()
			var azure pulumiapi.AzureOIDCConfiguration

			if azureInput["tenantId"].HasValue() {
				azure.TenantID = getSecretOrStringValue(azureInput["tenantId"])
			}
			if azureInput["clientId"].HasValue() {
				azure.ClientID = getSecretOrStringValue(azureInput["clientId"])
			}
			if azureInput["subscriptionId"].HasValue() {
				azure.SubscriptionID = getSecretOrStringValue(azureInput["subscriptionId"])
			}

			oidc.Azure = &azure
		}

		oc.OIDC = &oidc
	}

	return &oc
}

func getSecretOrStringValue(prop resource.PropertyValue) string {
	switch prop.V.(type) {
	case *resource.Secret:
		return prop.SecretValue().Element.StringValue()
	default:
		return prop.StringValue()
	}
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

	diffs := olds.Diff(news)
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
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "stack"} {
		if !news[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Configure(config PulumiServiceConfig) {}

func (ds *PulumiServiceDeploymentSettingsResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	q.Q(req.Id)
	var stack pulumiapi.StackName
	if err := stack.FromID(req.Id); err != nil {
		return nil, err
	}
	settings, err := ds.client.GetDeploymentSettings(ctx, stack)
	if err != nil {
		return nil, err
	}

	dsInput := PulumiServiceDeploymentSettingsInput{
		Stack:              stack,
		DeploymentSettings: *settings,
	}

	properties, err := plugin.MarshalProperties(
		dsInput.ToPropertyMap(),
		plugin.MarshalOptions{},
	)

	if err != nil {
		return nil, err
	}

	inputProperties, err := plugin.MarshalProperties(
		inputs,
		plugin.MarshalOptions{},
	)

	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     inputProperties,
	}, nil
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
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true})
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
	return &pulumirpc.CreateResponse{
		Id:         path.Join(inputs.Stack.OrgName, inputs.Stack.ProjectName, inputs.Stack.StackName),
		Properties: req.GetProperties(),
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// For simplicity, all updates are destructive, so we just call Create.
	return &pulumirpc.UpdateResponse{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Name() string {
	return "pulumiservice:index:DeploymentSettings"
}
