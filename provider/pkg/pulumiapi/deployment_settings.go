package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/apitype"
)

type DeploymentSettingsClient interface {
	CreateDeploymentSettings(
		ctx context.Context,
		stack StackIdentifier,
		ds DeploymentSettings,
	) (*DeploymentSettings, error)
	UpdateDeploymentSettings(
		ctx context.Context,
		stack StackIdentifier,
		ds DeploymentSettings,
	) (*DeploymentSettings, error)
	GetDeploymentSettings(ctx context.Context, stack StackIdentifier) (*DeploymentSettings, error)
	DeleteDeploymentSettings(ctx context.Context, stack StackIdentifier) error
}

type DeploymentSettings = apitype.DeploymentSettings

type OperationContext = apitype.OperationContext

type OperationContextOIDCConfiguration = apitype.OperationContextOIDCConfiguration

type OperationContextAWSOIDCConfiguration = apitype.OperationContextAWSOIDCConfiguration

type OperationContextGCPOIDCConfiguration = apitype.OperationContextGCPOIDCConfiguration

type OperationContextAzureOIDCConfiguration = apitype.OperationContextAzureOIDCConfiguration

type OperationContextOptions = apitype.OperationContextOptions

type DeploymentSettingsGitHub = apitype.DeploymentSettingsGitHub

type DeploymentSettingsVCS = apitype.DeploymentSettingsVCS

type DeploymentSettingsVCSAzureDevOps = apitype.DeploymentSettingsVCSAzureDevOps
type DeploymentSettingsVCSAzureDevOpsBuilder = apitype.DeploymentSettingsVCSAzureDevOpsBuilder

type DeploymentSettingsVCSBitbucket = apitype.DeploymentSettingsVCSBitbucket
type DeploymentSettingsVCSBitbucketBuilder = apitype.DeploymentSettingsVCSBitbucketBuilder

type DeploymentSettingsVCSCustom = apitype.DeploymentSettingsVCSCustom
type DeploymentSettingsVCSCustomBuilder = apitype.DeploymentSettingsVCSCustomBuilder

type DeploymentSettingsVCSGitHub = apitype.DeploymentSettingsVCSGitHub
type DeploymentSettingsVCSGitHubBuilder = apitype.DeploymentSettingsVCSGitHubBuilder

type DeploymentSettingsVCSGitLab = apitype.DeploymentSettingsVCSGitLab
type DeploymentSettingsVCSGitLabBuilder = apitype.DeploymentSettingsVCSGitLabBuilder

type DeploymentSettingsVCSBuilder = apitype.DeploymentSettingsVCSBuilder

type SourceContext = apitype.SourceContext

type ExecutorContext = apitype.ExecutorContext

type DockerImage = apitype.DockerImage

type SourceContextGit = apitype.SourceContextGit

type GitAuthConfig = apitype.GitAuthConfig

type SSHAuth = apitype.SSHAuth

type BasicAuth = apitype.BasicAuth

type SecretValue = apitype.SecretValue

type CacheOptions = apitype.CacheOptions

func (c *Client) CreateDeploymentSettings(
	ctx context.Context,
	stack StackIdentifier,
	ds DeploymentSettings,
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
	ctx context.Context,
	stack StackIdentifier,
	ds DeploymentSettings,
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
