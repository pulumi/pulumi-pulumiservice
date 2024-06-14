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

// This is a value for imported secrets, to hint that value needs to be replaced
// in generated code
const replaceMe = "<REPLACE WITH ACTUAL SECRET VALUE>"

type PulumiServiceDeploymentSettingsInput struct {
	pulumiapi.DeploymentSettings
	Stack pulumiapi.StackName
}

func (ds *PulumiServiceDeploymentSettingsInput) ToPropertyMap(plaintextSettings *pulumiapi.DeploymentSettings, oldCipherSettings *pulumiapi.DeploymentSettings, isInput bool) resource.PropertyMap {
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
						if plaintextSettings == nil {
							importSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey, isInput)
						} else {
							if oldCipherSettings == nil {
								createSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey,
									plaintextSettings.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey)
							} else {
								if oldCipherSettings.SourceContext != nil &&
									oldCipherSettings.SourceContext.Git != nil &&
									oldCipherSettings.SourceContext.Git.GitAuth != nil &&
									oldCipherSettings.SourceContext.Git.GitAuth.SSHAuth != nil {
									mergeSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey,
										plaintextSettings.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey,
										oldCipherSettings.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey)
								} else {
									importSecretValue(sshAuthPropertyMap, "sshPrivateKey", ds.SourceContext.Git.GitAuth.SSHAuth.SSHPrivateKey, isInput)
								}
							}
						}
					}
					if ds.SourceContext.Git.GitAuth.SSHAuth.Password.Value != "" {
						if plaintextSettings == nil {
							importSecretValue(sshAuthPropertyMap, "password", *ds.SourceContext.Git.GitAuth.SSHAuth.Password, isInput)
						} else {
							if oldCipherSettings == nil {
								createSecretValue(sshAuthPropertyMap, "password", *ds.SourceContext.Git.GitAuth.SSHAuth.Password,
									*plaintextSettings.SourceContext.Git.GitAuth.SSHAuth.Password)
							} else {
								if oldCipherSettings.SourceContext != nil &&
									oldCipherSettings.SourceContext.Git != nil &&
									oldCipherSettings.SourceContext.Git.GitAuth != nil &&
									oldCipherSettings.SourceContext.Git.GitAuth.SSHAuth != nil &&
									oldCipherSettings.SourceContext.Git.GitAuth.SSHAuth.Password != nil {
									mergeSecretValue(sshAuthPropertyMap, "password", *ds.SourceContext.Git.GitAuth.SSHAuth.Password,
										*plaintextSettings.SourceContext.Git.GitAuth.SSHAuth.Password,
										*oldCipherSettings.SourceContext.Git.GitAuth.SSHAuth.Password)
								} else {
									importSecretValue(sshAuthPropertyMap, "password", *ds.SourceContext.Git.GitAuth.SSHAuth.Password, isInput)
								}
							}
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
						if plaintextSettings == nil {
							importSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password, isInput)
						} else {
							if oldCipherSettings == nil {
								createSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password,
									plaintextSettings.SourceContext.Git.GitAuth.BasicAuth.Password)
							} else {
								if oldCipherSettings.SourceContext != nil &&
									oldCipherSettings.SourceContext.Git != nil &&
									oldCipherSettings.SourceContext.Git.GitAuth != nil &&
									oldCipherSettings.SourceContext.Git.GitAuth.BasicAuth != nil {
									mergeSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password,
										plaintextSettings.SourceContext.Git.GitAuth.BasicAuth.Password,
										oldCipherSettings.SourceContext.Git.GitAuth.BasicAuth.Password)
								} else {
									importSecretValue(basicAuthPropertyMap, "password", ds.SourceContext.Git.GitAuth.BasicAuth.Password, isInput)
								}
							}
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

	if ds.OperationContext != nil {
		ocMap := resource.PropertyMap{}
		if ds.OperationContext.PreRunCommands != nil {
			ocMap["preRunCommands"] = resource.NewPropertyValue(ds.OperationContext.PreRunCommands)
		}
		if ds.OperationContext.EnvironmentVariables != nil {
			evMap := resource.PropertyMap{}
			evMapCipher := resource.PropertyMap{}
			for k, v := range ds.OperationContext.EnvironmentVariables {
				if v.Secret {
					if plaintextSettings == nil {
						// Import
						if isInput {
							evMap[resource.PropertyKey(k)] = resource.MakeSecret(resource.NewPropertyValue(replaceMe))
						} else {
							evMap[resource.PropertyKey(k)] = resource.MakeSecret(resource.NewPropertyValue(""))
						}
						evMapCipher[resource.PropertyKey(k)] = resource.NewPropertyValue(v.Value)
					} else {
						if oldCipherSettings == nil {
							// Create
							evMap[resource.PropertyKey(k)] = resource.MakeSecret(resource.NewPropertyValue(plaintextSettings.OperationContext.EnvironmentVariables[k].Value))
							evMapCipher[resource.PropertyKey(k)] = resource.NewPropertyValue(v.Value)
						} else {
							if oldCipherSettings.OperationContext != nil {
								// Merge
								if v.Value == oldCipherSettings.OperationContext.EnvironmentVariables[k].Value {
									evMap[resource.PropertyKey(k)] = resource.MakeSecret(resource.NewPropertyValue(plaintextSettings.OperationContext.EnvironmentVariables[k].Value))
								} else {
									evMap[resource.PropertyKey(k)] = resource.MakeSecret(resource.NewPropertyValue(""))
								}
								evMapCipher[resource.PropertyKey(k)] = resource.NewPropertyValue(v.Value)
							} else {
								// Import
								if isInput {
									evMap[resource.PropertyKey(k)] = resource.MakeSecret(resource.NewPropertyValue(replaceMe))
								} else {
									evMap[resource.PropertyKey(k)] = resource.MakeSecret(resource.NewPropertyValue(""))
								}
								evMapCipher[resource.PropertyKey(k)] = resource.NewPropertyValue(v.Value)
							}
						}
					}
				} else {
					evMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v.Value)
					evMapCipher[resource.PropertyKey(k)] = resource.NewPropertyValue("")
				}
			}
			ocMap["environmentVariables"] = resource.PropertyValue{V: evMap}
			ocMap["environmentVariablesCipher"] = resource.PropertyValue{V: evMapCipher}
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
		ecMap["executorImage"] = resource.NewPropertyValue(ds.ExecutorContext.ExecutorImage.Reference)
		pm["executorContext"] = resource.PropertyValue{V: ecMap}
	}
	return pm
}

func importSecretValue(propertyMap resource.PropertyMap, propertyName string, cipherValue pulumiapi.SecretValue, isInput bool) {
	propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(""))
	propertyMap[resource.PropertyKey(propertyName+"Cipher")] = resource.NewPropertyValue(cipherValue.Value)

	// Adding this for code generation
	if isInput {
		propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(replaceMe))
	}
}

func createSecretValue(propertyMap resource.PropertyMap, propertyName string, cipherValue pulumiapi.SecretValue, plaintextValue pulumiapi.SecretValue) {
	propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(plaintextValue.Value))
	propertyMap[resource.PropertyKey(propertyName+"Cipher")] = resource.NewPropertyValue(cipherValue.Value)
}

func mergeSecretValue(propertyMap resource.PropertyMap, propertyName string, cipherValue pulumiapi.SecretValue, plaintextValue pulumiapi.SecretValue, oldCipherValue pulumiapi.SecretValue) {
	propertyMap[resource.PropertyKey(propertyName+"Cipher")] = resource.NewPropertyValue(cipherValue.Value)

	if cipherValue.Value == oldCipherValue.Value {
		propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(plaintextValue.Value))
	} else {
		propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(""))
	}
}

type PulumiServiceDeploymentSettingsResource struct {
	client pulumiapi.DeploymentSettingsClient
}

func (ds *PulumiServiceDeploymentSettingsResource) ToPulumiServiceDeploymentSettingsInput(inputMap resource.PropertyMap, gettingPlaintext bool) PulumiServiceDeploymentSettingsInput {
	input := PulumiServiceDeploymentSettingsInput{}

	input.Stack.OrgName = getSecretOrStringValue(inputMap["organization"])
	input.Stack.ProjectName = getSecretOrStringValue(inputMap["project"])
	input.Stack.StackName = getSecretOrStringValue(inputMap["stack"])

	if inputMap["agentPoolId"].HasValue() && inputMap["agentPoolId"].IsString() {
		input.AgentPoolId = inputMap["agentPoolId"].StringValue()
	}

	input.ExecutorContext = toExecutorContext(inputMap)
	input.GitHub = toGitHubConfig(inputMap)
	input.SourceContext = toSourceContext(inputMap, gettingPlaintext)
	input.OperationContext = toOperationContext(inputMap, gettingPlaintext)

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

func toSourceContext(inputMap resource.PropertyMap, gettingPlaintext bool) *pulumiapi.SourceContext {
	if !inputMap["sourceContext"].HasValue() || !inputMap["sourceContext"].IsObject() {
		return nil
	}

	scInput := inputMap["sourceContext"].ObjectValue()
	var sc pulumiapi.SourceContext

	if scInput["git"].HasValue() && scInput["git"].IsObject() {
		gitInput := scInput["git"].ObjectValue()
		var g pulumiapi.SourceContextGit

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
			var a pulumiapi.GitAuthConfig

			if authInput["sshAuth"].HasValue() && authInput["sshAuth"].IsObject() {
				sshInput := authInput["sshAuth"].ObjectValue()
				var s pulumiapi.SSHAuth

				if sshInput["sshPrivateKey"].HasValue() || sshInput["sshPrivateKeyCipher"].HasValue() {
					s.SSHPrivateKey = pulumiapi.SecretValue{
						Secret: true,
						Value:  getTwinSecretValue(sshInput, "sshPrivateKey", gettingPlaintext),
					}
				}
				if sshInput["password"].HasValue() || sshInput["passwordCipher"].HasValue() {
					s.Password = &pulumiapi.SecretValue{
						Secret: true,
						Value:  getTwinSecretValue(sshInput, "password", gettingPlaintext),
					}
				}

				a.SSHAuth = &s
			}

			if authInput["basicAuth"].HasValue() && authInput["basicAuth"].IsObject() {
				basicInput := authInput["basicAuth"].ObjectValue()
				var b pulumiapi.BasicAuth

				if basicInput["username"].HasValue() {
					b.UserName = pulumiapi.SecretValue{
						Value:  getSecretOrStringValue(basicInput["username"]),
						Secret: false,
					}
				}
				if basicInput["password"].HasValue() || basicInput["passwordCipher"].HasValue() {
					b.Password = pulumiapi.SecretValue{
						Value:  getTwinSecretValue(basicInput, "password", gettingPlaintext),
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

func toOperationContext(inputMap resource.PropertyMap, gettingPlaintext bool) *pulumiapi.OperationContext {
	if !inputMap["operationContext"].HasValue() || !inputMap["operationContext"].IsObject() {
		return nil
	}

	ocInput := inputMap["operationContext"].ObjectValue()
	var oc pulumiapi.OperationContext

	if gettingPlaintext {
		if ocInput["environmentVariables"].HasValue() && ocInput["environmentVariables"].IsObject() {
			ev := map[string]pulumiapi.SecretValue{}
			evInput := ocInput["environmentVariables"].ObjectValue()

			for k, v := range evInput {
				value := getSecretOrStringValue(v)
				ev[string(k)] = pulumiapi.SecretValue{Secret: v.IsSecret(), Value: value}
			}

			oc.EnvironmentVariables = ev
		}
	} else {
		if ocInput["environmentVariablesCipher"].HasValue() && ocInput["environmentVariablesCipher"].IsObject() {
			ev := map[string]pulumiapi.SecretValue{}
			evInput := ocInput["environmentVariablesCipher"].ObjectValue()

			for k, v := range evInput {
				value := getSecretOrStringValue(v)
				ev[string(k)] = pulumiapi.SecretValue{Secret: v.IsSecret(), Value: value}
			}

			oc.EnvironmentVariables = ev
		}
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
	case nil:
		return ""
	default:
		return prop.StringValue()
	}
}

func getTwinSecretValue(propertyMap resource.PropertyMap, key string, gettingPlaintext bool) string {
	if gettingPlaintext {
		if propertyMap[resource.PropertyKey(key)].HasValue() {
			return getSecretOrStringValue(propertyMap[resource.PropertyKey(key)])
		} else {
			return getSecretOrStringValue(propertyMap[resource.PropertyKey(key)])
		}
	} else {
		return getSecretOrStringValue(propertyMap[resource.PropertyKey(key+"Cipher")])
	}
}

func (ds *PulumiServiceDeploymentSettingsResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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

func (ds *PulumiServiceDeploymentSettingsResource) Configure(_ PulumiServiceConfig) {}

func (ds *PulumiServiceDeploymentSettingsResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	var stack pulumiapi.StackName
	if err := stack.FromID(req.Id); err != nil {
		return nil, err
	}
	settings, err := ds.client.GetDeploymentSettings(ctx, stack)
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

	var plaintextSettings *pulumiapi.DeploymentSettings = nil
	var ciphertextSettings *pulumiapi.DeploymentSettings = nil
	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true})
	if err != nil {
		return nil, err
	}
	if inputMap["stack"].HasValue() {
		tempPlain := ds.ToPulumiServiceDeploymentSettingsInput(inputMap, true)
		plaintextSettings = &tempPlain.DeploymentSettings
		tempCipher := ds.ToPulumiServiceDeploymentSettingsInput(inputMap, false)
		ciphertextSettings = &tempCipher.DeploymentSettings
	}

	properties, err := plugin.MarshalProperties(
		dsInput.ToPropertyMap(plaintextSettings, ciphertextSettings, false),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
		},
	)
	if err != nil {
		return nil, err
	}

	inputs, err := plugin.MarshalProperties(
		dsInput.ToPropertyMap(plaintextSettings, ciphertextSettings, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
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

func (ds *PulumiServiceDeploymentSettingsResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var stack pulumiapi.StackName
	if err := stack.FromID(req.Id); err != nil {
		return nil, err
	}

	err := ds.client.DeleteDeploymentSettings(ctx, stack)
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

	input := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap, true)
	settings := input.DeploymentSettings
	response, err := ds.client.CreateDeploymentSettings(ctx, input.Stack, settings)
	if err != nil {
		return nil, err
	}

	responseInput := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: *response,
		Stack:              input.Stack,
	}

	outputProperties, err := plugin.MarshalProperties(
		responseInput.ToPropertyMap(&input.DeploymentSettings, nil, false),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
		},
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
	inputsMap, err := plugin.UnmarshalProperties(req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true})
	if err != nil {
		return nil, err
	}

	input := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap, true)
	settings := input.DeploymentSettings
	response, err := ds.client.UpdateDeploymentSettings(ctx, input.Stack, settings)
	if err != nil {
		return nil, err
	}

	responseInput := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: *response,
		Stack:              input.Stack,
	}

	outputProperties, err := plugin.MarshalProperties(
		responseInput.ToPropertyMap(&input.DeploymentSettings, nil, false),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
		},
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
