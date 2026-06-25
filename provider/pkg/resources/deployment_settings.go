// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"fmt"
	"path"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type DeploymentSettings struct{}

var (
	_ infer.CustomCheck[DeploymentSettingsInput]                           = &DeploymentSettings{}
	_ infer.CustomCreate[DeploymentSettingsInput, DeploymentSettingsState] = &DeploymentSettings{}
	_ infer.CustomUpdate[DeploymentSettingsInput, DeploymentSettingsState] = &DeploymentSettings{}
	_ infer.CustomDelete[DeploymentSettingsState]                          = &DeploymentSettings{}
	_ infer.CustomRead[DeploymentSettingsInput, DeploymentSettingsState]   = &DeploymentSettings{}
)

func (*DeploymentSettings) Annotate(a infer.Annotator) {
	a.Describe(&DeploymentSettings{}, "Deployment settings configure Pulumi Deployments for a stack.\n\n"+
		"### Import\n\n"+
		"Deployment settings can be imported using the `id`, which for deployment settings is "+
		"`{org}/{project}/{stack}` e.g.,\n\n"+
		"```sh\n $ pulumi import pulumiservice:index:DeploymentSettings my_settings my-org/my-project/my-stack\n```\n\n")
	a.SetToken("index", "DeploymentSettings")
}

// DeploymentSettingsInput models the user-facing inputs for a DeploymentSettings resource.
type DeploymentSettingsInput struct {
	Organization     string                              `pulumi:"organization" provider:"replaceOnChanges"`
	Project          string                              `pulumi:"project"      provider:"replaceOnChanges"`
	Stack            string                              `pulumi:"stack"        provider:"replaceOnChanges"`
	AgentPoolID      *string                             `pulumi:"agentPoolId,optional"`
	ExecutorContext  *DeploymentSettingsExecutorContext  `pulumi:"executorContext,optional"`
	SourceContext    *DeploymentSettingsSourceContext    `pulumi:"sourceContext,optional"`
	GitHub           *DeploymentSettingsGitHubBlock      `pulumi:"github,optional"`
	Vcs              *DeploymentSettingsVcsBlock         `pulumi:"vcs,optional"`
	OperationContext *DeploymentSettingsOperationContext `pulumi:"operationContext,optional"`
	CacheOptions     *DeploymentSettingsCacheOptions     `pulumi:"cacheOptions,optional"`
}

func (i *DeploymentSettingsInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.Describe(&i.Stack, "Stack name.")
	a.Describe(&i.AgentPoolID, "The agent pool identifier to use for the deployment.")
	a.Describe(&i.ExecutorContext, "Settings related to the deployment executor.")
	a.Describe(&i.SourceContext, "Settings related to the source of the deployment.")
	a.Describe(&i.GitHub, "GitHub settings for the deployment.")
	a.Deprecate(&i.GitHub, "Use the 'vcs' property instead, which supports both GitHub and Azure DevOps.")
	a.Describe(&i.Vcs, "VCS settings for the deployment. Supports Azure DevOps and GitHub via "+
		"the 'provider' discriminator field.")
	a.Describe(&i.OperationContext, "Settings related to the Pulumi operation environment during the deployment.")
	a.Describe(&i.CacheOptions, "Dependency cache settings for the deployment")
}

// DeploymentSettingsState is the persisted state of a DeploymentSettings resource. It mirrors the input
// fields exactly so that refresh and update operate symmetrically.
type DeploymentSettingsState struct {
	DeploymentSettingsInput
}

// DeploymentSettingsExecutorContext defines the executor environment for the deployment.
type DeploymentSettingsExecutorContext struct {
	ExecutorImage string `pulumi:"executorImage"`
}

func (c *DeploymentSettingsExecutorContext) Annotate(a infer.Annotator) {
	a.Describe(c, "The executor context defines information about the executor where the deployment is executed. "+
		"If unspecified, the default 'pulumi/pulumi' image is used.")
	a.Describe(&c.ExecutorImage, "Allows overriding the default executor image with a custom image. "+
		"E.g. 'pulumi/pulumi-nodejs:latest'")
}

// DeploymentSettingsSourceContext is the source-of-truth for the deployment's source code.
type DeploymentSettingsSourceContext struct {
	Git *DeploymentSettingsGitSource `pulumi:"git,optional"`
}

func (s *DeploymentSettingsSourceContext) Annotate(a infer.Annotator) {
	a.Describe(s, "Settings related to the source of the deployment.")
	a.Describe(&s.Git, "Git source settings for a deployment.")
}

// DeploymentSettingsGitSource describes git settings for the deployment.
type DeploymentSettingsGitSource struct {
	RepoURL string                              `pulumi:"repoUrl,optional"`
	Branch  string                              `pulumi:"branch,optional"`
	Commit  string                              `pulumi:"commit,optional"`
	RepoDir string                              `pulumi:"repoDir,optional"`
	GitAuth *DeploymentSettingsGitSourceGitAuth `pulumi:"gitAuth,optional"`
}

func (g *DeploymentSettingsGitSource) Annotate(a infer.Annotator) {
	a.Describe(g, "Git source settings for a deployment.")
	a.Describe(&g.RepoURL, "The repository URL to use for git settings. "+
		"Should not be specified if there are `gitHub` settings for this deployment.")
	a.Describe(&g.Branch, "The branch to deploy. One of either `branch` or `commit` must be specified.")
	a.Describe(&g.Commit, "The commit to deploy. One of either `branch` or `commit` must be specified.")
	a.Describe(&g.RepoDir, "The directory within the repository where the Pulumi.yaml is located.")
	a.Describe(&g.GitAuth, "Git authentication configuration for this deployment. "+
		"Should not be specified if there are `gitHub` settings for this deployment.")
}

// DeploymentSettingsGitSourceGitAuth holds either SSH or basic-auth credentials for git access.
type DeploymentSettingsGitSourceGitAuth struct {
	SSHAuth   *DeploymentSettingsGitAuthSSHAuth   `pulumi:"sshAuth,optional"`
	BasicAuth *DeploymentSettingsGitAuthBasicAuth `pulumi:"basicAuth,optional"`
}

func (g *DeploymentSettingsGitSourceGitAuth) Annotate(a infer.Annotator) {
	a.Describe(g, "Git source settings for a deployment.")
	a.Describe(&g.SSHAuth, "SSH auth for git authentication. "+
		"Only one of `personalAccessToken`, `sshAuth`, or `basicAuth` must be defined.")
	a.Describe(&g.BasicAuth, "Basic auth for git authentication. "+
		"Only one of `personalAccessToken`, `sshAuth`, or `basicAuth` must be defined.")
}

// DeploymentSettingsGitAuthSSHAuth holds SSH authentication credentials.
type DeploymentSettingsGitAuthSSHAuth struct {
	SSHPrivateKey string  `pulumi:"sshPrivateKey" provider:"secret"`
	Password      *string `pulumi:"password,optional" provider:"secret"`
}

func (s *DeploymentSettingsGitAuthSSHAuth) Annotate(a infer.Annotator) {
	a.Describe(s, "Git source settings for a deployment.")
	a.Describe(&s.SSHPrivateKey, "SSH private key.")
	a.Describe(&s.Password, "Optional password for SSH authentication.")
}

// DeploymentSettingsGitAuthBasicAuth holds basic-auth credentials.
type DeploymentSettingsGitAuthBasicAuth struct {
	UserName string `pulumi:"username" provider:"secret"`
	Password string `pulumi:"password" provider:"secret"`
}

func (b *DeploymentSettingsGitAuthBasicAuth) Annotate(a infer.Annotator) {
	a.Describe(b, "Git source settings for a deployment.")
	a.Describe(&b.UserName, "User name for git basic authentication.")
	a.Describe(&b.Password, "Password for git basic authentication.")
}

// DeploymentSettingsGitHubBlock is the deprecated GitHub-specific deployment settings.
type DeploymentSettingsGitHubBlock struct {
	Repository          string   `pulumi:"repository,optional"`
	DeployCommits       bool     `pulumi:"deployCommits,optional"`
	PreviewPullRequests bool     `pulumi:"previewPullRequests,optional"`
	PullRequestTemplate bool     `pulumi:"pullRequestTemplate,optional"`
	Paths               []string `pulumi:"paths,optional"`
}

func (g *DeploymentSettingsGitHubBlock) Annotate(a infer.Annotator) {
	a.Describe(g, "GitHub settings for the deployment.")
	a.Describe(&g.Repository, "The GitHub repository in the format org/repo.")
	a.Describe(&g.DeployCommits, "Trigger a deployment running `pulumi up` on commit.")
	a.Describe(&g.PreviewPullRequests, "Trigger a deployment running `pulumi preview` when a PR is opened.")
	a.Describe(&g.PullRequestTemplate, "Use this stack as a template for pull request review stacks.")
	a.Describe(&g.Paths, "The paths within the repo that deployments should be filtered to.")
	a.SetDefault(&g.DeployCommits, true)
	a.SetDefault(&g.PreviewPullRequests, true)
	a.SetDefault(&g.PullRequestTemplate, false)
	a.SetToken("index", "DeploymentSettingsGithub")
}

// DeploymentSettingsVcsBlock is the VCS-provider-agnostic settings, discriminated by `Provider`.
type DeploymentSettingsVcsBlock struct {
	Provider            string   `pulumi:"provider"`
	Repository          string   `pulumi:"repository,optional"`
	InstallationID      string   `pulumi:"installationId,optional"`
	DeployCommits       bool     `pulumi:"deployCommits,optional"`
	PreviewPullRequests bool     `pulumi:"previewPullRequests,optional"`
	PullRequestTemplate bool     `pulumi:"pullRequestTemplate,optional"`
	Paths               []string `pulumi:"paths,optional"`
	DeployPullRequest   *int     `pulumi:"deployPullRequest,optional"`
}

func (v *DeploymentSettingsVcsBlock) Annotate(a infer.Annotator) {
	a.Describe(v, "VCS settings for the deployment, supporting multiple VCS providers.")
	a.Describe(&v.Provider, "The VCS provider type.")
	a.Describe(&v.Repository, "The repository identifier (e.g., 'ProjectName/RepoName' for Azure DevOps, "+
		"'org/repo' for GitHub).")
	a.Describe(&v.InstallationID, "The VCS integration installation ID. Use to disambiguate when an "+
		"organization has multiple integrations of the same provider type (e.g., two GitHub Apps). "+
		"If omitted, the API resolves the integration automatically from `provider` and `repository`.")
	a.Describe(&v.DeployCommits, "Trigger a deployment running `pulumi up` on commit.")
	a.Describe(&v.PreviewPullRequests, "Trigger a deployment running `pulumi preview` when a PR is opened.")
	a.Describe(&v.PullRequestTemplate, "Use this stack as a template for pull request review stacks.")
	a.Describe(&v.Paths, "The paths within the repo that deployments should be filtered to.")
	a.Describe(&v.DeployPullRequest, "Deploy a specific pull request number.")
	a.SetDefault(&v.DeployCommits, true)
	a.SetDefault(&v.PreviewPullRequests, true)
	a.SetDefault(&v.PullRequestTemplate, false)
	a.SetToken("index", "DeploymentSettingsVcs")
}

// DeploymentSettingsOperationContext describes the operation environment for the deployment.
type DeploymentSettingsOperationContext struct {
	PreRunCommands       []string                    `pulumi:"preRunCommands,optional"`
	EnvironmentVariables map[string]string           `pulumi:"environmentVariables,optional"`
	Options              *DeploymentOperationOptions `pulumi:"options,optional"`
	Oidc                 *DeploymentOperationOIDC    `pulumi:"oidc,optional"`
}

func (o *DeploymentSettingsOperationContext) Annotate(a infer.Annotator) {
	a.Describe(o, "Settings related to the Pulumi operation environment during the deployment.")
	a.Describe(&o.PreRunCommands, "Shell commands to run before the Pulumi operation executes.")
	a.Describe(&o.EnvironmentVariables, "Environment variables to set for the deployment.")
	a.Describe(&o.Options, "Options to override default behavior during the deployment.")
	a.Describe(&o.Oidc, "OIDC configuration to use during the deployment.")
}

// DeploymentOperationOptions is the schema's OperationContextOptions type.
type DeploymentOperationOptions struct {
	SkipInstallDependencies     bool   `pulumi:"skipInstallDependencies,optional"`
	SkipIntermediateDeployments bool   `pulumi:"skipIntermediateDeployments,optional"`
	Shell                       string `pulumi:"shell,optional"`
	DeleteAfterDestroy          bool   `pulumi:"deleteAfterDestroy,optional"`
}

func (o *DeploymentOperationOptions) Annotate(a infer.Annotator) {
	a.Describe(&o.SkipInstallDependencies, "Skip the default dependency installation step - use this to "+
		"customize the dependency installation (e.g. if using yarn or poetry)")
	a.Describe(&o.SkipIntermediateDeployments, "Skip intermediate deployments (Consolidate multiple deployments "+
		"of the same type into one deployment)")
	a.Describe(&o.Shell, "The shell to use to run commands during the deployment. Defaults to 'bash'.")
	a.Describe(&o.DeleteAfterDestroy, "Whether the stack should be deleted after it is destroyed.")
	a.SetToken("index", "OperationContextOptions")
}

// DeploymentOperationOIDC is the schema's OperationContextOIDC type.
type DeploymentOperationOIDC struct {
	AWS   *DeploymentAWSOIDCConfiguration   `pulumi:"aws,optional"`
	GCP   *DeploymentGCPOIDCConfiguration   `pulumi:"gcp,optional"`
	Azure *DeploymentAzureOIDCConfiguration `pulumi:"azure,optional"`
}

func (o *DeploymentOperationOIDC) Annotate(a infer.Annotator) {
	a.Describe(&o.AWS, "AWS-specific OIDC configuration.")
	a.Describe(&o.GCP, "GCP-specific OIDC configuration.")
	a.Describe(&o.Azure, "Azure-specific OIDC configuration.")
	a.SetToken("index", "OperationContextOIDC")
}

// DeploymentAWSOIDCConfiguration is the schema's AWSOIDCConfiguration type.
type DeploymentAWSOIDCConfiguration struct {
	Duration    string   `pulumi:"duration,optional"`
	PolicyARNs  []string `pulumi:"policyARNs,optional"`
	RoleARN     string   `pulumi:"roleARN"`
	SessionName string   `pulumi:"sessionName"`
}

func (o *DeploymentAWSOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(&o.Duration, "Duration of the assume-role session in “XhYmZs” format")
	a.Describe(&o.PolicyARNs, "Optional set of IAM policy ARNs that further restrict the assume-role session")
	a.Describe(&o.RoleARN, "The ARN of the role to assume using the OIDC token.")
	a.Describe(&o.SessionName, "The name of the assume-role session.")
	a.SetToken("index", "AWSOIDCConfiguration")
}

// DeploymentGCPOIDCConfiguration is the schema's GCPOIDCConfiguration type.
type DeploymentGCPOIDCConfiguration struct {
	ProjectID      string `pulumi:"projectId"`
	Region         string `pulumi:"region,optional"`
	WorkloadPoolID string `pulumi:"workloadPoolId"`
	ProviderID     string `pulumi:"providerId"`
	ServiceAccount string `pulumi:"serviceAccount"`
	TokenLifetime  string `pulumi:"tokenLifetime,optional"`
}

func (o *DeploymentGCPOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(&o.ProjectID, "The numerical ID of the GCP project.")
	a.Describe(&o.Region, "The region of the GCP project.")
	a.Describe(&o.WorkloadPoolID, "The ID of the workload pool to use.")
	a.Describe(&o.ProviderID, "The ID of the identity provider associated with the workload pool.")
	a.Describe(&o.ServiceAccount, "The email address of the service account to use.")
	a.Describe(&o.TokenLifetime, "The lifetime of the temporary credentials in “XhYmZs” format.")
	a.SetToken("index", "GCPOIDCConfiguration")
}

// DeploymentAzureOIDCConfiguration is the schema's AzureOIDCConfiguration type.
type DeploymentAzureOIDCConfiguration struct {
	ClientID       string `pulumi:"clientId"`
	TenantID       string `pulumi:"tenantId"`
	SubscriptionID string `pulumi:"subscriptionId"`
}

func (o *DeploymentAzureOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(&o.ClientID, "The client ID of the federated workload identity.")
	a.Describe(&o.TenantID, "The tenant ID of the federated workload identity.")
	a.Describe(&o.SubscriptionID, "The subscription ID of the federated workload identity.")
	a.SetToken("index", "AzureOIDCConfiguration")
}

// DeploymentSettingsCacheOptions is the schema's DeploymentSettingsCacheOptions type.
type DeploymentSettingsCacheOptions struct {
	Enable bool `pulumi:"enable,optional"`
}

func (c *DeploymentSettingsCacheOptions) Annotate(a infer.Annotator) {
	a.Describe(c, "Dependency cache settings for the deployment")
	a.Describe(&c.Enable, "Enable dependency caching")
	a.SetDefault(&c.Enable, false)
}

// --- CRUD implementation ---

// Check normalizes user input. In particular, the AWS OIDC `duration` field accepts shorthand
// like "1h" but the Pulumi Cloud API requires the canonical "1h0m0s" form, so we normalize here.
func (*DeploymentSettings) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[DeploymentSettingsInput], error) {
	inputs, failures, err := infer.DefaultCheck[DeploymentSettingsInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[DeploymentSettingsInput]{}, err
	}
	if inputs.OperationContext != nil &&
		inputs.OperationContext.Oidc != nil &&
		inputs.OperationContext.Oidc.AWS != nil &&
		inputs.OperationContext.Oidc.AWS.Duration != "" {
		normalized, normErr := normalizeDurationString(inputs.OperationContext.Oidc.AWS.Duration)
		if normErr != nil {
			failures = append(failures, p.CheckFailure{
				Property: "operationContext.oidc.aws.duration",
				Reason: fmt.Sprintf(
					"Failed to normalize duration string due to error: %s", normErr.Error()),
			})
		} else {
			inputs.OperationContext.Oidc.AWS.Duration = *normalized
		}
	}
	return infer.CheckResponse[DeploymentSettingsInput]{Inputs: inputs, Failures: failures}, nil
}

func (*DeploymentSettings) Create(
	ctx context.Context, req infer.CreateRequest[DeploymentSettingsInput],
) (infer.CreateResponse[DeploymentSettingsState], error) {
	id := deploymentSettingsResourceID(req.Inputs.Organization, req.Inputs.Project, req.Inputs.Stack)
	if req.DryRun {
		return infer.CreateResponse[DeploymentSettingsState]{
			ID:     id,
			Output: DeploymentSettingsState{DeploymentSettingsInput: req.Inputs},
		}, nil
	}
	settings, err := req.Inputs.toAPIDeploymentSettings()
	if err != nil {
		return infer.CreateResponse[DeploymentSettingsState]{}, err
	}
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.Inputs.Organization,
		ProjectName: req.Inputs.Project,
		StackName:   req.Inputs.Stack,
	}
	if _, err := config.GetClient(ctx).CreateDeploymentSettings(ctx, stack, settings); err != nil {
		return infer.CreateResponse[DeploymentSettingsState]{}, err
	}
	return infer.CreateResponse[DeploymentSettingsState]{
		ID:     id,
		Output: DeploymentSettingsState{DeploymentSettingsInput: req.Inputs},
	}, nil
}

func (*DeploymentSettings) Update(
	ctx context.Context, req infer.UpdateRequest[DeploymentSettingsInput, DeploymentSettingsState],
) (infer.UpdateResponse[DeploymentSettingsState], error) {
	if req.DryRun {
		return infer.UpdateResponse[DeploymentSettingsState]{
			Output: DeploymentSettingsState{DeploymentSettingsInput: req.Inputs},
		}, nil
	}
	settings, err := req.Inputs.toAPIDeploymentSettings()
	if err != nil {
		return infer.UpdateResponse[DeploymentSettingsState]{}, err
	}
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.Inputs.Organization,
		ProjectName: req.Inputs.Project,
		StackName:   req.Inputs.Stack,
	}
	if _, err := config.GetClient(ctx).UpdateDeploymentSettings(ctx, stack, settings); err != nil {
		return infer.UpdateResponse[DeploymentSettingsState]{}, err
	}
	return infer.UpdateResponse[DeploymentSettingsState]{
		Output: DeploymentSettingsState{DeploymentSettingsInput: req.Inputs},
	}, nil
}

func (*DeploymentSettings) Delete(
	ctx context.Context, req infer.DeleteRequest[DeploymentSettingsState],
) (infer.DeleteResponse, error) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.State.Organization,
		ProjectName: req.State.Project,
		StackName:   req.State.Stack,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteDeploymentSettings(ctx, stack)
}

func (*DeploymentSettings) Read(
	ctx context.Context, req infer.ReadRequest[DeploymentSettingsInput, DeploymentSettingsState],
) (infer.ReadResponse[DeploymentSettingsInput, DeploymentSettingsState], error) {
	stack, err := pulumiapi.NewStackIdentifier(req.ID)
	if err != nil {
		return infer.ReadResponse[DeploymentSettingsInput, DeploymentSettingsState]{}, err
	}
	settings, err := config.GetClient(ctx).GetDeploymentSettings(ctx, stack)
	if err != nil {
		return infer.ReadResponse[DeploymentSettingsInput, DeploymentSettingsState]{}, err
	}
	if settings == nil {
		return infer.ReadResponse[DeploymentSettingsInput, DeploymentSettingsState]{}, nil
	}
	inputs := deploymentSettingsInputFromAPI(stack, settings, req.Inputs)
	return infer.ReadResponse[DeploymentSettingsInput, DeploymentSettingsState]{
		ID:     req.ID,
		Inputs: inputs,
		State:  DeploymentSettingsState{DeploymentSettingsInput: inputs},
	}, nil
}

// --- Conversion helpers ---

func (i *DeploymentSettingsInput) toAPIDeploymentSettings() (pulumiapi.DeploymentSettings, error) {
	out := pulumiapi.DeploymentSettings{
		AgentPoolID: i.AgentPoolID,
	}
	if i.ExecutorContext != nil {
		out.Executor = &pulumiapi.ExecutorContext{
			ExecutorImage: &pulumiapi.DockerImage{Reference: i.ExecutorContext.ExecutorImage},
		}
	}
	if i.SourceContext != nil {
		out.SourceContext = i.SourceContext.toAPI()
	}
	if i.GitHub != nil {
		gh := pulumiapi.DeploymentSettingsGitHub{
			Repository:          i.GitHub.Repository,
			DeployCommits:       i.GitHub.DeployCommits,
			PreviewPullRequests: i.GitHub.PreviewPullRequests,
			PullRequestTemplate: i.GitHub.PullRequestTemplate,
			Paths:               i.GitHub.Paths,
		}
		out.GitHub = &gh
	}
	if i.Vcs != nil {
		vcs, err := i.Vcs.toAPI()
		if err != nil {
			return pulumiapi.DeploymentSettings{}, err
		}
		out.Vcs = vcs
	}
	if i.OperationContext != nil {
		out.Operation = i.OperationContext.toAPI()
	}
	if i.CacheOptions != nil {
		out.CacheOptions = &pulumiapi.CacheOptions{Enable: i.CacheOptions.Enable}
	}
	return out, nil
}

func (s *DeploymentSettingsSourceContext) toAPI() *pulumiapi.SourceContext {
	api := &pulumiapi.SourceContext{}
	if s.Git != nil {
		g := &pulumiapi.SourceContextGit{
			RepoURL: s.Git.RepoURL,
			Branch:  s.Git.Branch,
			Commit:  s.Git.Commit,
			RepoDir: s.Git.RepoDir,
		}
		if s.Git.GitAuth != nil {
			gauth := &pulumiapi.GitAuthConfig{}
			if s.Git.GitAuth.SSHAuth != nil {
				ssh := &pulumiapi.SSHAuth{
					SSHPrivateKey: pulumiapi.SecretValue{
						Value:  s.Git.GitAuth.SSHAuth.SSHPrivateKey,
						Secret: true,
					},
				}
				if s.Git.GitAuth.SSHAuth.Password != nil {
					ssh.Password = &pulumiapi.SecretValue{
						Value:  *s.Git.GitAuth.SSHAuth.Password,
						Secret: true,
					}
				}
				gauth.SSHAuth = ssh
			}
			if s.Git.GitAuth.BasicAuth != nil {
				gauth.BasicAuth = &pulumiapi.BasicAuth{
					UserName: pulumiapi.SecretValue{
						Value:  s.Git.GitAuth.BasicAuth.UserName,
						Secret: true,
					},
					Password: pulumiapi.SecretValue{
						Value:  s.Git.GitAuth.BasicAuth.Password,
						Secret: true,
					},
				}
			}
			g.GitAuth = gauth
		}
		api.Git = g
	}
	return api
}

func (v *DeploymentSettingsVcsBlock) toAPI() (pulumiapi.DeploymentSettingsVCS, error) {
	var vcs pulumiapi.DeploymentSettingsVCS
	switch v.Provider {
	case "azure_devops":
		vcs = pulumiapi.DeploymentSettingsVCSAzureDevOpsBuilder{}.Build()
	case "bitbucket":
		vcs = pulumiapi.DeploymentSettingsVCSBitbucketBuilder{}.Build()
	case "custom":
		vcs = pulumiapi.DeploymentSettingsVCSCustomBuilder{}.Build()
	case "github":
		vcs = pulumiapi.DeploymentSettingsVCSGitHubBuilder{}.Build()
	case "gitlab":
		vcs = pulumiapi.DeploymentSettingsVCSGitLabBuilder{}.Build()
	default:
		return nil, fmt.Errorf("unsupported VCS provider %q", v.Provider)
	}
	if v.Repository != "" {
		_ = vcs.SetRepository(v.Repository)
	}
	if v.InstallationID != "" {
		_ = vcs.SetInstallationID(v.InstallationID)
	}
	_ = vcs.SetDeployCommits(v.DeployCommits)
	_ = vcs.SetPreviewPullRequests(v.PreviewPullRequests)
	_ = vcs.SetPullRequestTemplate(v.PullRequestTemplate)
	if len(v.Paths) > 0 {
		_ = vcs.SetPaths(v.Paths)
	}
	if v.DeployPullRequest != nil {
		dpr := int64(*v.DeployPullRequest)
		_ = vcs.SetDeployPullRequest(&dpr)
	}
	return vcs, nil
}

func (o *DeploymentSettingsOperationContext) toAPI() *pulumiapi.OperationContext {
	api := &pulumiapi.OperationContext{
		PreRunCommands: o.PreRunCommands,
	}
	if len(o.EnvironmentVariables) > 0 {
		ev := make(map[string]pulumiapi.SecretValue, len(o.EnvironmentVariables))
		for k, val := range o.EnvironmentVariables {
			ev[k] = pulumiapi.SecretValue{Value: val}
		}
		api.EnvironmentVariables = ev
	}
	if o.Options != nil {
		api.Options = &pulumiapi.OperationContextOptions{
			SkipInstallDependencies:     o.Options.SkipInstallDependencies,
			SkipIntermediateDeployments: o.Options.SkipIntermediateDeployments,
			Shell:                       o.Options.Shell,
			DeleteAfterDestroy:          o.Options.DeleteAfterDestroy,
		}
	}
	if o.Oidc != nil {
		oidc := &pulumiapi.OperationContextOIDCConfiguration{}
		if o.Oidc.AWS != nil {
			oidc.AWS = &pulumiapi.OperationContextAWSOIDCConfiguration{
				Duration:    o.Oidc.AWS.Duration,
				PolicyARNs:  o.Oidc.AWS.PolicyARNs,
				RoleARN:     o.Oidc.AWS.RoleARN,
				SessionName: o.Oidc.AWS.SessionName,
			}
		}
		if o.Oidc.GCP != nil {
			oidc.GCP = &pulumiapi.OperationContextGCPOIDCConfiguration{
				ProjectID:      o.Oidc.GCP.ProjectID,
				Region:         o.Oidc.GCP.Region,
				WorkloadPoolID: o.Oidc.GCP.WorkloadPoolID,
				ProviderID:     o.Oidc.GCP.ProviderID,
				ServiceAccount: o.Oidc.GCP.ServiceAccount,
				TokenLifetime:  o.Oidc.GCP.TokenLifetime,
			}
		}
		if o.Oidc.Azure != nil {
			oidc.Azure = &pulumiapi.OperationContextAzureOIDCConfiguration{
				TenantID:       o.Oidc.Azure.TenantID,
				ClientID:       o.Oidc.Azure.ClientID,
				SubscriptionID: o.Oidc.Azure.SubscriptionID,
			}
		}
		api.OIDC = oidc
	}
	return api
}

// deploymentSettingsInputFromAPI rebuilds the input struct from the API response.
// Secret values returned by the API are ciphertext — the legacy implementation surfaced these
// directly. We preserve the previously-known plaintext from `previousInputs` when the API
// returns no plaintext, so a refresh on an unchanged secret does not appear to drift.
func deploymentSettingsInputFromAPI(
	stack pulumiapi.StackIdentifier,
	settings *pulumiapi.DeploymentSettings,
	previousInputs DeploymentSettingsInput,
) DeploymentSettingsInput {
	out := DeploymentSettingsInput{
		Organization: stack.OrgName,
		Project:      stack.ProjectName,
		Stack:        stack.StackName,
		AgentPoolID:  settings.AgentPoolID,
	}
	if settings.Executor != nil && settings.Executor.ExecutorImage != nil &&
		settings.Executor.ExecutorImage.Reference != "" {
		out.ExecutorContext = &DeploymentSettingsExecutorContext{
			ExecutorImage: settings.Executor.ExecutorImage.Reference,
		}
	}
	if settings.SourceContext != nil {
		out.SourceContext = sourceContextFromAPI(settings.SourceContext, previousInputs.SourceContext)
	}
	if settings.GitHub != nil {
		out.GitHub = &DeploymentSettingsGitHubBlock{
			Repository:          settings.GitHub.Repository,
			DeployCommits:       settings.GitHub.DeployCommits,
			PreviewPullRequests: settings.GitHub.PreviewPullRequests,
			PullRequestTemplate: settings.GitHub.PullRequestTemplate,
			Paths:               settings.GitHub.Paths,
		}
	}
	if settings.Vcs != nil {
		provider, _ := settings.Vcs.GetDiscriminatorValue()
		vcs := &DeploymentSettingsVcsBlock{
			Provider:            provider,
			Repository:          settings.Vcs.Repository(),
			InstallationID:      settings.Vcs.InstallationID(),
			DeployCommits:       settings.Vcs.DeployCommits(),
			PreviewPullRequests: settings.Vcs.PreviewPullRequests(),
			PullRequestTemplate: settings.Vcs.PullRequestTemplate(),
			Paths:               settings.Vcs.Paths(),
		}
		if dpr := settings.Vcs.DeployPullRequest(); dpr != nil {
			dprInt := int(*dpr)
			vcs.DeployPullRequest = &dprInt
		}
		out.Vcs = vcs
	}
	if settings.Operation != nil {
		out.OperationContext = operationContextFromAPI(settings.Operation, previousInputs.OperationContext)
	}
	if settings.CacheOptions != nil {
		out.CacheOptions = &DeploymentSettingsCacheOptions{Enable: settings.CacheOptions.Enable}
	}
	return out
}

func sourceContextFromAPI(
	api *pulumiapi.SourceContext, prev *DeploymentSettingsSourceContext,
) *DeploymentSettingsSourceContext {
	if api.Git == nil {
		return &DeploymentSettingsSourceContext{}
	}
	g := &DeploymentSettingsGitSource{
		RepoURL: api.Git.RepoURL,
		Branch:  api.Git.Branch,
		Commit:  api.Git.Commit,
		RepoDir: api.Git.RepoDir,
	}
	var prevAuth *DeploymentSettingsGitSourceGitAuth
	if prev != nil && prev.Git != nil {
		prevAuth = prev.Git.GitAuth
	}
	switch {
	case prevAuth != nil:
		// Git auth is write-only secret material: Pulumi Cloud stores it but does
		// not faithfully echo it on read (omitting it, or returning only
		// ciphertext). The previously-declared value is therefore authoritative —
		// preferring it keeps refresh from reporting a spurious `sourceContext`
		// diff. Genuine credential changes still flow through Create/Update.
		g.GitAuth = prevAuth
	case api.Git.GitAuth != nil:
		// Import: no prior state to fall back on, so reconstruct from the API.
		g.GitAuth = gitAuthFromAPI(api.Git.GitAuth, nil)
	}
	return &DeploymentSettingsSourceContext{Git: g}
}

func gitAuthFromAPI(
	api *pulumiapi.GitAuthConfig, prev *DeploymentSettingsGitSourceGitAuth,
) *DeploymentSettingsGitSourceGitAuth {
	out := &DeploymentSettingsGitSourceGitAuth{}
	if api.SSHAuth != nil {
		ssh := &DeploymentSettingsGitAuthSSHAuth{
			SSHPrivateKey: secretValueOrPrev(api.SSHAuth.SSHPrivateKey, prevSSHPrivateKey(prev)),
		}
		if api.SSHAuth.Password != nil {
			pw := secretValueOrPrev(*api.SSHAuth.Password, prevSSHPassword(prev))
			ssh.Password = &pw
		}
		out.SSHAuth = ssh
	}
	if api.BasicAuth != nil {
		out.BasicAuth = &DeploymentSettingsGitAuthBasicAuth{
			UserName: secretValueOrPrev(api.BasicAuth.UserName, prevBasicAuthUserName(prev)),
			Password: secretValueOrPrev(api.BasicAuth.Password, prevBasicAuthPassword(prev)),
		}
	}
	return out
}

// secretValueOrPrev reconciles a value returned by the Pulumi Cloud API with the
// previously-recorded plaintext from state. Non-secret values are echoed back as plaintext and
// are authoritative — using them surfaces out-of-band edits as drift. Secret values come back as
// ciphertext (`{"secret": "<ciphertext>"}` decodes to Secret=true with the ciphertext in Value),
// which carries no usable plaintext, so the known plaintext from state is kept instead. Only when
// no prior plaintext exists (import) is the ciphertext surfaced as-is.
func secretValueOrPrev(api pulumiapi.SecretValue, prev string) string {
	if api.Secret && prev != "" {
		return prev
	}
	if api.Value != "" {
		return api.Value
	}
	return prev
}

func prevSSHPrivateKey(prev *DeploymentSettingsGitSourceGitAuth) string {
	if prev == nil || prev.SSHAuth == nil {
		return ""
	}
	return prev.SSHAuth.SSHPrivateKey
}

func prevSSHPassword(prev *DeploymentSettingsGitSourceGitAuth) string {
	if prev == nil || prev.SSHAuth == nil || prev.SSHAuth.Password == nil {
		return ""
	}
	return *prev.SSHAuth.Password
}

func prevBasicAuthUserName(prev *DeploymentSettingsGitSourceGitAuth) string {
	if prev == nil || prev.BasicAuth == nil {
		return ""
	}
	return prev.BasicAuth.UserName
}

func prevBasicAuthPassword(prev *DeploymentSettingsGitSourceGitAuth) string {
	if prev == nil || prev.BasicAuth == nil {
		return ""
	}
	return prev.BasicAuth.Password
}

func operationContextFromAPI(
	api *pulumiapi.OperationContext, prev *DeploymentSettingsOperationContext,
) *DeploymentSettingsOperationContext {
	out := &DeploymentSettingsOperationContext{
		PreRunCommands: api.PreRunCommands,
	}
	if len(api.EnvironmentVariables) > 0 {
		ev := make(map[string]string, len(api.EnvironmentVariables))
		var prevEV map[string]string
		if prev != nil {
			prevEV = prev.EnvironmentVariables
		}
		for k, v := range api.EnvironmentVariables {
			ev[k] = secretValueOrPrev(v, prevEV[k])
		}
		out.EnvironmentVariables = ev
	}
	if api.Options != nil {
		out.Options = &DeploymentOperationOptions{
			SkipInstallDependencies:     api.Options.SkipInstallDependencies,
			SkipIntermediateDeployments: api.Options.SkipIntermediateDeployments,
			Shell:                       api.Options.Shell,
			DeleteAfterDestroy:          api.Options.DeleteAfterDestroy,
		}
	}
	if api.OIDC != nil {
		oidc := &DeploymentOperationOIDC{}
		if api.OIDC.AWS != nil {
			oidc.AWS = &DeploymentAWSOIDCConfiguration{
				Duration:    api.OIDC.AWS.Duration,
				PolicyARNs:  api.OIDC.AWS.PolicyARNs,
				RoleARN:     api.OIDC.AWS.RoleARN,
				SessionName: api.OIDC.AWS.SessionName,
			}
		}
		if api.OIDC.GCP != nil {
			oidc.GCP = &DeploymentGCPOIDCConfiguration{
				ProjectID:      api.OIDC.GCP.ProjectID,
				Region:         api.OIDC.GCP.Region,
				WorkloadPoolID: api.OIDC.GCP.WorkloadPoolID,
				ProviderID:     api.OIDC.GCP.ProviderID,
				ServiceAccount: api.OIDC.GCP.ServiceAccount,
				TokenLifetime:  api.OIDC.GCP.TokenLifetime,
			}
		}
		if api.OIDC.Azure != nil {
			oidc.Azure = &DeploymentAzureOIDCConfiguration{
				TenantID:       api.OIDC.Azure.TenantID,
				ClientID:       api.OIDC.Azure.ClientID,
				SubscriptionID: api.OIDC.Azure.SubscriptionID,
			}
		}
		out.Oidc = oidc
	}
	return out
}

// --- ID helpers ---

func deploymentSettingsResourceID(org, project, stack string) string {
	return path.Join(org, project, stack)
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
