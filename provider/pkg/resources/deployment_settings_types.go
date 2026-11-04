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
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/mapper"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// The structs below define the schema of pulumiservice:index:DeploymentSettings
// and its types. Each type is named after its schema token.
//
// A pointer on a field the schema marks required keeps "absent" apart from the
// zero value, because the request sent to Pulumi Cloud differs between the two.

type DeploymentSettingsInput struct {
	Organization     string                              `pulumi:"organization"              provider:"replaceOnChanges"`
	Project          string                              `pulumi:"project"                   provider:"replaceOnChanges"`
	Stack            string                              `pulumi:"stack"                     provider:"replaceOnChanges"`
	AgentPoolID      *string                             `pulumi:"agentPoolId,optional"`
	ExecutorContext  *DeploymentSettingsExecutorContext  `pulumi:"executorContext,optional"`
	SourceContext    *DeploymentSettingsSourceContext    `pulumi:"sourceContext,optional"`
	GitHub           *DeploymentSettingsGithub           `pulumi:"github,optional"`
	Vcs              *DeploymentSettingsVcs              `pulumi:"vcs,optional"`
	OperationContext *DeploymentSettingsOperationContext `pulumi:"operationContext,optional"`
	CacheOptions     *DeploymentSettingsCacheOptions     `pulumi:"cacheOptions,optional"`
}

func (in *DeploymentSettingsInput) Annotate(a infer.Annotator) {
	a.Describe(&in.Organization, "Organization name.")
	a.Describe(&in.Project, "Project name.")
	a.Describe(&in.Stack, "Stack name.")
	a.Describe(&in.AgentPoolID, "The agent pool identifier to use for the deployment.")
	a.Describe(&in.ExecutorContext, "Settings related to the deployment executor.")
	a.Describe(&in.SourceContext, "Settings related to the source of the deployment.")
	a.Describe(&in.GitHub, "GitHub settings for the deployment.")
	a.Deprecate(&in.GitHub, githubDeprecation)
	a.Describe(&in.Vcs, vcsDescription)
	a.Describe(&in.OperationContext, "Settings related to the Pulumi operation environment during the deployment.")
	a.Describe(&in.CacheOptions, "Dependency cache settings for the deployment")
}

// DeploymentSettingsState is the resource's output. It has the same properties
// as [DeploymentSettingsInput], without replaceOnChanges, which applies only to
// inputs.
type DeploymentSettingsState struct {
	Organization     string                              `pulumi:"organization"`
	Project          string                              `pulumi:"project"`
	Stack            string                              `pulumi:"stack"`
	AgentPoolID      *string                             `pulumi:"agentPoolId,optional"`
	ExecutorContext  *DeploymentSettingsExecutorContext  `pulumi:"executorContext,optional"`
	SourceContext    *DeploymentSettingsSourceContext    `pulumi:"sourceContext,optional"`
	GitHub           *DeploymentSettingsGithub           `pulumi:"github,optional"`
	Vcs              *DeploymentSettingsVcs              `pulumi:"vcs,optional"`
	OperationContext *DeploymentSettingsOperationContext `pulumi:"operationContext,optional"`
	CacheOptions     *DeploymentSettingsCacheOptions     `pulumi:"cacheOptions,optional"`
}

func (s *DeploymentSettingsState) Annotate(a infer.Annotator) {
	a.Describe(&s.Organization, "Organization name.")
	a.Describe(&s.Project, "Project name.")
	a.Describe(&s.Stack, "Stack name.")
	a.Describe(&s.AgentPoolID, "The agent pool identifier to use for the deployment.")
	a.Describe(&s.ExecutorContext, "Settings related to the deployment executor.")
	a.Describe(&s.SourceContext, "Settings related to the source of the deployment.")
	a.Describe(&s.GitHub, "GitHub settings for the deployment.")
	a.Deprecate(&s.GitHub, githubDeprecation)
	a.Describe(&s.Vcs, vcsDescription)
	a.Describe(&s.OperationContext, "Settings related to the Pulumi operation environment during the deployment.")
	a.Describe(&s.CacheOptions, "Dependency cache settings for the deployment")
}

const (
	githubDeprecation = "Use the 'vcs' property instead, which supports both GitHub and Azure DevOps."
	vcsDescription    = "VCS settings for the deployment. Supports Azure DevOps and GitHub via the 'provider' " +
		"discriminator field."
	escRecommendation     = "Pulumi ESC is the recommended way to supply cloud credentials to Pulumi Deployments."
	cloudCredentialsGuide = "[Supplying Cloud Credentials to Pulumi Deployments]" +
		"(https://www.pulumi.com/docs/deployments/guides/cloud-credentials/)"
)

type DeploymentSettingsExecutorContext struct {
	ExecutorImage *string                                     `pulumi:"executorImage"`
	Credentials   *DeploymentSettingsExecutorImageCredentials `pulumi:"credentials,optional"`
}

func (ec *DeploymentSettingsExecutorContext) Annotate(a infer.Annotator) {
	a.Describe(ec, "The executor context defines information about the executor where the deployment is "+
		"executed. If unspecified, the default 'pulumi/pulumi' image is used.")
	a.Describe(&ec.ExecutorImage, "Allows overriding the default executor image with a custom image. "+
		"E.g. 'pulumi/pulumi-nodejs:latest'")
	a.Describe(&ec.Credentials, "Credentials for pulling `executorImage` from a private container registry. "+
		"Only needed when the image is not publicly accessible.")
	a.SetToken("index", "DeploymentSettingsExecutorContext")
}

type DeploymentSettingsExecutorImageCredentials struct {
	Username string  `pulumi:"username"`
	Password *string `pulumi:"password" provider:"secret"`
}

func (c *DeploymentSettingsExecutorImageCredentials) Annotate(a infer.Annotator) {
	a.Describe(c, "Credentials for pulling the executor image from a private container registry.")
	a.Describe(&c.Username, "Username for authenticating with the container registry.")
	a.Describe(&c.Password, "Password or access token for authenticating with the container registry.")
	a.SetToken("index", "DeploymentSettingsExecutorImageCredentials")
}

type DeploymentSettingsSourceContext struct {
	Git *DeploymentSettingsGitSource `pulumi:"git,optional"`
}

func (sc *DeploymentSettingsSourceContext) Annotate(a infer.Annotator) {
	a.Describe(sc, "Settings related to the source of the deployment.")
	a.Describe(&sc.Git, "Git source settings for a deployment.")
	a.SetToken("index", "DeploymentSettingsSourceContext")
}

type DeploymentSettingsGitSource struct {
	RepoURL string                              `pulumi:"repoUrl,optional"`
	GitAuth *DeploymentSettingsGitSourceGitAuth `pulumi:"gitAuth,optional"`
	Branch  string                              `pulumi:"branch,optional"`
	Commit  string                              `pulumi:"commit,optional"`
	RepoDir string                              `pulumi:"repoDir,optional"`
}

func (g *DeploymentSettingsGitSource) Annotate(a infer.Annotator) {
	a.Describe(g, "Git source settings for a deployment.")
	a.Describe(&g.RepoURL, "The repository URL to use for git settings. Should not be specified if there are "+
		"`gitHub` settings for this deployment.")
	a.Describe(&g.GitAuth, "Git authentication configuration for this deployment. Should not be specified if "+
		"there are `gitHub` settings for this deployment.")
	a.Describe(&g.Branch, "The branch to deploy. One of either `branch` or `commit` must be specified.")
	a.Describe(&g.Commit, "The commit to deploy. One of either `branch` or `commit` must be specified.")
	a.Describe(&g.RepoDir, "The directory within the repository where the Pulumi.yaml is located.")
	a.SetToken("index", "DeploymentSettingsGitSource")
}

type DeploymentSettingsGitSourceGitAuth struct {
	SSHAuth   *DeploymentSettingsGitAuthSSHAuth   `pulumi:"sshAuth,optional"`
	BasicAuth *DeploymentSettingsGitAuthBasicAuth `pulumi:"basicAuth,optional"`
}

func (g *DeploymentSettingsGitSourceGitAuth) Annotate(a infer.Annotator) {
	a.Describe(g, "Git source settings for a deployment.")
	a.Describe(&g.SSHAuth, "SSH auth for git authentication. Only one of `personalAccessToken`, `sshAuth`, "+
		"or `basicAuth` must be defined.")
	a.Describe(&g.BasicAuth, "Basic auth for git authentication. Only one of `personalAccessToken`, `sshAuth`, "+
		"or `basicAuth` must be defined.")
	a.SetToken("index", "DeploymentSettingsGitSourceGitAuth")
}

type DeploymentSettingsGitAuthSSHAuth struct {
	SSHPrivateKey *string `pulumi:"sshPrivateKey"     provider:"secret"`
	Password      *string `pulumi:"password,optional" provider:"secret"`
}

func (s *DeploymentSettingsGitAuthSSHAuth) Annotate(a infer.Annotator) {
	a.Describe(s, "Git source settings for a deployment.")
	a.Describe(&s.SSHPrivateKey, "SSH private key.")
	a.Describe(&s.Password, "Optional password for SSH authentication.")
	a.SetToken("index", "DeploymentSettingsGitAuthSSHAuth")
}

type DeploymentSettingsGitAuthBasicAuth struct {
	Username *string `pulumi:"username" provider:"secret"`
	Password *string `pulumi:"password" provider:"secret"`
}

func (b *DeploymentSettingsGitAuthBasicAuth) Annotate(a infer.Annotator) {
	a.Describe(b, "Git source settings for a deployment.")
	a.Describe(&b.Username, "User name for git basic authentication.")
	a.Describe(&b.Password, "Password for git basic authentication.")
	a.SetToken("index", "DeploymentSettingsGitAuthBasicAuth")
}

type DeploymentSettingsGithub struct {
	Repository          string   `pulumi:"repository,optional"`
	DeployCommits       bool     `pulumi:"deployCommits,optional"`
	PreviewPullRequests bool     `pulumi:"previewPullRequests,optional"`
	PullRequestTemplate bool     `pulumi:"pullRequestTemplate,optional"`
	Paths               []string `pulumi:"paths,optional"`
	DeployTags          bool     `pulumi:"deployTags,optional"`
	TagFilters          []string `pulumi:"tagFilters,optional"`
}

func (gh *DeploymentSettingsGithub) Annotate(a infer.Annotator) {
	a.Describe(gh, "GitHub settings for the deployment.")
	a.Describe(&gh.Repository, "The GitHub repository in the format org/repo.")
	a.Describe(&gh.DeployCommits, "Trigger a deployment running `pulumi up` on commit. Must be `false` when "+
		"`deployTags` is `true`, because Pulumi Cloud rejects the two together.")
	a.Describe(&gh.PreviewPullRequests, "Trigger a deployment running `pulumi preview` when a PR is opened.")
	a.Describe(&gh.PullRequestTemplate, "Use this stack as a template for pull request review stacks.")
	a.Describe(&gh.Paths, "The paths within the repo that deployments should be filtered to.")
	a.Describe(&gh.DeployTags, "Trigger a deployment running `pulumi up` when a tag matching `tagFilters` is "+
		"pushed. Can't be combined with `deployCommits`, which defaults to `true`, so set `deployCommits` to "+
		"`false` when you enable this.")
	a.Describe(&gh.TagFilters, "Glob patterns matching the tag names that trigger a deployment when "+
		"`deployTags` is true. Prefix a pattern with `!` to exclude matching tags, for example `!*-rc*`. An "+
		"empty list matches all tags.")
	a.SetDefault(&gh.DeployCommits, true)
	a.SetDefault(&gh.PreviewPullRequests, true)
	a.SetDefault(&gh.PullRequestTemplate, false)
	a.SetDefault(&gh.DeployTags, false)
	a.SetToken("index", "DeploymentSettingsGithub")
}

type DeploymentSettingsVcs struct {
	Provider            string   `pulumi:"provider"`
	Repository          string   `pulumi:"repository,optional"`
	InstallationID      string   `pulumi:"installationId,optional"`
	DeployCommits       bool     `pulumi:"deployCommits,optional"`
	PreviewPullRequests bool     `pulumi:"previewPullRequests,optional"`
	PullRequestTemplate bool     `pulumi:"pullRequestTemplate,optional"`
	Paths               []string `pulumi:"paths,optional"`
	DeployTags          bool     `pulumi:"deployTags,optional"`
	TagFilters          []string `pulumi:"tagFilters,optional"`
	DeployPullRequest   *int     `pulumi:"deployPullRequest,optional"`
}

func (v *DeploymentSettingsVcs) Annotate(a infer.Annotator) {
	a.Describe(v, "VCS settings for the deployment, supporting multiple VCS providers.")
	a.Describe(&v.Provider, "The VCS provider type.")
	a.Describe(&v.Repository, "The repository identifier (e.g., 'ProjectName/RepoName' for Azure DevOps, "+
		"'org/repo' for GitHub).")
	a.Describe(&v.InstallationID, "The VCS integration installation ID. Use to disambiguate when an "+
		"organization has multiple integrations of the same provider type (e.g., two GitHub Apps). If omitted, "+
		"the API resolves the integration automatically from `provider` and `repository`.")
	a.Describe(&v.DeployCommits, "Trigger a deployment running `pulumi up` on commit. Must be `false` when "+
		"`deployTags` is `true`, because Pulumi Cloud rejects the two together.")
	a.Describe(&v.PreviewPullRequests, "Trigger a deployment running `pulumi preview` when a PR is opened.")
	a.Describe(&v.PullRequestTemplate, "Use this stack as a template for pull request review stacks.")
	a.Describe(&v.Paths, "The paths within the repo that deployments should be filtered to.")
	a.Describe(&v.DeployTags, "Trigger a deployment running `pulumi up` when a tag matching `tagFilters` is "+
		"pushed. Can't be combined with `deployCommits`, which defaults to `true`, so set `deployCommits` to "+
		"`false` when you enable this.")
	a.Describe(&v.TagFilters, "Glob patterns matching the tag names that trigger a deployment when "+
		"`deployTags` is true. Prefix a pattern with `!` to exclude matching tags, for example `!*-rc*`. An "+
		"empty list matches all tags.")
	a.Describe(&v.DeployPullRequest, "Deploy a specific pull request number.")
	a.SetDefault(&v.DeployCommits, true)
	a.SetDefault(&v.PreviewPullRequests, true)
	a.SetDefault(&v.PullRequestTemplate, false)
	a.SetDefault(&v.DeployTags, false)
	a.SetToken("index", "DeploymentSettingsVcs")
}

type DeploymentSettingsOperationContext struct {
	PreRunCommands       []string                 `pulumi:"preRunCommands,optional"`
	EnvironmentVariables map[string]string        `pulumi:"environmentVariables,optional"`
	Options              *OperationContextOptions `pulumi:"options,optional"`
	OIDC                 *OperationContextOIDC    `pulumi:"oidc,optional"`
}

func (oc *DeploymentSettingsOperationContext) Annotate(a infer.Annotator) {
	a.Describe(oc, "Settings related to the Pulumi operation environment during the deployment.")
	a.Describe(&oc.PreRunCommands, "Shell commands to run before the Pulumi operation executes.")
	a.Describe(&oc.EnvironmentVariables, "Environment variables to set for the deployment.")
	a.Describe(&oc.Options, "Options to override default behavior during the deployment.")
	a.Describe(&oc.OIDC, "Deployments OIDC configuration for obtaining cloud credentials during the "+
		"deployment. "+escRecommendation+" Deployments OIDC still fits when a customer-managed runner can't "+
		"reach Pulumi Cloud to use ESC, or when pre-run commands need credentials and aren't wrapped in "+
		"`pulumi env run`. See "+cloudCredentialsGuide+".")
	a.SetToken("index", "DeploymentSettingsOperationContext")
}

type OperationContextOptions struct {
	SkipInstallDependencies     bool   `pulumi:"skipInstallDependencies,optional"`
	SkipIntermediateDeployments bool   `pulumi:"skipIntermediateDeployments,optional"`
	Shell                       string `pulumi:"shell,optional"`
	DeleteAfterDestroy          bool   `pulumi:"deleteAfterDestroy,optional"`
}

func (o *OperationContextOptions) Annotate(a infer.Annotator) {
	a.Describe(&o.SkipInstallDependencies, "Skip the default dependency installation step - use this to "+
		"customize the dependency installation (e.g. if using yarn or poetry)")
	a.Describe(&o.SkipIntermediateDeployments, "Skip intermediate deployments (Consolidate multiple "+
		"deployments of the same type into one deployment)")
	a.Describe(&o.Shell, "The shell to use to run commands during the deployment. Defaults to 'bash'.")
	a.Describe(&o.DeleteAfterDestroy, "Whether the stack should be deleted after it is destroyed.")
	a.SetToken("index", "OperationContextOptions")
}

type OperationContextOIDC struct {
	AWS   *AWSOIDCConfiguration   `pulumi:"aws,optional"`
	GCP   *GCPOIDCConfiguration   `pulumi:"gcp,optional"`
	Azure *AzureOIDCConfiguration `pulumi:"azure,optional"`
}

func (o *OperationContextOIDC) Annotate(a infer.Annotator) {
	a.Describe(o, "Deployments OIDC configuration for obtaining cloud credentials during a deployment.\n\n"+
		escRecommendation+" Deployments OIDC predates Pulumi ESC and still works, but for most users ESC is "+
		"the better choice. Deployments OIDC is still a good fit when:\n\n"+
		"- A customer-managed runner can't reach Pulumi Cloud over the network to use ESC.\n"+
		"- Pre-run commands need credentials. An ESC environment applies only to the Pulumi operation itself, "+
		"unless the command is wrapped in `pulumi env run`.\n\n"+
		"See "+cloudCredentialsGuide+".")
	a.Describe(&o.AWS, "AWS-specific OIDC configuration.")
	a.Describe(&o.GCP, "GCP-specific OIDC configuration.")
	a.Describe(&o.Azure, "Azure-specific OIDC configuration.")
	a.SetToken("index", "OperationContextOIDC")
}

type AWSOIDCConfiguration struct {
	Duration    string   `pulumi:"duration,optional"`
	PolicyARNs  []string `pulumi:"policyARNs,optional"`
	RoleARN     string   `pulumi:"roleARN"`
	SessionName string   `pulumi:"sessionName"`
}

func (o *AWSOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(o, "AWS configuration for Deployments OIDC. "+escRecommendation+" See "+cloudCredentialsGuide+".")
	a.Describe(&o.Duration, "Duration of the assume-role session in “XhYmZs” format")
	a.Describe(&o.PolicyARNs, "Optional set of IAM policy ARNs that further restrict the assume-role session")
	a.Describe(&o.RoleARN, "The ARN of the role to assume using the OIDC token.")
	a.Describe(&o.SessionName, "The name of the assume-role session.")
	a.SetToken("index", "AWSOIDCConfiguration")
}

type GCPOIDCConfiguration struct {
	ProjectID      string `pulumi:"projectId"`
	Region         string `pulumi:"region,optional"`
	WorkloadPoolID string `pulumi:"workloadPoolId"`
	ProviderID     string `pulumi:"providerId"`
	ServiceAccount string `pulumi:"serviceAccount"`
	TokenLifetime  string `pulumi:"tokenLifetime,optional"`
}

func (o *GCPOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(o, "GCP configuration for Deployments OIDC. "+escRecommendation+" See "+cloudCredentialsGuide+".")
	a.Describe(&o.ProjectID, "The GCP project number, a numeric value such as `123456789012`. This is not "+
		"the project ID (such as `my-project`). To look up the number, run "+
		"`gcloud projects describe <project-id> --format='value(projectNumber)'`.")
	a.Describe(&o.Region, "The region of the GCP project.")
	a.Describe(&o.WorkloadPoolID, "The ID of the workload pool to use.")
	a.Describe(&o.ProviderID, "The ID of the identity provider associated with the workload pool.")
	a.Describe(&o.ServiceAccount, "The email address of the service account to use.")
	a.Describe(&o.TokenLifetime, "The lifetime of the temporary credentials in “XhYmZs” format.")
	a.SetToken("index", "GCPOIDCConfiguration")
}

type AzureOIDCConfiguration struct {
	ClientID       string `pulumi:"clientId"`
	TenantID       string `pulumi:"tenantId"`
	SubscriptionID string `pulumi:"subscriptionId"`
}

func (o *AzureOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(o, "Azure configuration for Deployments OIDC. "+escRecommendation+" See "+cloudCredentialsGuide+".")
	a.Describe(&o.ClientID, "The client ID of the federated workload identity.")
	a.Describe(&o.TenantID, "The tenant ID of the federated workload identity.")
	a.Describe(&o.SubscriptionID, "The subscription ID of the federated workload identity.")
	a.SetToken("index", "AzureOIDCConfiguration")
}

type DeploymentSettingsCacheOptions struct {
	Enable bool `pulumi:"enable,optional"`
}

func (c *DeploymentSettingsCacheOptions) Annotate(a infer.Annotator) {
	a.Describe(c, "Dependency cache settings for the deployment")
	a.Describe(&c.Enable, "Enable dependency caching")
	a.SetDefault(&c.Enable, false)
	a.SetToken("index", "DeploymentSettingsCacheOptions")
}

// decodeDeploymentSettingsInput decodes a DeploymentSettings property map, with
// secrets kept, into the settings sent to Pulumi Cloud.
func decodeDeploymentSettingsInput(props resource.PropertyMap) (PulumiServiceDeploymentSettingsInput, error) {
	var in DeploymentSettingsInput
	// Like the hand-written decoder this replaced, tolerate missing and unknown keys.
	if err := mapper.MapI(props.MapRepl(nil, plainValue), &in); err != nil {
		return PulumiServiceDeploymentSettingsInput{}, err
	}
	return in.toAPI(secretEnvironmentVariables(props)), nil
}

// plainValue is a MapRepl replacer that strips secret wrappers at any depth.
func plainValue(v resource.PropertyValue) (any, bool) {
	if v.IsSecret() {
		return v.SecretValue().Element.MapRepl(nil, plainValue), true
	}
	return nil, false
}

// secretEnvironmentVariables reports which operationContext.environmentVariables
// are secret. A secret operationContext or environmentVariables makes them all secret.
func secretEnvironmentVariables(props resource.PropertyMap) map[string]bool {
	oc := props["operationContext"]
	if !oc.HasValue() {
		return nil
	}
	ev := unwrapSecret(oc).ObjectValue()["environmentVariables"]
	if !ev.HasValue() {
		return nil
	}
	cascade := oc.IsSecret() || ev.IsSecret()
	secrets := map[string]bool{}
	for k, v := range unwrapSecret(ev).ObjectValue() {
		secrets[string(k)] = cascade || v.IsSecret()
	}
	return secrets
}

func unwrapSecret(v resource.PropertyValue) resource.PropertyValue {
	if v.IsSecret() {
		return v.SecretValue().Element
	}
	return v
}

// secretValue flags a value as secret, which Pulumi Cloud needs in order to
// encrypt it rather than store it in plaintext. A missing value stays zero.
func secretValue(v *string) pulumiapi.SecretValue {
	if v == nil {
		return pulumiapi.SecretValue{}
	}
	return pulumiapi.SecretValue{Secret: true, Value: *v}
}

func (in *DeploymentSettingsInput) toAPI(secretEnvVars map[string]bool) PulumiServiceDeploymentSettingsInput {
	out := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     in.Organization,
			ProjectName: in.Project,
			StackName:   in.Stack,
		},
	}
	out.AgentPoolID = in.AgentPoolID
	out.Executor = in.ExecutorContext.toAPI()
	out.GitHub = in.GitHub.toAPI()
	out.SourceContext = in.SourceContext.toAPI()
	out.Operation = in.OperationContext.toAPI(secretEnvVars)
	if in.CacheOptions != nil {
		out.CacheOptions = &pulumiapi.CacheOptions{Enable: in.CacheOptions.Enable}
	}
	out.Vcs = in.Vcs.toAPI()
	return out
}

func (ec *DeploymentSettingsExecutorContext) toAPI() *pulumiapi.ExecutorContext {
	if ec == nil {
		return nil
	}
	var out pulumiapi.ExecutorContext
	if ec.ExecutorImage != nil {
		out.ExecutorImage = &pulumiapi.DockerImage{Reference: *ec.ExecutorImage}
		if ec.Credentials != nil {
			out.ExecutorImage.Credentials = &pulumiapi.DockerImageCredentials{
				Username: ec.Credentials.Username,
				Password: secretValue(ec.Credentials.Password),
			}
		}
	}
	return &out
}

func (gh *DeploymentSettingsGithub) toAPI() *pulumiapi.DeploymentSettingsGitHub {
	if gh == nil {
		return nil
	}
	return &pulumiapi.DeploymentSettingsGitHub{
		Repository:          gh.Repository,
		DeployCommits:       gh.DeployCommits,
		PreviewPullRequests: gh.PreviewPullRequests,
		PullRequestTemplate: gh.PullRequestTemplate,
		Paths:               gh.Paths,
		DeployTags:          gh.DeployTags,
		TagFilters:          gh.TagFilters,
	}
}

func (v *DeploymentSettingsVcs) toAPI() pulumiapi.DeploymentSettingsVCS {
	if v == nil {
		return nil
	}
	base := pulumiapi.DeploymentSettingsVCSBuilder{
		Repository:          v.Repository,
		InstallationID:      v.InstallationID,
		DeployCommits:       v.DeployCommits,
		PreviewPullRequests: v.PreviewPullRequests,
		PullRequestTemplate: v.PullRequestTemplate,
		Paths:               v.Paths,
		DeployTags:          v.DeployTags,
		TagFilters:          v.TagFilters,
	}
	if v.DeployPullRequest != nil {
		pr := int64(*v.DeployPullRequest)
		base.DeployPullRequest = &pr
	}
	switch v.Provider {
	case "azure_devops":
		return pulumiapi.DeploymentSettingsVCSAzureDevOpsBuilder{DeploymentSettingsVCSBuilder: base}.Build()
	case "bitbucket":
		return pulumiapi.DeploymentSettingsVCSBitbucketBuilder{DeploymentSettingsVCSBuilder: base}.Build()
	case "custom":
		return pulumiapi.DeploymentSettingsVCSCustomBuilder{DeploymentSettingsVCSBuilder: base}.Build()
	case "github":
		return pulumiapi.DeploymentSettingsVCSGitHubBuilder{DeploymentSettingsVCSBuilder: base}.Build()
	case "gitlab":
		return pulumiapi.DeploymentSettingsVCSGitLabBuilder{DeploymentSettingsVCSBuilder: base}.Build()
	default:
		return nil
	}
}

func (sc *DeploymentSettingsSourceContext) toAPI() *pulumiapi.SourceContext {
	if sc == nil {
		return nil
	}
	var out pulumiapi.SourceContext
	if g := sc.Git; g != nil {
		out.Git = &pulumiapi.SourceContextGit{
			RepoURL: g.RepoURL,
			Branch:  g.Branch,
			Commit:  g.Commit,
			RepoDir: g.RepoDir,
		}
		if a := g.GitAuth; a != nil {
			var auth pulumiapi.GitAuthConfig
			if s := a.SSHAuth; s != nil {
				auth.SSHAuth = &pulumiapi.SSHAuth{SSHPrivateKey: secretValue(s.SSHPrivateKey)}
				if s.Password != nil {
					password := secretValue(s.Password)
					auth.SSHAuth.Password = &password
				}
			}
			if b := a.BasicAuth; b != nil {
				auth.BasicAuth = &pulumiapi.BasicAuth{
					UserName: secretValue(b.Username),
					Password: secretValue(b.Password),
				}
			}
			out.Git.GitAuth = &auth
		}
	}
	return &out
}

func (oc *DeploymentSettingsOperationContext) toAPI(secretEnvVars map[string]bool) *pulumiapi.OperationContext {
	if oc == nil {
		return nil
	}
	out := pulumiapi.OperationContext{PreRunCommands: oc.PreRunCommands}
	if oc.EnvironmentVariables != nil {
		out.EnvironmentVariables = map[string]pulumiapi.SecretValue{}
		for k, v := range oc.EnvironmentVariables {
			out.EnvironmentVariables[k] = pulumiapi.SecretValue{Secret: secretEnvVars[k], Value: v}
		}
	}
	if o := oc.Options; o != nil {
		// Shell is never sent: the hand-written decoder this replaced read the
		// misspelled key "Shell", so it was always dropped.
		out.Options = &pulumiapi.OperationContextOptions{
			SkipInstallDependencies:     o.SkipInstallDependencies,
			SkipIntermediateDeployments: o.SkipIntermediateDeployments,
			DeleteAfterDestroy:          o.DeleteAfterDestroy,
		}
	}
	if oidc := oc.OIDC; oidc != nil {
		out.OIDC = &pulumiapi.OperationContextOIDCConfiguration{}
		if aws := oidc.AWS; aws != nil {
			out.OIDC.AWS = &pulumiapi.OperationContextAWSOIDCConfiguration{
				RoleARN:     aws.RoleARN,
				Duration:    aws.Duration,
				SessionName: aws.SessionName,
				PolicyARNs:  aws.PolicyARNs,
			}
		}
		if gcp := oidc.GCP; gcp != nil {
			out.OIDC.GCP = &pulumiapi.OperationContextGCPOIDCConfiguration{
				ProjectID:      gcp.ProjectID,
				Region:         gcp.Region,
				WorkloadPoolID: gcp.WorkloadPoolID,
				ProviderID:     gcp.ProviderID,
				ServiceAccount: gcp.ServiceAccount,
				TokenLifetime:  gcp.TokenLifetime,
			}
		}
		if azure := oidc.Azure; azure != nil {
			out.OIDC.Azure = &pulumiapi.OperationContextAzureOIDCConfiguration{
				TenantID:       azure.TenantID,
				ClientID:       azure.ClientID,
				SubscriptionID: azure.SubscriptionID,
			}
		}
	}
	return &out
}
