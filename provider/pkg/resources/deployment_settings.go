package resources

import (
	"context"
	"fmt"
	"path"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceDeploymentSettingsInput struct {
	pulumiapi.DeploymentSettings
	Stack pulumiapi.StackIdentifier
}

// plaintextInputSettings are the latest inputs of the resource, containing plaintext values wrapped in Secrets
// currentStateCipherSettings are the latest outputs/properties of the resource, containing ciphertext strings of secret values
// isInput is a flag that selects whether to generating an input PropertyMap that contains plaintext (true) or an output PropertyMap that contains ciphertext (false)
func (ds *PulumiServiceDeploymentSettingsInput) ToPropertyMap(plaintextInputSettings *pulumiapi.DeploymentSettings, currentStateCipherSettings *pulumiapi.DeploymentSettings, isInput bool) resource.PropertyMap {
	// Below flags are used throughout this method and direct the serialization of twin value secrets
	// Twin value secrets are values whose plaintext cannot be retrieved from the API, thus forcing the development of this fairly complex system
	// When plaintextInputSettings is passed in, but currentStateCipherSettings is not, that means the resource is being created or updated
	createMode := plaintextInputSettings != nil && currentStateCipherSettings == nil
	// When both plaintextInputSettings and currentStateCipherSettings are passed in, that means an existing resource is being refreshed, and it's necessary to merge values
	// In case we are merging, but some of the properties don't previously exist, we will merge with empty value, setting plaintext to be empty string
	mergeMode := plaintextInputSettings != nil && currentStateCipherSettings != nil
	// If neither one is passed in, we are importing an existing resource into the state

	pm := resource.PropertyMap{}
	pm["organization"] = resource.NewPropertyValue(ds.Stack.OrgName)
	pm["project"] = resource.NewPropertyValue(ds.Stack.ProjectName)
	pm["stack"] = resource.NewPropertyValue(ds.Stack.StackName)

	if ds.AgentPoolId != "" {
		pm["agentPoolId"] = resource.NewPropertyValue(ds.AgentPoolId)
	}

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
						if mergeMode {
							var plaintextValue *pulumiapi.SecretValue
							var currentCipherValue *pulumiapi.SecretValue
							if currentStateCipherSettings.SourceContext != nil &&
								currentStateCipherSettings.SourceContext.Git != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth.SSHAuth != nil {
								plaintextValue = &plaintextInputSettings.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey
								currentCipherValue = &currentStateCipherSettings.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey
							}
							util.MergeSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey, plaintextValue, currentCipherValue, isInput)
						} else if createMode {
							util.CreateSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey,
								plaintextInputSettings.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey, isInput)
						} else {
							util.ImportSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey, isInput)
						}
					}
					if ds.SourceContext.Git.GitAuth.SSHAuth.Password != nil && ds.SourceContext.Git.GitAuth.SSHAuth.Password.Value != "" {
						if mergeMode {
							var plaintextValue *pulumiapi.SecretValue
							var currentCipherValue *pulumiapi.SecretValue
							if currentStateCipherSettings.SourceContext != nil &&
								currentStateCipherSettings.SourceContext.Git != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth.SSHAuth != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth.SSHAuth.Password != nil {
								plaintextValue = plaintextInputSettings.SourceContext.Git.GitAuth.SSHAuth.Password
								currentCipherValue = currentStateCipherSettings.SourceContext.Git.GitAuth.SSHAuth.Password
							}
							util.MergeSecretValue(sshAuthPropertyMap, "password", *ds.SourceContext.Git.GitAuth.SSHAuth.Password, plaintextValue, currentCipherValue, isInput)
						} else if createMode {
							util.CreateSecretValue(sshAuthPropertyMap, "password", *ds.SourceContext.Git.GitAuth.SSHAuth.Password,
								*plaintextInputSettings.SourceContext.Git.GitAuth.SSHAuth.Password, isInput)
						} else {
							util.ImportSecretValue(sshAuthPropertyMap, "password", *ds.SourceContext.Git.GitAuth.SSHAuth.Password, isInput)
						}
					}
					gitAuthPropertyMap["sshAuth"] = resource.PropertyValue{V: sshAuthPropertyMap}
				}
				if ds.SourceContext.Git.GitAuth.BasicAuth != nil {
					basicAuthPropertyMap := resource.PropertyMap{}
					if ds.SourceContext.Git.GitAuth.BasicAuth.UserName.Value != "" {
						basicAuthPropertyMap["username"] = resource.NewPropertyValue(ds.SourceContext.Git.GitAuth.BasicAuth.UserName.Value)
					}
					if ds.SourceContext.Git.GitAuth.BasicAuth.Password.Value != "" {
						if mergeMode {
							var plaintextValue *pulumiapi.SecretValue
							var currentCipherValue *pulumiapi.SecretValue
							if currentStateCipherSettings.SourceContext != nil &&
								currentStateCipherSettings.SourceContext.Git != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth.BasicAuth != nil {
								plaintextValue = &plaintextInputSettings.SourceContext.Git.GitAuth.BasicAuth.Password
								currentCipherValue = &currentStateCipherSettings.SourceContext.Git.GitAuth.BasicAuth.Password
							}
							util.MergeSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password, plaintextValue, currentCipherValue, isInput)
						} else if createMode {
							util.CreateSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password,
								plaintextInputSettings.SourceContext.Git.GitAuth.BasicAuth.Password, isInput)
						} else {
							util.ImportSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password, isInput)
						}
					}
					gitAuthPropertyMap["basicAuth"] = resource.PropertyValue{V: basicAuthPropertyMap}
				}
				gitPropertyMap["gitAuth"] = resource.PropertyValue{V: gitAuthPropertyMap}
			}
			scMap["git"] = resource.PropertyValue{V: gitPropertyMap}
		}
		if ds.SourceContext.Template != nil {
			templatePropertyMap := resource.PropertyMap{}
			if ds.SourceContext.Template.SourceURL != "" {
				templatePropertyMap["sourceUrl"] = resource.NewPropertyValue(ds.SourceContext.Template.SourceURL)
			}
			scMap["template"] = resource.PropertyValue{V: templatePropertyMap}
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
					if mergeMode {
						var plaintextValue pulumiapi.SecretValue
						var currentCipherValue pulumiapi.SecretValue
						if currentStateCipherSettings.OperationContext != nil {
							plaintextValue = plaintextInputSettings.OperationContext.EnvironmentVariables[k]
							currentCipherValue = currentStateCipherSettings.OperationContext.EnvironmentVariables[k]
						}
						util.MergeSecretValue(evMap, k, v, &plaintextValue, &currentCipherValue, isInput)
					} else if createMode {
						util.CreateSecretValue(evMap, k, v,
							plaintextInputSettings.OperationContext.EnvironmentVariables[k], isInput)
					} else {
						util.ImportSecretValue(evMap, k, v, isInput)
					}
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
					oidcMap["azure"] = resource.PropertyValue{V: azureMap}
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
		ecMap["executorImage"] = resource.NewPropertyValue(ds.ExecutorContext.ExecutorImage.Reference)
		pm["executorContext"] = resource.PropertyValue{V: ecMap}
	}

	if ds.CacheOptions != nil {
		coMap := resource.PropertyMap{}
		coMap["enable"] = resource.NewPropertyValue(ds.CacheOptions.Enable)
		pm["cacheOptions"] = resource.PropertyValue{V: coMap}
	}

	return pm
}

type PulumiServiceDeploymentSettingsResource struct {
	Client pulumiapi.DeploymentSettingsClient
}

func (ds *PulumiServiceDeploymentSettingsResource) ToPulumiServiceDeploymentSettingsInput(inputMap resource.PropertyMap) PulumiServiceDeploymentSettingsInput {
	input := PulumiServiceDeploymentSettingsInput{}

	input.Stack.OrgName = util.GetSecretOrStringValue(inputMap["organization"])
	input.Stack.ProjectName = util.GetSecretOrStringValue(inputMap["project"])
	input.Stack.StackName = util.GetSecretOrStringValue(inputMap["stack"])

	if inputMap["agentPoolId"].HasValue() {
		input.AgentPoolId = util.GetSecretOrStringValue(inputMap["agentPoolId"])
	}

	input.ExecutorContext = toExecutorContext(inputMap)
	input.GitHub = toGitHubConfig(inputMap)
	input.SourceContext = toSourceContext(inputMap)
	input.OperationContext = toOperationContext(inputMap)
	input.CacheOptions = toCacheOptions(inputMap)

	return input
}

func toExecutorContext(inputMap resource.PropertyMap) *apitype.ExecutorContext {
	if !inputMap["executorContext"].HasValue() {
		return nil
	}

	ecInput := util.GetSecretOrObjectValue(inputMap["executorContext"])
	var ec apitype.ExecutorContext

	if ecInput["executorImage"].HasValue() {
		ec.ExecutorImage = &apitype.DockerImage{
			Reference: util.GetSecretOrStringValue(ecInput["executorImage"]),
		}
	}

	return &ec
}

func toGitHubConfig(inputMap resource.PropertyMap) *pulumiapi.GitHubConfiguration {
	if !inputMap["github"].HasValue() {
		return nil
	}

	githubInput := util.GetSecretOrObjectValue(inputMap["github"])
	var github pulumiapi.GitHubConfiguration

	if githubInput["repository"].HasValue() {
		github.Repository = util.GetSecretOrStringValue(githubInput["repository"])
	}

	if githubInput["deployCommits"].HasValue() {
		github.DeployCommits = util.GetSecretOrBoolValue(githubInput["deployCommits"])
	}
	if githubInput["previewPullRequests"].HasValue() {
		github.PreviewPullRequests = util.GetSecretOrBoolValue(githubInput["previewPullRequests"])
	}
	if githubInput["pullRequestTemplate"].HasValue() {
		github.PullRequestTemplate = util.GetSecretOrBoolValue(githubInput["pullRequestTemplate"])
	}
	if githubInput["paths"].HasValue() {
		pathsInput := util.GetSecretOrArrayValue(githubInput["paths"])
		paths := make([]string, len(pathsInput))

		for i, v := range pathsInput {
			paths[i] = util.GetSecretOrStringValue(v)
		}

		github.Paths = paths
	}

	return &github
}

func toSourceContext(inputMap resource.PropertyMap) *pulumiapi.SourceContext {
	if !inputMap["sourceContext"].HasValue() {
		return nil
	}

	scInput := util.GetSecretOrObjectValue(inputMap["sourceContext"])
	cascadeSecret := inputMap["sourceContext"].IsSecret()
	var sc pulumiapi.SourceContext

	if scInput["git"].HasValue() {
		gitInput := util.GetSecretOrObjectValue(scInput["git"])
		cascadeSecret = cascadeSecret || scInput["git"].IsSecret()
		var g pulumiapi.SourceContextGit

		if gitInput["repoUrl"].HasValue() {
			g.RepoURL = util.GetSecretOrStringValue(gitInput["repoUrl"])
		}
		if gitInput["branch"].HasValue() {
			g.Branch = util.GetSecretOrStringValue(gitInput["branch"])
		}
		if gitInput["commit"].HasValue() {
			g.Commit = util.GetSecretOrStringValue(gitInput["commit"])
		}
		if gitInput["repoDir"].HasValue() {
			g.RepoDir = util.GetSecretOrStringValue(gitInput["repoDir"])
		}

		if gitInput["gitAuth"].HasValue() {
			authInput := util.GetSecretOrObjectValue(gitInput["gitAuth"])
			cascadeSecret = cascadeSecret || gitInput["gitAuth"].IsSecret()
			var a pulumiapi.GitAuthConfig

			if authInput["sshAuth"].HasValue() {
				sshInput := util.GetSecretOrObjectValue(authInput["sshAuth"])
				var s pulumiapi.SSHAuth

				if sshInput["sshPrivateKey"].HasValue() || sshInput["sshPrivateKeyCipher"].HasValue() {
					s.SSHPrivateKey = pulumiapi.SecretValue{
						Secret: true,
						Value:  util.GetSecretOrStringValue(sshInput["sshPrivateKey"]),
					}
				}
				if sshInput["password"].HasValue() || sshInput["passwordCipher"].HasValue() {
					s.Password = &pulumiapi.SecretValue{
						Secret: true,
						Value:  util.GetSecretOrStringValue(sshInput["password"]),
					}
				}

				a.SSHAuth = &s
			}

			if authInput["basicAuth"].HasValue() {
				basicInput := util.GetSecretOrObjectValue(authInput["basicAuth"])
				cascadeSecret = cascadeSecret || authInput["basicAuth"].IsSecret()
				var b pulumiapi.BasicAuth

				if basicInput["username"].HasValue() {
					b.UserName = pulumiapi.SecretValue{
						Value:  util.GetSecretOrStringValue(basicInput["username"]),
						Secret: cascadeSecret || basicInput["username"].IsSecret(),
					}
				}
				if basicInput["password"].HasValue() || basicInput["passwordCipher"].HasValue() {
					b.Password = pulumiapi.SecretValue{
						Value:  util.GetSecretOrStringValue(basicInput["password"]),
						Secret: true,
					}
				}

				a.BasicAuth = &b
			}

			g.GitAuth = &a
		}

		sc.Git = &g
	}

	if scInput["template"].HasValue() {
		templateInput := util.GetSecretOrObjectValue(scInput["template"])
		var t pulumiapi.SourceContextTemplate

		if templateInput["sourceUrl"].HasValue() {
			t.SourceURL = util.GetSecretOrStringValue(templateInput["sourceUrl"])
		}

		sc.Template = &t
	}

	return &sc
}

func toOperationContext(inputMap resource.PropertyMap) *pulumiapi.OperationContext {
	if !inputMap["operationContext"].HasValue() {
		return nil
	}

	ocInput := util.GetSecretOrObjectValue(inputMap["operationContext"])
	cascadeSecret := inputMap["operationContext"].IsSecret()
	var oc pulumiapi.OperationContext

	if ocInput["environmentVariables"].HasValue() {
		ev := map[string]pulumiapi.SecretValue{}
		evInput := util.GetSecretOrObjectValue(ocInput["environmentVariables"])
		cascadeSecret = cascadeSecret || ocInput["environmentVariables"].IsSecret()

		for k, v := range evInput {
			value := util.GetSecretOrStringValue(v)
			ev[string(k)] = pulumiapi.SecretValue{Secret: v.IsSecret() || cascadeSecret, Value: value}
		}

		oc.EnvironmentVariables = ev
	}

	if ocInput["preRunCommands"].HasValue() {
		pcInput := util.GetSecretOrArrayValue(ocInput["preRunCommands"])
		pc := make([]string, len(pcInput))

		for i, v := range pcInput {
			pc[i] = util.GetSecretOrStringValue(v)
		}

		oc.PreRunCommands = pc
	}

	if ocInput["options"].HasValue() {
		oInput := util.GetSecretOrObjectValue(ocInput["options"])
		var o pulumiapi.OperationContextOptions

		if oInput["skipInstallDependencies"].HasValue() {
			o.SkipInstallDependencies = util.GetSecretOrBoolValue(oInput["skipInstallDependencies"])
		}

		if oInput["skipIntermediateDeployments"].HasValue() {
			o.SkipIntermediateDeployments = util.GetSecretOrBoolValue(oInput["skipIntermediateDeployments"])
		}

		if oInput["Shell"].HasValue() {
			o.Shell = util.GetSecretOrStringValue(oInput["Shell"])
		}

		if oInput["deleteAfterDestroy"].HasValue() {
			o.DeleteAfterDestroy = util.GetSecretOrBoolValue(oInput["deleteAfterDestroy"])
		}

		oc.Options = &o
	}

	if ocInput["oidc"].HasValue() {
		oidcInput := util.GetSecretOrObjectValue(ocInput["oidc"])
		var oidc pulumiapi.OIDCConfiguration

		if oidcInput["aws"].HasValue() {
			awsInput := util.GetSecretOrObjectValue(oidcInput["aws"])
			var aws pulumiapi.AWSOIDCConfiguration

			if awsInput["roleARN"].HasValue() {
				aws.RoleARN = util.GetSecretOrStringValue(awsInput["roleARN"])
			}
			if awsInput["duration"].HasValue() {
				aws.Duration = util.GetSecretOrStringValue(awsInput["duration"])
			}
			if awsInput["sessionName"].HasValue() {
				aws.SessionName = util.GetSecretOrStringValue(awsInput["sessionName"])
			}
			if awsInput["policyARNs"].HasValue() {
				policyARNsInput := util.GetSecretOrArrayValue(awsInput["policyARNs"])
				policyARNs := make([]string, len(policyARNsInput))

				for i, v := range policyARNsInput {
					policyARNs[i] = util.GetSecretOrStringValue(v)
				}

				aws.PolicyARNs = policyARNs
			}

			oidc.AWS = &aws
		}

		if oidcInput["gcp"].HasValue() {
			gcpInput := util.GetSecretOrObjectValue(oidcInput["gcp"])
			var gcp pulumiapi.GCPOIDCConfiguration

			if gcpInput["projectId"].HasValue() {
				gcp.ProjectID = util.GetSecretOrStringValue(gcpInput["projectId"])
			}
			if gcpInput["region"].HasValue() {
				gcp.Region = util.GetSecretOrStringValue(gcpInput["region"])
			}
			if gcpInput["workloadPoolId"].HasValue() {
				gcp.WorkloadPoolID = util.GetSecretOrStringValue(gcpInput["workloadPoolId"])
			}
			if gcpInput["providerId"].HasValue() {
				gcp.ProviderID = util.GetSecretOrStringValue(gcpInput["providerId"])
			}
			if gcpInput["serviceAccount"].HasValue() {
				gcp.ServiceAccount = util.GetSecretOrStringValue(gcpInput["serviceAccount"])
			}
			if gcpInput["tokenLifetime"].HasValue() {
				gcp.TokenLifetime = util.GetSecretOrStringValue(gcpInput["tokenLifetime"])
			}

			oidc.GCP = &gcp
		}

		if oidcInput["azure"].HasValue() {
			azureInput := util.GetSecretOrObjectValue(oidcInput["azure"])
			var azure pulumiapi.AzureOIDCConfiguration

			if azureInput["tenantId"].HasValue() {
				azure.TenantID = util.GetSecretOrStringValue(azureInput["tenantId"])
			}
			if azureInput["clientId"].HasValue() {
				azure.ClientID = util.GetSecretOrStringValue(azureInput["clientId"])
			}
			if azureInput["subscriptionId"].HasValue() {
				azure.SubscriptionID = util.GetSecretOrStringValue(azureInput["subscriptionId"])
			}

			oidc.Azure = &azure
		}

		oc.OIDC = &oidc
	}

	return &oc
}

func toCacheOptions(inputMap resource.PropertyMap) *pulumiapi.CacheOptions {
	if !inputMap["cacheOptions"].HasValue() {
		return nil
	}

	coInput := util.GetSecretOrObjectValue(inputMap["cacheOptions"])
	var co pulumiapi.CacheOptions

	if coInput["enable"].HasValue() {
		co.Enable = util.GetSecretOrBoolValue(coInput["enable"])
	}

	return &co
}

func (ds *PulumiServiceDeploymentSettingsResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
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

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		Replaces:            replaces,
		DetailedDiff:        detailedDiffs,
		HasDetailedDiff:     true,
		DeleteBeforeReplace: true,
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(req.GetNews(), util.KeepSecretsUnmarshal)
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

	// Normalizing duration input
	if news["operationContext"].HasValue() {
		operationContext := util.GetSecretOrObjectValue(news["operationContext"])
		if operationContext["oidc"].HasValue() {
			oidc := util.GetSecretOrObjectValue(operationContext["oidc"])
			if oidc["aws"].HasValue() {
				aws := util.GetSecretOrObjectValue(oidc["aws"])
				if aws["duration"].HasValue() {
					durationString := util.GetSecretOrStringValue(aws["duration"])
					normalized, err := normalizeDurationString(durationString)
					if err != nil {
						failures = append(failures, &pulumirpc.CheckFailure{
							Reason:   fmt.Sprintf("Failed to normalize duration string due to error: %s", err.Error()),
							Property: string("operationContext.oidc.aws.duration"),
						})
					} else {
						if aws["duration"].IsSecret() {
							aws["duration"] = resource.MakeSecret(resource.NewStringProperty(*normalized))
						} else {
							aws["duration"] = resource.NewStringProperty(*normalized)
						}
					}
				}
			}
		}
	}

	checkedNews, err := plugin.MarshalProperties(news, util.StandardMarshal)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: checkedNews, Failures: failures}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	stack, err := pulumiapi.NewStackIdentifier(req.GetId())
	if err != nil {
		return nil, err
	}
	settings, err := ds.Client.GetDeploymentSettings(ctx, stack)
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

	var plaintextSettings *pulumiapi.DeploymentSettings
	var ciphertextSettings *pulumiapi.DeploymentSettings
	propertyMap, err := plugin.UnmarshalProperties(req.GetProperties(), util.KeepSecretsUnmarshal)
	if err != nil {
		return nil, err
	}
	inputMap, err := plugin.UnmarshalProperties(req.GetInputs(), util.KeepSecretsUnmarshal)
	if err != nil {
		return nil, err
	}
	if propertyMap["stack"].HasValue() {
		tempPlain := ds.ToPulumiServiceDeploymentSettingsInput(inputMap)
		plaintextSettings = &tempPlain.DeploymentSettings
		tempCipher := ds.ToPulumiServiceDeploymentSettingsInput(propertyMap)
		ciphertextSettings = &tempCipher.DeploymentSettings
	}

	properties, err := plugin.MarshalProperties(
		dsInput.ToPropertyMap(plaintextSettings, ciphertextSettings, false), util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	inputs, err := plugin.MarshalProperties(
		dsInput.ToPropertyMap(plaintextSettings, ciphertextSettings, true), util.StandardMarshal,
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

func (ds *PulumiServiceDeploymentSettingsResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	stack, err := pulumiapi.NewStackIdentifier(req.GetId())
	if err != nil {
		return nil, err
	}

	err = ds.Client.DeleteDeploymentSettings(ctx, stack)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(), util.KeepSecretsUnmarshal)
	if err != nil {
		return nil, err
	}

	input := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap)
	settings := input.DeploymentSettings
	response, err := ds.Client.CreateDeploymentSettings(ctx, input.Stack, settings)
	if err != nil {
		return nil, err
	}

	responseInput := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: *response,
		Stack:              input.Stack,
	}

	outputProperties, err := plugin.MarshalProperties(
		responseInput.ToPropertyMap(&input.DeploymentSettings, nil, false), util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.Stack.OrgName, input.Stack.ProjectName, input.Stack.StackName),
		Properties: outputProperties,
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	inputsMap, err := plugin.UnmarshalProperties(req.GetNews(), util.KeepSecretsUnmarshal)
	if err != nil {
		return nil, err
	}

	input := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap)
	settings := input.DeploymentSettings
	response, err := ds.Client.UpdateDeploymentSettings(ctx, input.Stack, settings)
	if err != nil {
		return nil, err
	}

	responseInput := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: *response,
		Stack:              input.Stack,
	}

	outputProperties, err := plugin.MarshalProperties(
		responseInput.ToPropertyMap(&input.DeploymentSettings, nil, false), util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Name() string {
	return "pulumiservice:index:DeploymentSettings"
}

func normalizeDurationString(input string) (*string, error) {
	if input == "" {
		return nil, fmt.Errorf("empty value provided for duration string")
	}
	duration, err := time.ParseDuration(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration string `%s` due to error: %w", input, err)
	}
	result := duration.String()

	return &result, nil
}
