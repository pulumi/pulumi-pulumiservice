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

	if inputMap["executorContext"].HasValue() && inputMap["executorContext"].IsObject() {
		ecInput := inputMap["executorContext"].ObjectValue()
		var ec apitype.ExecutorContext

		if ecInput["executorImage"].HasValue() && ecInput["executorImage"].IsString() {
			ec.ExecutorImage = ecInput["executorImage"].StringValue()
		}

		input.ExecutorContext = &ec
	}

	if inputMap["github"].HasValue() && inputMap["github"].IsObject() {
		githubInput := inputMap["github"].ObjectValue()
		var github pulumiapi.GitHubConfiguration

		if githubInput["repository"].HasValue() && githubInput["repository"].IsString() {
			github.Repository = githubInput["repository"].StringValue()
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
				if v.IsString() {
					paths[i] = v.StringValue()
				}
			}

			github.Paths = paths
		}
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

			if gitInput["gitAuth"].HasValue() && gitInput["gitAuth"].IsObject() {
				authInput := gitInput["gitAuth"].ObjectValue()
				var a apitype.GitAuthConfig

				if authInput["sshAuth"].HasValue() && authInput["sshAuth"].IsObject() {
					sshInput := authInput["sshAuth"].ObjectValue()
					var s apitype.SSHAuth

					if sshInput["sshPrivateKey"].HasValue() && sshInput["sshPrivateKey"].IsString() {
						s.SSHPrivateKey = apitype.SecretValue{
							Value:  sshInput["sshPrivateKey"].StringValue(),
							Secret: true,
						}
					}
					if sshInput["password"].HasValue() && sshInput["password"].IsSecret() {
						s.Password = &apitype.SecretValue{
							Value:  sshInput["password"].StringValue(),
							Secret: true,
						}
					}

					a.SSHAuth = &s
				}

				if authInput["basicAuth"].HasValue() && authInput["basicAuth"].IsObject() {
					basicInput := authInput["basicAuth"].ObjectValue()
					var b apitype.BasicAuth

					if basicInput["username"].HasValue() && basicInput["username"].IsString() {
						b.UserName = apitype.SecretValue{
							Value:  basicInput["username"].StringValue(),
							Secret: false,
						}
					}
					if basicInput["password"].HasValue() && basicInput["password"].IsString() {
						b.Password = apitype.SecretValue{
							Value:  basicInput["password"].StringValue(),
							Secret: true,
						}
					}

					a.BasicAuth = &b
				}

				g.GitAuth = &a
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

		if ocInput["oidc"].HasValue() && ocInput["oidc"].IsObject() {
			oidcInput := ocInput["oidc"].ObjectValue()
			var oidc pulumiapi.OIDCConfiguration

			if oidcInput["aws"].HasValue() && oidcInput["aws"].IsObject() {
				awsInput := oidcInput["aws"].ObjectValue()
				var aws pulumiapi.AWSOIDCConfiguration

				if awsInput["roleARN"].HasValue() && awsInput["roleARN"].IsString() {
					aws.RoleARN = awsInput["roleARN"].StringValue()
				}
				if awsInput["duration"].HasValue() && awsInput["duration"].IsString() {
					aws.Duration = awsInput["duration"].StringValue()
				}
				if awsInput["sessionName"].HasValue() && awsInput["sessionName"].IsString() {
					aws.SessionName = awsInput["sessionName"].StringValue()
				}
				if awsInput["policyARNs"].HasValue() && awsInput["policyARNs"].IsArray() {
					policyARNsInput := awsInput["policyARNs"].ArrayValue()
					policyARNs := make([]string, len(policyARNsInput))

					for i, v := range policyARNsInput {
						if v.IsString() {
							policyARNs[i] = v.StringValue()
						}
					}

					aws.PolicyARNs = policyARNs
				}

				oidc.AWS = &aws
			}

			if oidcInput["gcp"].HasValue() && oidcInput["gcp"].IsObject() {
				gcpInput := oidcInput["gcp"].ObjectValue()
				var gcp pulumiapi.GCPOIDCConfiguration

				if gcpInput["projectId"].HasValue() && gcpInput["projectId"].IsString() {
					gcp.ProjectID = gcpInput["projectId"].StringValue()
				}
				if gcpInput["region"].HasValue() && gcpInput["region"].IsString() {
					gcp.Region = gcpInput["region"].StringValue()
				}
				if gcpInput["workloadPoolId"].HasValue() && gcpInput["workloadPoolId"].IsString() {
					gcp.WorkloadPoolID = gcpInput["workloadPoolId"].StringValue()
				}
				if gcpInput["providerId"].HasValue() && gcpInput["providerId"].IsString() {
					gcp.ProviderID = gcpInput["providerId"].StringValue()
				}
				if gcpInput["serviceAccount"].HasValue() && gcpInput["serviceAccount"].IsString() {
					gcp.ServiceAccount = gcpInput["serviceAccount"].StringValue()
				}
				if gcpInput["tokenLifetime"].HasValue() && gcpInput["tokenLifetime"].IsString() {
					gcp.TokenLifetime = gcpInput["tokenLifetime"].StringValue()
				}

				oidc.GCP = &gcp
			}

			if oidcInput["azure"].HasValue() && oidcInput["azure"].IsObject() {
				azureInput := oidcInput["azure"].ObjectValue()
				var azure pulumiapi.AzureOIDCConfiguration

				if azureInput["tenantId"].HasValue() && azureInput["tenantId"].IsString() {
					azure.TenantID = azureInput["tenantId"].StringValue()
				}
				if azureInput["clientId"].HasValue() && azureInput["clientId"].IsString() {
					azure.ClientID = azureInput["clientId"].StringValue()
				}
				if azureInput["subscriptionId"].HasValue() && azureInput["subscriptionId"].IsString() {
					azure.SubscriptionID = azureInput["subscriptionId"].StringValue()
				}

				oidc.Azure = &azure
			}

			oc.OIDC = &oidc
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
	// For simplicity, all updates are destructive, so we just call Create.
	return &pulumirpc.UpdateResponse{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Name() string {
	return "pulumiservice:index:DeploymentSettings"
}
