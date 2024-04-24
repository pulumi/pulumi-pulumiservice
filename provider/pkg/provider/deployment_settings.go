package provider

import (
	"context"
	"fmt"
	"path"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceDeploymentSettingsInput struct {
	pulumiapi.DeploymentSettings
	Stack pulumiapi.StackName
}

const FixMe = "<value is secret and must be replaced>"

func (ds *PulumiServiceDeploymentSettingsInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organization"] = resource.NewPropertyValue(ds.Stack.OrgName)
	pm["project"] = resource.NewPropertyValue(ds.Stack.ProjectName)
	pm["stack"] = resource.NewPropertyValue(ds.Stack.StackName)
	pm["agentPoolId"] = resource.NewPropertyValue(ds.AgentPoolId)

	if ds.SourceContext != nil {
		scMap := resource.PropertyMap{}
		if ds.SourceContext.Git != nil {
			gitPropertyMap := resource.PropertyMap{}
			if ds.SourceContext.Git.RepoURL != "" {
				gitPropertyMap["repoUrl"] = resource.NewPropertyValue(ds.SourceContext.Git.RepoURL)
			}
			if ds.SourceContext.Git.Commit != "" {
				gitPropertyMap["commit"] = resource.NewPropertyValue(ds.SourceContext.Git.Commit)
			}
			if ds.SourceContext.Git.Branch != "" {
				gitPropertyMap["branch"] = resource.NewPropertyValue(ds.SourceContext.Git.Branch)
			}
			if ds.SourceContext.Git.RepoDir != "" {
				gitPropertyMap["repoDir"] = resource.NewPropertyValue(ds.SourceContext.Git.RepoDir)
			}
			if ds.SourceContext.Git.GitAuth != nil {
				gitAuthPropertyMap := resource.PropertyMap{}
				if ds.SourceContext.Git.GitAuth.SSHAuth != nil {
					sshAuthPropertyMap := resource.PropertyMap{}
					if ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey.Value != "" {
						sshAuthPropertyMap["sshPrivateKey"] = resource.NewPropertyValue(FixMe)
					}
					if ds.SourceContext.Git.GitAuth.SSHAuth.Password.Value != "" {
						sshAuthPropertyMap["password"] = resource.NewPropertyValue(FixMe)
					}
					gitAuthPropertyMap["sshAuth"] = resource.PropertyValue{V: sshAuthPropertyMap}
				}
				if ds.SourceContext.Git.GitAuth.BasicAuth != nil {
					basicAuthPropertyMap := resource.PropertyMap{}
					if ds.SourceContext.Git.GitAuth.BasicAuth.UserName.Value != "" {
						basicAuthPropertyMap["username"] = resource.NewPropertyValue(ds.SourceContext.Git.GitAuth.BasicAuth.UserName.Value)
					}
					if ds.SourceContext.Git.GitAuth.BasicAuth.Password.Value != "" {
						basicAuthPropertyMap["password"] = resource.NewPropertyValue("fix me")
					}
					gitAuthPropertyMap["basicAuth"] = resource.NewPropertyValue(basicAuthPropertyMap)
				}
				gitPropertyMap["gitAuth"] = resource.PropertyValue{V: gitAuthPropertyMap}
			}
			scMap["git"] = resource.PropertyValue{V: gitPropertyMap}
		}
		pm["sourceContext"] = resource.PropertyValue{V: scMap}
	}

	if ds.OperationContext != nil {
		ocMap := resource.PropertyMap{}
		if ds.OperationContext.PreRunCommands != nil {
			ocMap["preRunCommands"] = resource.NewPropertyValue(ds.OperationContext.PreRunCommands)
		}
		if ds.OperationContext.EnvironmentVariables != nil {
			evMap := resource.PropertyMap{}
			for k, v := range ds.OperationContext.EnvironmentVariables {
				if v.Secret {
					evMap[resource.PropertyKey(k)] = resource.NewPropertyValue(FixMe)
				} else {
					evMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.Value)
				}
			}
			ocMap["environmentVariables"] = resource.PropertyValue{V: evMap}
		}
		if ds.OperationContext.Options != nil {
			optionsMap := resource.PropertyMap{}
			if ds.OperationContext.Options.Shell != "" {
				optionsMap["shell"] = resource.NewPropertyValue(ds.OperationContext.Options.Shell)
			}
			if ds.OperationContext.Options.SkipInstallDependencies {
				optionsMap["skipInstallDependencies"] = resource.NewPropertyValue(true)
			}
			if ds.OperationContext.Options.SkipIntermediateDeployments {
				optionsMap["skipIntermediateDeployments"] = resource.NewPropertyValue(true)
			}
			if ds.OperationContext.Options.DeleteAfterDestroy {
				optionsMap["deleteAfterDestroy"] = resource.NewPropertyValue(true)
			}
			ocMap["options"] = resource.PropertyValue{V: optionsMap}
		}
		if ds.OperationContext.OIDC != nil {
			if ds.OperationContext.OIDC.AWS != nil || ds.OperationContext.OIDC.GCP != nil || ds.OperationContext.OIDC.Azure != nil {
				oidcMap := resource.PropertyMap{}
				if ds.OperationContext.OIDC.AWS != nil {
					awsMap := resource.PropertyMap{}
					if ds.OperationContext.OIDC.AWS.RoleARN != "" {
						awsMap["roleARN"] = resource.NewPropertyValue(ds.OperationContext.OIDC.AWS.RoleARN)
					}
					if ds.OperationContext.OIDC.AWS.SessionName != "" {
						awsMap["sessionName"] = resource.NewPropertyValue(ds.OperationContext.OIDC.AWS.SessionName)
					}
					if ds.OperationContext.OIDC.AWS.PolicyARNs != nil {
						awsMap["policyARNs"] = resource.NewPropertyValue(ds.OperationContext.OIDC.AWS.PolicyARNs)
					}
					if ds.OperationContext.OIDC.AWS.Duration != "" {
						awsMap["duration"] = resource.NewPropertyValue(ds.OperationContext.OIDC.AWS.Duration)
					}
					oidcMap["aws"] = resource.PropertyValue{V: awsMap}
				}
				if ds.OperationContext.OIDC.GCP != nil {
					gcpMap := resource.PropertyMap{}
					if ds.OperationContext.OIDC.GCP.ProviderID != "" {
						gcpMap["providerId"] = resource.NewPropertyValue(ds.OperationContext.OIDC.GCP.ProviderID)
					}
					if ds.OperationContext.OIDC.GCP.ServiceAccount != "" {
						gcpMap["serviceAccount"] = resource.NewPropertyValue(ds.OperationContext.OIDC.GCP.ServiceAccount)
					}
					if ds.OperationContext.OIDC.GCP.Region != "" {
						gcpMap["region"] = resource.NewPropertyValue(ds.OperationContext.OIDC.GCP.Region)
					}
					if ds.OperationContext.OIDC.GCP.WorkloadPoolID != "" {
						gcpMap["workloadPoolId"] = resource.NewPropertyValue(ds.OperationContext.OIDC.GCP.WorkloadPoolID)
					}
					if ds.OperationContext.OIDC.GCP.ProjectID != "" {
						gcpMap["projectId"] = resource.NewPropertyValue(ds.OperationContext.OIDC.GCP.ProjectID)
					}
					if ds.OperationContext.OIDC.GCP.TokenLifetime != "" {
						gcpMap["tokenLifetime"] = resource.NewPropertyValue(ds.OperationContext.OIDC.GCP.TokenLifetime)
					}
					oidcMap["gcp"] = resource.PropertyValue{V: gcpMap}
				}
				if ds.OperationContext.OIDC.Azure != nil {
					azureMap := resource.PropertyMap{}
					if ds.OperationContext.OIDC.Azure.TenantID != "" {
						azureMap["tenantId"] = resource.NewPropertyValue(ds.OperationContext.OIDC.Azure.TenantID)
					}
					if ds.OperationContext.OIDC.Azure.ClientID != "" {
						azureMap["clientId"] = resource.NewPropertyValue(ds.OperationContext.OIDC.Azure.ClientID)
					}
					if ds.OperationContext.OIDC.Azure.SubscriptionID != "" {
						azureMap["subscriptionId"] = resource.NewPropertyValue(ds.OperationContext.OIDC.Azure.SubscriptionID)
					}
				}
				ocMap["oidc"] = resource.PropertyValue{V: oidcMap}
			}
		}
		pm["operationContext"] = resource.PropertyValue{V: ocMap}
	}

	if ds.GitHub != nil {
		githubMap := resource.PropertyMap{}
		githubMap["previewPullRequests"] = resource.NewPropertyValue(ds.GitHub.PreviewPullRequests)
		githubMap["deployCommits"] = resource.NewPropertyValue(ds.GitHub.DeployCommits)
		githubMap["pullRequestTemplate"] = resource.NewPropertyValue(ds.GitHub.PullRequestTemplate)
		if ds.GitHub.Repository != "" {
			githubMap["repository"] = resource.NewPropertyValue(ds.GitHub.Repository)
		}
		if len(ds.GitHub.Paths) > 0 {
			githubMap["paths"] = resource.NewPropertyValue(ds.GitHub.Paths)
		}
		pm["github"] = resource.PropertyValue{V: githubMap}
	}

	if ds.ExecutorContext != nil && ds.ExecutorContext.ExecutorImage != nil && ds.ExecutorContext.ExecutorImage.Reference != "" {
		ecMap := resource.PropertyMap{}
		ecMap["executorImage"] = resource.NewPropertyValue(ds.ExecutorContext.ExecutorImage)
		pm["executorContext"] = resource.PropertyValue{V: ecMap}
	}
	return pm
}

type PulumiServiceDeploymentSettingsResource struct{}

func (ds *PulumiServiceDeploymentSettingsResource) ToPulumiServiceDeploymentSettingsInput(inputMap resource.PropertyMap) PulumiServiceDeploymentSettingsInput {
	input := PulumiServiceDeploymentSettingsInput{}

	input.Stack.OrgName = getSecretOrStringValue(inputMap["organization"])
	input.Stack.ProjectName = getSecretOrStringValue(inputMap["project"])
	input.Stack.StackName = getSecretOrStringValue(inputMap["stack"])

	if inputMap["agentPoolId"].HasValue() && inputMap["agentPoolId"].IsString() {
		input.AgentPoolId = inputMap["agentPoolId"].StringValue()
	}

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
		ec.ExecutorImage = &apitype.DockerImage{
			Reference: getSecretOrStringValue(ecInput["executorImage"]),
		}
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
		github.PreviewPullRequests = githubInput["previewPullRequests"].BoolValue()
	}
	if githubInput["pullRequestTemplate"].HasValue() && githubInput["pullRequestTemplate"].IsBool() {
		github.PullRequestTemplate = githubInput["pullRequestTemplate"].BoolValue()
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

		if oInput["skipIntermediateDeployments"].HasValue() && oInput["skipIntermediateDeployments"].IsBool() {
			o.SkipIntermediateDeployments = oInput["skipIntermediateDeployments"].BoolValue()
		}

		if oInput["Shell"].HasValue() && oInput["Shell"].IsString() {
			o.Shell = oInput["Shell"].StringValue()
		}

		if oInput["deleteAfterDestroy"].HasValue() && oInput["deleteAfterDestroy"].IsBool() {
			o.DeleteAfterDestroy = oInput["deleteAfterDestroy"].BoolValue()
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

func (ds *PulumiServiceDeploymentSettingsResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return considerAllChangesReplaces(req)
}

func (ds *PulumiServiceDeploymentSettingsResource) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
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

func (ds *PulumiServiceDeploymentSettingsResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	var stack pulumiapi.StackName
	if err := stack.FromID(req.Id); err != nil {
		return nil, err
	}
	settings, err := GetClient[pulumiapi.DeploymentSettingsClient](ctx).GetDeploymentSettings(ctx, stack)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		// Empty response causes the resource to be deleted from the state.
		return &pulumirpc.ReadResponse{Id: "", Properties: nil}, nil
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

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     properties,
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputs := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap)
	if err != nil {
		return nil, err
	}
	err = GetClient[pulumiapi.DeploymentSettingsClient](ctx).DeleteDeploymentSettings(ctx, inputs.Stack)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true})
	if err != nil {
		return nil, err
	}
	inputs := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap)
	settings := inputs.DeploymentSettings
	err = GetClient[pulumiapi.DeploymentSettingsClient](ctx).CreateDeploymentSettings(ctx, inputs.Stack, settings)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{
		Id:         path.Join(inputs.Stack.OrgName, inputs.Stack.ProjectName, inputs.Stack.StackName),
		Properties: req.GetProperties(),
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Update(context.Context, *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// For simplicity, all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (ds *PulumiServiceDeploymentSettingsResource) Name() string {
	return "pulumiservice:index:DeploymentSettings"
}

func considerAllChangesReplaces(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	for k, v := range dd {
		v.Kind = v.Kind.AsReplace()
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		DetailedDiff:        detailedDiffs,
		DeleteBeforeReplace: true,
		HasDetailedDiff:     true,
	}, nil
}
