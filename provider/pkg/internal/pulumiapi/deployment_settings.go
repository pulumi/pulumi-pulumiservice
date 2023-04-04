package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

type DeploymentSettings struct {
	OperationContext *OperationContext        `json:"operationContext,omitempty"`
	GitHub           *GitHubConfiguration     `json:"gitHub,omitempty"`
	SourceContext    *apitype.SourceContext   `json:"sourceContext,omitempty"`
	ExecutorContext  *apitype.ExecutorContext `json:"executorContext,omitempty"`
}

type OperationContext struct {
	Options              *OperationContextOptions       `json:"options,omitempty"`
	PreRunCommands       []string                       `json:"PreRunCommands,omitempty"`
	EnvironmentVariables map[string]apitype.SecretValue `json:"environmentVariables"`
	OIDC                 *OIDCConfiguration             `json:"oidc"`
}

type OIDCConfiguration struct {
	AWS   *AWSOIDCConfiguration   `json:"aws,omitempty"`
	GCP   *GCPOIDCConfiguration   `json:"gcp,omitempty"`
	Azure *AzureOIDCConfiguration `json:"azure,omitempty"`
}

type AWSOIDCConfiguration struct {
	Duration    string   `json:"duration,omitempty"`
	PolicyARNs  []string `json:"policyArns,omitempty"`
	RoleARN     string   `json:"roleArn"`
	SessionName string   `json:"sessionName"`
}

type GCPOIDCConfiguration struct {
	ProjectID      string `json:"projectId,omitempty"`
	Region         string `json:"region"`
	WorkloadPoolID string `json:"workloadPoolId,omitempty"`
	ProviderID     string `json:"providerId,omitempty"`
	ServiceAccount string `json:"serviceAccount,omitempty"`
	TokenLifetime  string `json:"tokenLifetime"`
}

type AzureOIDCConfiguration struct {
	ClientID       string `json:"clientId"`
	TenantID       string `json:"tenantId"`
	SubscriptionID string `json:"subscriptionId"`
}

type OperationContextOptions struct {
	SkipInstallDependencies bool   `json:"skipInstallDependencies,omitempty"`
	Shell                   string `json:"shell"`
}

type GitHubConfiguration struct {
	Repository          string   `json:"repository,omitempty"`
	DeployCommits       bool     `json:"deployCommits,omitempty"`
	PreviewPullRequests bool     `json:"previewPullRequests,omitempty"`
	Paths               []string `json:"paths,omitempty"`
}

func (c *Client) CreateDeploymentSettings(ctx context.Context, stack StackName, ds DeploymentSettings) error {
	apiPath := path.Join("preview", stack.OrgName, stack.ProjectName, stack.StackName, "deployment", "settings")
	_, err := c.do(ctx, http.MethodPost, apiPath, ds, nil)
	if err != nil {
		return fmt.Errorf("failed to create deployment settings for stack (%s/%s/%s): %w", stack.OrgName, stack.ProjectName, stack.StackName, err)
	}
	return nil
}

//
//func (c *Client) GetDeploymentSettings(ctx context.Context, stackName StackName, tagName string) (*StackTag, error) {
//	apiPath := path.Join(
//		"stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName, "deployment", "settings",
//	)
//	var ds deploymentSettings
//	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &s)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get stack tag: %w", err)
//	}
//	tagValue, ok := s.Tags[tagName]
//	if !ok {
//		return nil, nil
//	}
//	return &StackTag{
//		Name:  tagName,
//		Value: tagValue,
//	}, nil
//}

func (c *Client) DeleteDeploymentSettings(ctx context.Context, stackName StackName) error {
	apiPath := path.Join(
		"preview", stackName.OrgName, stackName.ProjectName, stackName.StackName, "deployment", "settings",
	)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	return nil
}
