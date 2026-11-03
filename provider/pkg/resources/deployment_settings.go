package resources

import (
	"context"
	"fmt"
	"path"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	gcOrganization = "organization"
	gcStack        = "stack"
	gcAWS          = "aws"
)

// Property names along the paths to the nested credentials.
const (
	gcSourceContext   = "sourceContext"
	gcGit             = "git"
	gcGitAuth         = "gitAuth"
	gcSSHAuth         = "sshAuth"
	gcSSHPrivateKey   = "sshPrivateKey"
	gcBasicAuth       = "basicAuth"
	gcUsername        = "username"
	gcPassword        = "password"
	gcExecutorContext = "executorContext"
	gcCredentials     = "credentials"
)

type PulumiServiceDeploymentSettingsInput struct {
	pulumiapi.DeploymentSettings
	Stack pulumiapi.StackIdentifier
}

// plaintextInputSettings are the latest inputs of the resource, containing
// plaintext values wrapped in Secrets currentStateCipherSettings are the
// latest outputs/properties of the resource, containing ciphertext strings of
// secret values isInput is a flag that selects whether to generating an input
// PropertyMap that contains plaintext (true) or an output PropertyMap that
// contains ciphertext (false)
func (ds *PulumiServiceDeploymentSettingsInput) ToPropertyMap(
	plaintextInputSettings *pulumiapi.DeploymentSettings,
	currentStateCipherSettings *pulumiapi.DeploymentSettings,
	isInput bool,
) resource.PropertyMap {
	// Below flags are used throughout this method and direct the serialization
	// of twin value secrets Twin value secrets are values whose plaintext
	// cannot be retrieved from the API, thus forcing the development of this
	// fairly complex system When plaintextInputSettings is passed in, but
	// currentStateCipherSettings is not, that means the resource is being
	// created or updated
	createMode := plaintextInputSettings != nil && currentStateCipherSettings == nil
	// When both plaintextInputSettings and currentStateCipherSettings are
	// passed in, that means an existing resource is being refreshed, and it's
	// necessary to merge values In case we are merging, but some of the
	// properties don't previously exist, we will merge with empty value,
	// setting plaintext to be empty string
	mergeMode := plaintextInputSettings != nil && currentStateCipherSettings != nil
	// If neither one is passed in, we are importing an existing resource into
	// the state

	pm := resource.PropertyMap{}
	pm[gcOrganization] = resource.NewPropertyValue(ds.Stack.OrgName)
	pm[gcProject] = resource.NewPropertyValue(ds.Stack.ProjectName)
	pm[gcStack] = resource.NewPropertyValue(ds.Stack.StackName)

	if ds.AgentPoolID != nil {
		pm["agentPoolId"] = resource.NewPropertyValue(*ds.AgentPoolID)
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
							util.MergeSecretValue(
								sshAuthPropertyMap,
								"sshPrivateKey",
								ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey,
								plaintextValue,
								currentCipherValue,
								isInput,
							)
						} else if createMode {
							util.CreateSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey,
								plaintextInputSettings.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey, isInput)
						} else {
							util.ImportSecretValue(
								sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey, isInput,
							)
						}
					}
					if ds.SourceContext.Git.GitAuth.SSHAuth.Password != nil &&
						ds.SourceContext.Git.GitAuth.SSHAuth.Password.Value != "" {
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
							util.MergeSecretValue(
								sshAuthPropertyMap,
								"password",
								*ds.SourceContext.Git.GitAuth.SSHAuth.Password,
								plaintextValue,
								currentCipherValue,
								isInput,
							)
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
						if mergeMode {
							var plaintextValue *pulumiapi.SecretValue
							var currentCipherValue *pulumiapi.SecretValue
							if currentStateCipherSettings.SourceContext != nil &&
								currentStateCipherSettings.SourceContext.Git != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth != nil &&
								currentStateCipherSettings.SourceContext.Git.GitAuth.BasicAuth != nil {
								plaintextValue = &plaintextInputSettings.SourceContext.Git.GitAuth.BasicAuth.UserName
								currentCipherValue = &currentStateCipherSettings.SourceContext.Git.GitAuth.BasicAuth.UserName
							}
							util.MergeSecretValue(
								basicAuthPropertyMap,
								gcUsername,
								ds.SourceContext.Git.GitAuth.BasicAuth.UserName,
								plaintextValue,
								currentCipherValue,
								isInput,
							)
						} else if createMode {
							util.CreateSecretValue(basicAuthPropertyMap, gcUsername, ds.SourceContext.Git.GitAuth.BasicAuth.UserName,
								plaintextInputSettings.SourceContext.Git.GitAuth.BasicAuth.UserName, isInput)
						} else {
							util.ImportSecretValue(
								basicAuthPropertyMap, gcUsername, ds.SourceContext.Git.GitAuth.BasicAuth.UserName, isInput,
							)
						}
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
							util.MergeSecretValue(
								basicAuthPropertyMap,
								"password",
								ds.SourceContext.Git.GitAuth.BasicAuth.Password,
								plaintextValue,
								currentCipherValue,
								isInput,
							)
						} else if createMode {
							util.CreateSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password,
								plaintextInputSettings.SourceContext.Git.GitAuth.BasicAuth.Password, isInput)
						} else {
							util.ImportSecretValue(
								basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password, isInput,
							)
						}
					}
					gitAuthPropertyMap["basicAuth"] = resource.PropertyValue{V: basicAuthPropertyMap}
				}
				gitPropertyMap["gitAuth"] = resource.PropertyValue{V: gitAuthPropertyMap}
			}
			scMap["git"] = resource.PropertyValue{V: gitPropertyMap}
		}
		pm["sourceContext"] = resource.PropertyValue{V: scMap}
	}

	if ds.Operation != nil {
		ocMap := resource.PropertyMap{}
		if ds.Operation.PreRunCommands != nil {
			ocMap["preRunCommands"] = resource.NewPropertyValue(ds.Operation.PreRunCommands)
		}
		if ds.Operation.EnvironmentVariables != nil {
			evMap := resource.PropertyMap{}
			for k, v := range ds.Operation.EnvironmentVariables {
				if v.Secret {
					if mergeMode {
						var plaintextValue pulumiapi.SecretValue
						var currentCipherValue pulumiapi.SecretValue
						if currentStateCipherSettings.Operation != nil {
							plaintextValue = plaintextInputSettings.Operation.EnvironmentVariables[k]
							currentCipherValue = currentStateCipherSettings.Operation.EnvironmentVariables[k]
						}
						util.MergeSecretValue(evMap, k, v, &plaintextValue, &currentCipherValue, isInput)
					} else if createMode {
						util.CreateSecretValue(evMap, k, v,
							plaintextInputSettings.Operation.EnvironmentVariables[k], isInput)
					} else {
						util.ImportSecretValue(evMap, k, v, isInput)
					}
				} else {
					evMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.Value)
				}
			}
			ocMap["environmentVariables"] = resource.PropertyValue{V: evMap}
		}
		if ds.Operation.Options != nil {
			optionsMap := resource.PropertyMap{}
			if ds.Operation.Options.Shell != "" {
				optionsMap["shell"] = resource.NewPropertyValue(ds.Operation.Options.Shell)
			}
			if ds.Operation.Options.SkipInstallDependencies {
				optionsMap["skipInstallDependencies"] = resource.NewPropertyValue(true)
			}
			if ds.Operation.Options.SkipIntermediateDeployments {
				optionsMap["skipIntermediateDeployments"] = resource.NewPropertyValue(true)
			}
			if ds.Operation.Options.DeleteAfterDestroy {
				optionsMap["deleteAfterDestroy"] = resource.NewPropertyValue(true)
			}
			ocMap["options"] = resource.PropertyValue{V: optionsMap}
		}
		if ds.Operation.OIDC != nil {
			if ds.Operation.OIDC.AWS != nil || ds.Operation.OIDC.GCP != nil ||
				ds.Operation.OIDC.Azure != nil {
				oidcMap := resource.PropertyMap{}
				if ds.Operation.OIDC.AWS != nil {
					awsMap := resource.PropertyMap{}
					if ds.Operation.OIDC.AWS.RoleARN != "" {
						awsMap["roleARN"] = resource.NewPropertyValue(ds.Operation.OIDC.AWS.RoleARN)
					}
					if ds.Operation.OIDC.AWS.SessionName != "" {
						awsMap["sessionName"] = resource.NewPropertyValue(ds.Operation.OIDC.AWS.SessionName)
					}
					if ds.Operation.OIDC.AWS.PolicyARNs != nil {
						awsMap["policyARNs"] = resource.NewPropertyValue(ds.Operation.OIDC.AWS.PolicyARNs)
					}
					if ds.Operation.OIDC.AWS.Duration != "" {
						awsMap["duration"] = resource.NewPropertyValue(ds.Operation.OIDC.AWS.Duration)
					}
					oidcMap[gcAWS] = resource.PropertyValue{V: awsMap}
				}
				if ds.Operation.OIDC.GCP != nil {
					gcpMap := resource.PropertyMap{}
					if ds.Operation.OIDC.GCP.ProviderID != "" {
						gcpMap["providerId"] = resource.NewPropertyValue(ds.Operation.OIDC.GCP.ProviderID)
					}
					if ds.Operation.OIDC.GCP.ServiceAccount != "" {
						gcpMap["serviceAccount"] = resource.NewPropertyValue(
							ds.Operation.OIDC.GCP.ServiceAccount,
						)
					}
					if ds.Operation.OIDC.GCP.Region != "" {
						gcpMap["region"] = resource.NewPropertyValue(ds.Operation.OIDC.GCP.Region)
					}
					if ds.Operation.OIDC.GCP.WorkloadPoolID != "" {
						gcpMap["workloadPoolId"] = resource.NewPropertyValue(
							ds.Operation.OIDC.GCP.WorkloadPoolID,
						)
					}
					if ds.Operation.OIDC.GCP.ProjectID != "" {
						gcpMap["projectId"] = resource.NewPropertyValue(ds.Operation.OIDC.GCP.ProjectID)
					}
					if ds.Operation.OIDC.GCP.TokenLifetime != "" {
						gcpMap["tokenLifetime"] = resource.NewPropertyValue(ds.Operation.OIDC.GCP.TokenLifetime)
					}
					oidcMap["gcp"] = resource.PropertyValue{V: gcpMap}
				}
				if ds.Operation.OIDC.Azure != nil {
					azureMap := resource.PropertyMap{}
					if ds.Operation.OIDC.Azure.TenantID != "" {
						azureMap["tenantId"] = resource.NewPropertyValue(ds.Operation.OIDC.Azure.TenantID)
					}
					if ds.Operation.OIDC.Azure.ClientID != "" {
						azureMap["clientId"] = resource.NewPropertyValue(ds.Operation.OIDC.Azure.ClientID)
					}
					if ds.Operation.OIDC.Azure.SubscriptionID != "" {
						azureMap["subscriptionId"] = resource.NewPropertyValue(
							ds.Operation.OIDC.Azure.SubscriptionID,
						)
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
		githubMap["deployTags"] = resource.NewPropertyValue(ds.GitHub.DeployTags)
		if len(ds.GitHub.TagFilters) > 0 {
			githubMap["tagFilters"] = resource.NewPropertyValue(ds.GitHub.TagFilters)
		}
		pm["github"] = resource.PropertyValue{V: githubMap}
	}

	if ds.Vcs != nil {
		vcsMap := resource.PropertyMap{}
		provider, _ := ds.Vcs.GetDiscriminatorValue()
		vcsMap["provider"] = resource.NewPropertyValue(provider)
		vcsMap["deployCommits"] = resource.NewPropertyValue(ds.Vcs.DeployCommits())
		vcsMap["previewPullRequests"] = resource.NewPropertyValue(ds.Vcs.PreviewPullRequests())
		vcsMap["pullRequestTemplate"] = resource.NewPropertyValue(ds.Vcs.PullRequestTemplate())
		if ds.Vcs.Repository() != "" {
			vcsMap["repository"] = resource.NewPropertyValue(ds.Vcs.Repository())
		}
		if ds.Vcs.InstallationID() != "" {
			vcsMap["installationId"] = resource.NewPropertyValue(ds.Vcs.InstallationID())
		}
		if len(ds.Vcs.Paths()) > 0 {
			vcsMap["paths"] = resource.NewPropertyValue(ds.Vcs.Paths())
		}
		vcsMap["deployTags"] = resource.NewPropertyValue(ds.Vcs.DeployTags())
		if len(ds.Vcs.TagFilters()) > 0 {
			vcsMap["tagFilters"] = resource.NewPropertyValue(ds.Vcs.TagFilters())
		}
		if ds.Vcs.DeployPullRequest() != nil {
			vcsMap["deployPullRequest"] = resource.NewPropertyValue(float64(*ds.Vcs.DeployPullRequest()))
		}
		pm["vcs"] = resource.PropertyValue{V: vcsMap}
	}

	if ds.Executor != nil && ds.Executor.ExecutorImage != nil &&
		ds.Executor.ExecutorImage.Reference != "" {
		ecMap := resource.PropertyMap{}
		ecMap["executorImage"] = resource.NewPropertyValue(ds.Executor.ExecutorImage.Reference)
		if ds.Executor.ExecutorImage.Credentials != nil {
			credentialsPropertyMap := resource.PropertyMap{}
			if ds.Executor.ExecutorImage.Credentials.Username != "" {
				credentialsPropertyMap["username"] = resource.NewPropertyValue(
					ds.Executor.ExecutorImage.Credentials.Username,
				)
			}
			if ds.Executor.ExecutorImage.Credentials.Password.Value != "" {
				if mergeMode {
					var plaintextValue *pulumiapi.SecretValue
					var currentCipherValue *pulumiapi.SecretValue
					if executorImageCredentials(plaintextInputSettings) != nil &&
						executorImageCredentials(currentStateCipherSettings) != nil {
						plaintextValue = &executorImageCredentials(plaintextInputSettings).Password
						currentCipherValue = &executorImageCredentials(currentStateCipherSettings).Password
					}
					util.MergeSecretValue(
						credentialsPropertyMap,
						"password",
						ds.Executor.ExecutorImage.Credentials.Password,
						plaintextValue,
						currentCipherValue,
						isInput,
					)
				} else if createMode && executorImageCredentials(plaintextInputSettings) != nil {
					util.CreateSecretValue(credentialsPropertyMap, "password", ds.Executor.ExecutorImage.Credentials.Password,
						executorImageCredentials(plaintextInputSettings).Password, isInput)
				} else {
					util.ImportSecretValue(
						credentialsPropertyMap, "password", ds.Executor.ExecutorImage.Credentials.Password, isInput,
					)
				}
			}
			ecMap["credentials"] = resource.PropertyValue{V: credentialsPropertyMap}
		}
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

// executorImageCredentials walks the executor image credentials out of a settings
// struct, tolerating a nil at any level of the chain.
func executorImageCredentials(settings *pulumiapi.DeploymentSettings) *pulumiapi.DockerImageCredentials {
	if settings == nil || settings.Executor == nil || settings.Executor.ExecutorImage == nil {
		return nil
	}
	return settings.Executor.ExecutorImage.Credentials
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
		gcOrganization: true,
		gcProject:      true,
		gcStack:        true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec // safe conversion from plugin.DiffKind
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

func (ds *PulumiServiceDeploymentSettingsResource) Check(
	req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(req.GetNews(), util.KeepSecretsUnmarshal)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{gcOrganization, gcProject, gcStack} {
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
			if oidc[gcAWS].HasValue() {
				aws := util.GetSecretOrObjectValue(oidc[gcAWS])
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

	// Force the nested credentials to be secret. Their `"secret": true` in the
	// schema is codegen-only, and for a property nested inside a type only the
	// .NET SDK acts on it — so in every other language a program passing a
	// plaintext literal would have that literal written to state verbatim. See
	// util.MakeNestedSecret.
	for _, path := range [][]resource.PropertyKey{
		{gcSourceContext, gcGit, gcGitAuth, gcSSHAuth, gcSSHPrivateKey},
		{gcSourceContext, gcGit, gcGitAuth, gcSSHAuth, gcPassword},
		{gcSourceContext, gcGit, gcGitAuth, gcBasicAuth, gcUsername},
		{gcSourceContext, gcGit, gcGitAuth, gcBasicAuth, gcPassword},
		{gcExecutorContext, gcCredentials, gcPassword},
	} {
		util.MakeNestedSecret(news, path...)
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
	if propertyMap[gcStack].HasValue() {
		tempPlain, err := decodeDeploymentSettingsInput(inputMap)
		if err != nil {
			return nil, err
		}
		plaintextSettings = &tempPlain.DeploymentSettings
		tempCipher, err := decodeDeploymentSettingsInput(propertyMap)
		if err != nil {
			return nil, err
		}
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

func (ds *PulumiServiceDeploymentSettingsResource) Create(
	req *pulumirpc.CreateRequest,
) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(), util.KeepSecretsUnmarshal)
	if err != nil {
		return nil, err
	}

	input, err := decodeDeploymentSettingsInput(inputsMap)
	if err != nil {
		return nil, err
	}
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

func (ds *PulumiServiceDeploymentSettingsResource) Update(
	req *pulumirpc.UpdateRequest,
) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	inputsMap, err := plugin.UnmarshalProperties(req.GetNews(), util.KeepSecretsUnmarshal)
	if err != nil {
		return nil, err
	}

	input, err := decodeDeploymentSettingsInput(inputsMap)
	if err != nil {
		return nil, err
	}
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
