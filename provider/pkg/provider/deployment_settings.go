package provider

import (
	"path"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
)

var (
	// Life-cycle participation
	_ infer.CustomResource[DeploymentSettingsInput, DeploymentSettingsState] = (*DeploymentSettings)(nil)
	_ infer.CustomRead[DeploymentSettingsInput, DeploymentSettingsState]     = (*DeploymentSettings)(nil)
	_ infer.CustomDelete[DeploymentSettingsState]                            = (*DeploymentSettings)(nil)

	// Schema documentation
	_ infer.Annotated = (*DeploymentSettings)(nil)
	_ infer.Annotated = (*DeploymentSettingsInput)(nil)
	_ infer.Annotated = (*DeploymentSettingsState)(nil)
	_ infer.Annotated = (*DeploymentSettingsSourceContext)(nil)
	_ infer.Annotated = (*DeploymentSettingsExecutorContext)(nil)
	_ infer.Annotated = (*DeploymentSettingsGithub)(nil)
	_ infer.Annotated = (*DeploymentSettingsOperationContext)(nil)
	_ infer.Annotated = (*DeploymentSettingsGitSource)(nil)
)

type DeploymentSettings struct{}

func (p *DeploymentSettings) Annotate(a infer.Annotator) {
	a.Describe(p, "Deployment settings configure Pulumi Deployments for a stack.\n\n### Import\n\nDeployment settings can be imported using the `id`, which for deployment settings is `{org}/{project}/{stack}` e.g.,\n\n```sh\n $ pulumi import pulumiservice:index:DeploymentSettings my_settings my-org/my-project/my-stack\n```\n\n")
}

type DeploymentSettingsInput struct {
	// Required fields
	Organization  string                          `pulumi:"organization"`
	Stack         string                          `pulumi:"stack"`
	Project       string                          `pulumi:"project"`
	SourceContext DeploymentSettingsSourceContext `pulumi:"sourceContext"`

	// Optional fields
	ExecutorContext  *DeploymentSettingsExecutorContext  `pulumi:"executorContext,optional"`
	Github           *DeploymentSettingsGithub           `pulumi:"github,optional"`
	OperationContext *DeploymentSettingsOperationContext `pulumi:"operationContext,optional"`
	AgentPoolId      *string                             `pulumi:"agentPoolId,optional"`
}

func (p DeploymentSettingsInput) asStack() pulumiapi.StackName {
	return pulumiapi.StackName{
		OrgName:     p.Organization,
		ProjectName: p.Project,
		StackName:   p.Stack,
	}
}

func (p DeploymentSettingsInput) asAPI() pulumiapi.DeploymentSettings {
	return pulumiapi.DeploymentSettings{
		OperationContext: p.OperationContext.asAPI(),
		GitHub:           p.Github.asAPI(),
		SourceContext:    p.SourceContext.asAPI(),
		ExecutorContext:  p.ExecutorContext.asAPI(),
		AgentPoolId:      derefOrZero(p.AgentPoolId),
	}
}

func derefOrZero[T any](v *T) T {
	var t T
	if v != nil {
		t = *v
	}
	return t
}

func (p *DeploymentSettingsInput) Annotate(a infer.Annotator) {
	a.Describe(&p.Organization, "Organization name.")
	a.Describe(&p.Stack, "Stack name.")
	a.Describe(&p.Project, "Project name.")
	a.Describe(&p.SourceContext, "Settings related to the source of the deployment.")
	a.Describe(&p.ExecutorContext, "Settings related to the deployment executor.")
	a.Describe(&p.Github, "GitHub settings for the deployment.")
	a.Describe(&p.OperationContext, "Settings related to the Pulumi operation environment during the deployment.")
	a.Describe(&p.AgentPoolId, "The agent pool identifier to use for the deployment.")
}

type DeploymentSettingsSourceContext struct {
	Git *DeploymentSettingsGitSource `pulumi:"git,optional"`
}

func (p *DeploymentSettingsSourceContext) Annotate(a infer.Annotator) {
	a.Describe(&p, "Settings related to the source of the deployment.")
	a.Describe(&p.Git, "Git source settings for a deployment.")
}

func (p *DeploymentSettingsSourceContext) asAPI() *apitype.SourceContext {
	if p == nil {
		return nil
	}
	return &apitype.SourceContext{Git: p.Git.asAPI()}
}

type DeploymentSettingsGitSource struct {
	RepoUrl string                              `pulumi:"repoUrl,optional"`
	GitAuth *DeploymentSettingsGitSourceGitAuth `pulumi:"gitAuth,optional"`
	Branch  string                              `pulumi:"branch,optional"`
	Commit  string                              `pulumi:"commit,optional"`
	RepoDir string                              `pulumi:"repoDir,optional"`
}

func (p *DeploymentSettingsGitSource) Annotate(a infer.Annotator) {
	a.Describe(&p, "Git source settings for a deployment.")
	a.Describe(&p.RepoUrl, "The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.")
	a.Describe(&p.GitAuth, "Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.")
	a.Describe(&p.Branch, "The branch to deploy. One of either `branch` or `commit` must be specified.")
	a.Describe(&p.Commit, "The commit to deploy. One of either `branch` or `commit` must be specified.")
	a.Describe(&p.RepoDir, "The directory within the repository where the Pulumi.yaml is located.")
}

func (p *DeploymentSettingsGitSource) asAPI() *apitype.SourceContextGit {
	if p == nil {
		return nil
	}
	return &apitype.SourceContextGit{
		RepoURL: p.RepoUrl,
		Branch:  p.Branch,
		RepoDir: p.RepoDir,
		Commit:  p.Commit,
		GitAuth: p.GitAuth.asAPI(),
	}
}

type DeploymentSettingsGitSourceGitAuth struct {
	SshAuth   *DeploymentSettingsGitAuthSSHAuth   `pulumi:"sshAuth,optional"`
	BasicAuth *DeploymentSettingsGitAuthBasicAuth `pulumi:"basicAuth,optional"`
}

func (p *DeploymentSettingsGitSourceGitAuth) Annotate(a infer.Annotator) {
	// TODO[DOCS] This is the same description as DeploymentSettingsGitSource. I don't think it should be
	a.Describe(&p, "Git source settings for a deployment.")

	// TODO[DOCS][BUG] The description says there are 3 options here, but I only see 2.
	//
	// Also, these types should be a OneOf, but I don't think that the
	// pulumi-go-provider has a way to express that yes.
	a.Describe(&p.SshAuth, "SSH auth for git authentication. Only one of `personalAccessToken`, `sshAuth`, or `basicAuth` must be defined.")
	a.Describe(&p.BasicAuth, "Basic auth for git authentication. Only one of `personalAccessToken`, `sshAuth`, or `basicAuth` must be defined.")
}

func (p *DeploymentSettingsGitSourceGitAuth) asAPI() *apitype.GitAuthConfig {
	if p == nil {
		return nil
	}
	return &apitype.GitAuthConfig{
		PersonalAccessToken: nil,
		SSHAuth:             p.SshAuth.asAPI(),
		BasicAuth:           p.BasicAuth.asAPI(),
	}
}

type DeploymentSettingsGitAuthSSHAuth struct {
	SshPrivateKey string  `pulumi:"sshPrivateKey" provider:"secret"`
	Password      *string `pulumi:"password,optional" provider:"secret"`
}

func (p *DeploymentSettingsGitAuthSSHAuth) Annotate(a infer.Annotator) {
	a.Describe(&p, "Git source settings for a deployment.")
	a.Describe(&p.SshPrivateKey, "SSH private key.")
	a.Describe(&p.Password, "Optional password for SSH authentication.")
}

func (p *DeploymentSettingsGitAuthSSHAuth) asAPI() *apitype.SSHAuth {
	if p == nil {
		return nil
	}
	return &apitype.SSHAuth{
		SSHPrivateKey: apitype.SecretValue{Value: p.SshPrivateKey},
		Password:      &apitype.SecretValue{Value: derefOrZero(p.Password)},
	}
}

type DeploymentSettingsGitAuthBasicAuth struct {
	Username string `pulumi:"username" provider:"secret"`
	Password string `pulumi:"password" provider:"secret"`
}

func (p *DeploymentSettingsGitAuthBasicAuth) Annotate(a infer.Annotator) {
	a.Describe(&p, "Git source settings for a deployment.")
	a.Describe(&p.Username, "User name for git basic authentication.")
	a.Describe(&p.Password, "Password for git basic authentication.")
}

func (p *DeploymentSettingsGitAuthBasicAuth) asAPI() *apitype.BasicAuth {
	if p == nil {
		return nil
	}
	return &apitype.BasicAuth{
		UserName: apitype.SecretValue{Value: p.Username},
		Password: apitype.SecretValue{Value: p.Password},
	}
}

// TODO[USABILITY]: We should embed [DeploymentSettingsInput] in [DeploymentSettingsState]
// so users can access their inputs as outputs, and to improve diffs.
type DeploymentSettingsState struct {
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Stack        string `pulumi:"stack"`
}

func (p DeploymentSettingsState) asStack() pulumiapi.StackName {
	return pulumiapi.StackName{
		OrgName:     p.Organization,
		ProjectName: p.Project,
		StackName:   p.Stack,
	}
}

func (p *DeploymentSettingsState) Annotate(a infer.Annotator) {
	a.Describe(&p.Organization, "Organization name.")
	a.Describe(&p.Stack, "Stack name.")
	a.Describe(&p.Project, "Project name.")
}

type DeploymentSettingsExecutorContext struct {
	ExecutorImage string `pulumi:"executorImage"`
}

func (p *DeploymentSettingsExecutorContext) Annotate(a infer.Annotator) {
	a.Describe(&p, "The executor context defines information about the executor where the deployment is executed. If unspecified, the default 'pulumi/pulumi' image is used.")
	a.Describe(&p.ExecutorImage, "Allows overriding the default executor image with a custom image. E.g. 'pulumi/pulumi-nodejs:latest'")
}

func (p *DeploymentSettingsExecutorContext) asAPI() *apitype.ExecutorContext {
	if p == nil {
		return nil
	}
	return &apitype.ExecutorContext{
		WorkingDirectory: "", // TODO[CHECK]: Do we set the working directory?
		ExecutorImage: &apitype.DockerImage{
			Reference:   p.ExecutorImage,
			Credentials: nil, // TODO[CHECK]: Do we set Credentials?
		},
	}
}

type DeploymentSettingsGithub struct {
	Repository          *string  `pulumi:"repository,optional"`
	DeployCommits       *bool    `pulumi:"deployCommits,optional"`
	PreviewPullRequests *bool    `pulumi:"previewPullRequests,optional"`
	PullRequestTemplate *bool    `pulumi:"pullRequestTemplate,optional"`
	Paths               []string `pulumi:"paths,optional"`
}

func (p *DeploymentSettingsGithub) Annotate(a infer.Annotator) {
	a.Describe(&p, "GitHub settings for the deployment.")

	a.Describe(&p.Repository, "The GitHub repository in the format org/repo.")
	a.Describe(&p.DeployCommits, "Trigger a deployment running `pulumi up` on commit.")
	a.SetDefault(&p.DeployCommits, true)
	a.Describe(&p.PreviewPullRequests, "Trigger a deployment running `pulumi preview` when a PR is opened.")
	a.SetDefault(&p.PreviewPullRequests, true)
	a.Describe(&p.PullRequestTemplate, "Use this stack as a template for pull request review stacks.")
	a.SetDefault(&p.PullRequestTemplate, false)
	a.Describe(&p.Paths, "The paths within the repo that deployments should be filtered to.")
}

func (p *DeploymentSettingsGithub) asAPI() *pulumiapi.GitHubConfiguration {
	if p == nil {
		return nil
	}
	return &pulumiapi.GitHubConfiguration{
		Repository:          derefOrZero(p.Repository),
		DeployCommits:       derefOrZero(p.DeployCommits),
		PreviewPullRequests: derefOrZero(p.PreviewPullRequests),
		PullRequestTemplate: derefOrZero(p.PullRequestTemplate),
		Paths:               p.Paths,
	}
}

type DeploymentSettingsOperationContext struct {
	PreRunCommands       []string                 `pulumi:"preRunCommands,optional"`
	EnvironmentVariables map[string]string        `pulumi:"environmentVariables,optional"`
	Options              *OperationContextOptions `pulumi:"options,optional"`
	Oidc                 *OperationContextOIDC    `pulumi:"oidc,optional"`
}

func (p *DeploymentSettingsOperationContext) Annotate(a infer.Annotator) {
	a.Describe(&p, "Settings related to the Pulumi operation environment during the deployment.")
	a.Describe(&p.PreRunCommands, "Shell commands to run before the Pulumi operation executes.")
	a.Describe(&p.EnvironmentVariables, "Environment variables to set for the deployment.")
	a.Describe(&p.Options, "Options to override default behavior during the deployment.")
	a.Describe(&p.Oidc, "OIDC configuration to use during the deployment.")
}

func (p *DeploymentSettingsOperationContext) asAPI() *pulumiapi.OperationContext {
	if p == nil {
		return nil
	}

	return &pulumiapi.OperationContext{
		Options:              p.Options.asAPI(),
		PreRunCommands:       p.PreRunCommands,
		EnvironmentVariables: map[string]apitype.SecretValue{}, // TODO
		OIDC:                 p.Oidc.asAPI(),
	}
}

type OperationContextOptions struct {
	SkipInstallDependencies     *bool   `pulumi:"skipInstallDependencies,optional"`
	SkipIntermediateDeployments *bool   `pulumi:"skipIntermediateDeployments,optional"`
	Shell                       *string `pulumi:"shell,optional"`
}

func (p *OperationContextOptions) Annotate(a infer.Annotator) {
	a.Describe(&p.SkipInstallDependencies, "Skip the default dependency installation step "+
		"- use this to customize the dependency installation (e.g. if using yarn or poetry)")
	a.Describe(&p.SkipIntermediateDeployments, "Skip intermediate deployments (Consolidate "+
		"multiple deployments of the same type into one deployment)")
	a.Describe(&p.Shell, "The shell to use to run commands during the deployment. Defaults to 'bash'.")
}

func (p *OperationContextOptions) asAPI() *pulumiapi.OperationContextOptions {
	if p == nil {
		return nil
	}
	return &pulumiapi.OperationContextOptions{
		SkipInstallDependencies:     derefOrZero(p.SkipInstallDependencies),
		SkipIntermediateDeployments: derefOrZero(p.SkipIntermediateDeployments),
		Shell:                       derefOrZero(p.Shell),
	}
}

type OperationContextOIDC struct {
	Aws   *AWSOIDCConfiguration   `pulumi:"aws,optional"`
	Gcp   *GCPOIDCConfiguration   `pulumi:"gcp,optional"`
	Azure *AzureOIDCConfiguration `pulumi:"azure,optional"`
}

func (p *OperationContextOIDC) Annotate(a infer.Annotator) {
	a.Describe(&p.Aws, "AWS-specific OIDC configuration.")
	a.Describe(&p.Gcp, "GCP-specific OIDC configuration.")
	a.Describe(&p.Azure, "Azure-specific OIDC configuration.")
}

func (p *OperationContextOIDC) asAPI() *pulumiapi.OIDCConfiguration {
	if p == nil {
		return nil
	}
	return &pulumiapi.OIDCConfiguration{
		AWS:   p.Aws.asAPI(),
		GCP:   p.Gcp.asAPI(),
		Azure: p.Azure.asAPI(),
	}
}

type AWSOIDCConfiguration struct {
	Duration    string   `pulumi:"duration,optional"`
	PolicyARNs  []string `pulumi:"policyARNs,optional"`
	RoleARN     string   `pulumi:"roleARN"`
	SessionName string   `pulumi:"sessionName"`
}

func (p *AWSOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(&p.Duration, "Duration of the assume-role session in “XhYmZs” format")
	a.Describe(&p.PolicyARNs, "Optional set of IAM policy ARNs that further restrict the assume-role session")
	a.Describe(&p.RoleARN, "The ARN of the role to assume using the OIDC token.")
	a.Describe(&p.SessionName, "The name of the assume-role session.")
}

func (p *AWSOIDCConfiguration) asAPI() *pulumiapi.AWSOIDCConfiguration {
	if p == nil {
		return nil
	}
	return &pulumiapi.AWSOIDCConfiguration{
		Duration:    p.Duration,
		PolicyARNs:  p.PolicyARNs,
		RoleARN:     p.RoleARN,
		SessionName: p.SessionName,
	}
}

type GCPOIDCConfiguration struct {
	ProjectId      string `pulumi:"projectId"`
	Region         string `pulumi:"region,optional"`
	WorkloadPoolId string `pulumi:"workloadPoolId"`
	ProviderId     string `pulumi:"providerId"`
	ServiceAccount string `pulumi:"serviceAccount"`
	TokenLifetime  string `pulumi:"tokenLifetime,optional"`
}

func (p *GCPOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(&p.ProjectId, "The numerical ID of the GCP project.")
	a.Describe(&p.Region, "The region of the GCP project.")
	a.Describe(&p.WorkloadPoolId, "The ID of the workload pool to use.")
	a.Describe(&p.ProviderId, "The ID of the identity provider associated with the workload pool.")
	a.Describe(&p.ServiceAccount, "The email address of the service account to use.")
	a.Describe(&p.TokenLifetime, "The lifetime of the temporary credentials in “XhYmZs” format.")
}

func (p *GCPOIDCConfiguration) asAPI() *pulumiapi.GCPOIDCConfiguration {
	if p == nil {
		return nil
	}
	return &pulumiapi.GCPOIDCConfiguration{
		ProjectID:      p.ProjectId,
		Region:         p.Region,
		WorkloadPoolID: p.WorkloadPoolId,
		ProviderID:     p.ProjectId,
		ServiceAccount: p.ServiceAccount,
		TokenLifetime:  p.TokenLifetime,
	}
}

type AzureOIDCConfiguration struct {
	ClientId       string `pulumi:"clientId"`
	TenantId       string `pulumi:"tenantId"`
	SubscriptionId string `pulumi:"subscriptionId"`
}

func (p *AzureOIDCConfiguration) Annotate(a infer.Annotator) {
	a.Describe(&p.ClientId, "The client ID of the federated workload identity.")
	a.Describe(&p.TenantId, "The tenant ID of the federated workload identity.")
	a.Describe(&p.SubscriptionId, "The subscription ID of the federated workload identity.")
}

func (p *AzureOIDCConfiguration) asAPI() *pulumiapi.AzureOIDCConfiguration {
	if p == nil {
		return nil
	}
	return &pulumiapi.AzureOIDCConfiguration{
		ClientID:       p.ClientId,
		TenantID:       p.TenantId,
		SubscriptionID: p.SubscriptionId,
	}
}

type DeploymentSettingsStateOld struct {
	pulumiapi.DeploymentSettings
	Stack pulumiapi.StackName
}

const FixMe = "<value is secret and must be replaced>"

func (*DeploymentSettings) Read(
	ctx p.Context, id string, inputs DeploymentSettingsInput, _ DeploymentSettingsState,
) (string, DeploymentSettingsInput, DeploymentSettingsState, error) {

	var stack pulumiapi.StackName
	if err := stack.FromID(id); err != nil {
		return "", DeploymentSettingsInput{}, DeploymentSettingsState{}, err
	}
	settings, err := GetConfig(ctx).Client.GetDeploymentSettings(ctx, stack)
	if err != nil || settings == nil {
		return "", DeploymentSettingsInput{}, DeploymentSettingsState{}, err
	}

	return id, DeploymentSettingsInput{
			Organization:  stack.OrgName,
			Project:       stack.ProjectName,
			Stack:         stack.StackName,
			SourceContext: func() DeploymentSettingsSourceContext { panic("TODO") }(),
		}, DeploymentSettingsState{
			Organization: stack.OrgName,
			Project:      stack.ProjectName,
			Stack:        stack.StackName,
		}, nil
}

func (*DeploymentSettings) Delete(ctx p.Context, id string, props DeploymentSettingsState) error {
	return GetConfig(ctx).Client.DeleteDeploymentSettings(ctx, props.asStack())
}

func (*DeploymentSettings) Create(
	ctx p.Context, name string, inputs DeploymentSettingsInput, preview bool,
) (string, DeploymentSettingsState, error) {
	if preview {
		return "", DeploymentSettingsState{}, nil
	}

	stack := inputs.asStack()

	err := GetConfig(ctx).Client.CreateDeploymentSettings(ctx, stack, inputs.asAPI())
	if err != nil {
		return "", DeploymentSettingsState{}, err
	}

	return path.Join(stack.OrgName, stack.ProjectName, stack.StackName),
		DeploymentSettingsState{
			Organization: stack.OrgName,
			Project:      stack.ProjectName,
			Stack:        stack.StackName,
		}, nil
}
