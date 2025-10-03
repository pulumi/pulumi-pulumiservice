package pulumiapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// DeploymentSettingsClient provides methods for managing deployment settings.
type DeploymentSettingsClient interface {
	CreateDeploymentSettings(
		ctx context.Context, stack StackIdentifier, ds DeploymentSettings,
	) (*DeploymentSettings, error)
	UpdateDeploymentSettings(
		ctx context.Context, stack StackIdentifier, ds DeploymentSettings,
	) (*DeploymentSettings, error)
	GetDeploymentSettings(ctx context.Context, stack StackIdentifier) (*DeploymentSettings, error)
	DeleteDeploymentSettings(ctx context.Context, stack StackIdentifier) error
}

// DeploymentSettings represents deployment configuration for a stack.
type DeploymentSettings struct {
	OperationContext *OperationContext        `json:"operationContext,omitempty"`
	GitHub           *GitHubConfiguration     `json:"gitHub,omitempty"`
	SourceContext    *SourceContext           `json:"sourceContext,omitempty"`
	ExecutorContext  *apitype.ExecutorContext `json:"executorContext,omitempty"`
	AgentPoolId      string                   `json:"agentPoolID,omitempty"`
	Source           *string                  `json:"source,omitempty"`
	CacheOptions     *CacheOptions            `json:"cacheOptions,omitempty"`
}

// OperationContext represents operational settings for deployments.
type OperationContext struct {
	Options              *OperationContextOptions `json:"options,omitempty"`
	PreRunCommands       []string                 `json:"PreRunCommands,omitempty"`
	EnvironmentVariables map[string]SecretValue   `json:"environmentVariables,omitempty"`
	OIDC                 *OIDCConfiguration       `json:"oidc,omitempty"`
}

// OIDCConfiguration represents OIDC settings for cloud providers.
type OIDCConfiguration struct {
	AWS   *AWSOIDCConfiguration   `json:"aws,omitempty"`
	GCP   *GCPOIDCConfiguration   `json:"gcp,omitempty"`
	Azure *AzureOIDCConfiguration `json:"azure,omitempty"`
}

// AWSOIDCConfiguration represents AWS OIDC configuration.
type AWSOIDCConfiguration struct {
	Duration    string   `json:"duration,omitempty"`
	PolicyARNs  []string `json:"policyArns,omitempty"`
	RoleARN     string   `json:"roleArn"`
	SessionName string   `json:"sessionName"`
}

// GCPOIDCConfiguration represents GCP OIDC configuration.
type GCPOIDCConfiguration struct {
	ProjectID      string `json:"projectId,omitempty"`
	Region         string `json:"region"`
	WorkloadPoolID string `json:"workloadPoolId,omitempty"`
	ProviderID     string `json:"providerId,omitempty"`
	ServiceAccount string `json:"serviceAccount,omitempty"`
	TokenLifetime  string `json:"tokenLifetime"`
}

// AzureOIDCConfiguration represents Azure OIDC configuration.
type AzureOIDCConfiguration struct {
	ClientID       string `json:"clientId"`
	TenantID       string `json:"tenantId"`
	SubscriptionID string `json:"subscriptionId"`
}

// OperationContextOptions represents options for operation context.
type OperationContextOptions struct {
	SkipInstallDependencies     bool   `json:"skipInstallDependencies,omitempty"`
	SkipIntermediateDeployments bool   `json:"skipIntermediateDeployments,omitempty"`
	Shell                       string `json:"shell,omitempty"`
	DeleteAfterDestroy          bool   `json:"deleteAfterDestroy,omitempty"`
}

// GitHubConfiguration represents GitHub-specific deployment settings.
type GitHubConfiguration struct {
	Repository          string   `json:"repository,omitempty"`
	DeployCommits       bool     `json:"deployCommits,omitempty"`
	PreviewPullRequests bool     `json:"previewPullRequests,omitempty"`
	PullRequestTemplate bool     `json:"pullRequestTemplate,omitempty"`
	Paths               []string `json:"paths,omitempty"`
}

// SourceContext represents source code configuration.
type SourceContext struct {
	Git *SourceContextGit `json:"git,omitempty"`
}

// SourceContextGit represents Git source configuration.
type SourceContextGit struct {
	RepoURL string         `json:"repoURL"`
	Branch  string         `json:"branch"`
	RepoDir string         `json:"repoDir,omitempty"`
	Commit  string         `json:"commit,omitempty"`
	GitAuth *GitAuthConfig `json:"gitAuth,omitempty"`
}

// GitAuthConfig represents Git authentication configuration.
type GitAuthConfig struct {
	PersonalAccessToken *SecretValue `json:"accessToken,omitempty"`
	SSHAuth             *SSHAuth     `json:"sshAuth,omitempty"`
	BasicAuth           *BasicAuth   `json:"basicAuth,omitempty"`
}

// SSHAuth represents SSH authentication credentials.
type SSHAuth struct {
	SSHPrivateKey SecretValue  `json:"sshPrivateKey"`
	Password      *SecretValue `json:"password,omitempty"`
}

// BasicAuth represents basic authentication credentials.
type BasicAuth struct {
	UserName SecretValue `json:"userName"`
	Password SecretValue `json:"password"`
}

// SecretValue represents an encrypted secret value.
type SecretValue struct {
	Value  string // Plaintext if Secret is false; ciphertext otherwise.
	Secret bool
}

// CacheOptions represents caching configuration for deployments.
type CacheOptions struct {
	Enable bool `json:"enable"`
}

type secretCiphertextValue struct {
	Ciphertext string `json:"ciphertext"`
}

type secretWorkflowValue struct {
	Secret string `json:"secret" yaml:"secret"`
}

func (v SecretValue) MarshalJSON() ([]byte, error) {
	if v.Secret {
		return json.Marshal(secretWorkflowValue{Secret: v.Value})
	}
	return json.Marshal(v.Value)
}

func (v *SecretValue) UnmarshalJSON(bytes []byte) error {
	var secret secretCiphertextValue
	if err := json.Unmarshal(bytes, &secret); err == nil {
		v.Value, v.Secret = secret.Ciphertext, true
		return nil
	}

	var plaintext string
	if err := json.Unmarshal(bytes, &plaintext); err != nil {
		return err
	}
	v.Value, v.Secret = plaintext, false
	return nil
}

func (c *Client) CreateDeploymentSettings(
	ctx context.Context, stack StackIdentifier, ds DeploymentSettings,
) (*DeploymentSettings, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "settings")
	var resultDS = &DeploymentSettings{}
	_, err := c.do(ctx, http.MethodPut, apiPath, ds, resultDS)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment settings for stack (%s): %w", stack.String(), err)
	}
	return resultDS, nil
}

func (c *Client) UpdateDeploymentSettings(
	ctx context.Context, stack StackIdentifier, ds DeploymentSettings,
) (*DeploymentSettings, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "settings")
	var resultDS = &DeploymentSettings{}
	_, err := c.do(ctx, http.MethodPut, apiPath, ds, resultDS)
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment settings for stack (%s): %w", stack.String(), err)
	}
	return resultDS, nil
}

func (c *Client) GetDeploymentSettings(ctx context.Context, stack StackIdentifier) (*DeploymentSettings, error) {
	apiPath := path.Join(
		"stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "settings",
	)
	var ds DeploymentSettings
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &ds)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			// Important: we return nil here to hint it was not found
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get deployment settings for stack (%s): %w", stack.String(), err)
	}
	return &ds, nil
}

func (c *Client) DeleteDeploymentSettings(ctx context.Context, stack StackIdentifier) error {
	apiPath := path.Join(
		"stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "settings",
	)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete deployment settings for stack (%s): %w", stack.String(), err)
	}
	return nil
}
