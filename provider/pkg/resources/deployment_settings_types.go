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
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/mapper"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// The structs below mirror pulumiservice:index:DeploymentSettings and its types
// in manual-schema.json. Each type is named after its schema token.
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

type DeploymentSettingsExecutorContext struct {
	ExecutorImage *string                                     `pulumi:"executorImage"`
	Credentials   *DeploymentSettingsExecutorImageCredentials `pulumi:"credentials,optional"`
}

type DeploymentSettingsExecutorImageCredentials struct {
	Username string  `pulumi:"username"`
	Password *string `pulumi:"password" provider:"secret"`
}

type DeploymentSettingsSourceContext struct {
	Git *DeploymentSettingsGitSource `pulumi:"git,optional"`
}

type DeploymentSettingsGitSource struct {
	RepoURL string                              `pulumi:"repoUrl,optional"`
	GitAuth *DeploymentSettingsGitSourceGitAuth `pulumi:"gitAuth,optional"`
	Branch  string                              `pulumi:"branch,optional"`
	Commit  string                              `pulumi:"commit,optional"`
	RepoDir string                              `pulumi:"repoDir,optional"`
}

type DeploymentSettingsGitSourceGitAuth struct {
	SSHAuth   *DeploymentSettingsGitAuthSSHAuth   `pulumi:"sshAuth,optional"`
	BasicAuth *DeploymentSettingsGitAuthBasicAuth `pulumi:"basicAuth,optional"`
}

type DeploymentSettingsGitAuthSSHAuth struct {
	SSHPrivateKey *string `pulumi:"sshPrivateKey"     provider:"secret"`
	Password      *string `pulumi:"password,optional" provider:"secret"`
}

type DeploymentSettingsGitAuthBasicAuth struct {
	Username *string `pulumi:"username" provider:"secret"`
	Password *string `pulumi:"password" provider:"secret"`
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

type DeploymentSettingsOperationContext struct {
	PreRunCommands       []string                 `pulumi:"preRunCommands,optional"`
	EnvironmentVariables map[string]string        `pulumi:"environmentVariables,optional"`
	Options              *OperationContextOptions `pulumi:"options,optional"`
	OIDC                 *OperationContextOIDC    `pulumi:"oidc,optional"`
}

type OperationContextOptions struct {
	SkipInstallDependencies     bool   `pulumi:"skipInstallDependencies,optional"`
	SkipIntermediateDeployments bool   `pulumi:"skipIntermediateDeployments,optional"`
	Shell                       string `pulumi:"shell,optional"`
	DeleteAfterDestroy          bool   `pulumi:"deleteAfterDestroy,optional"`
}

type OperationContextOIDC struct {
	AWS   *AWSOIDCConfiguration   `pulumi:"aws,optional"`
	GCP   *GCPOIDCConfiguration   `pulumi:"gcp,optional"`
	Azure *AzureOIDCConfiguration `pulumi:"azure,optional"`
}

type AWSOIDCConfiguration struct {
	Duration    string   `pulumi:"duration,optional"`
	PolicyARNs  []string `pulumi:"policyARNs,optional"`
	RoleARN     string   `pulumi:"roleARN"`
	SessionName string   `pulumi:"sessionName"`
}

type GCPOIDCConfiguration struct {
	ProjectID      string `pulumi:"projectId"`
	Region         string `pulumi:"region,optional"`
	WorkloadPoolID string `pulumi:"workloadPoolId"`
	ProviderID     string `pulumi:"providerId"`
	ServiceAccount string `pulumi:"serviceAccount"`
	TokenLifetime  string `pulumi:"tokenLifetime,optional"`
}

type AzureOIDCConfiguration struct {
	ClientID       string `pulumi:"clientId"`
	TenantID       string `pulumi:"tenantId"`
	SubscriptionID string `pulumi:"subscriptionId"`
}

type DeploymentSettingsCacheOptions struct {
	Enable bool `pulumi:"enable,optional"`
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
