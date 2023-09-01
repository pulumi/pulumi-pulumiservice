package pulumiapi

import (
	"context"
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"net/http"
	"path"
)

type CreateDeploymentRequest struct {
	InheritSettings bool                    `json:"inheritSettings"`
	Operation       apitype.PulumiOperation `json:"operation"`
}

type CreateDeploymentResponse struct {
	ID         string `json:"id"`
	Version    int    `json:"version"`
	ConsoleURL string `json:"consoleUrl"`
}

type GetDeploymentResponse struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Status  string `json:"status"`
}

func (c *Client) CreateDeployment(ctx context.Context, stack StackName, args CreateDeploymentRequest) (*CreateDeploymentResponse, error) {
	apiPath := path.Join("preview", "stacks", stack.String(), "deployments")
	var resp CreateDeploymentResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, args, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment for stack (%s): %w", stack.String(), err)
	}
	return &resp, nil
}

func (c *Client) GetDeployment(ctx context.Context, stack StackName, deploymentID string) (*GetDeploymentResponse, error) {
	apiPath := path.Join("preview", "stacks", stack.String(), "deployments", deploymentID)
	var resp GetDeploymentResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment for stack (%s): %w", stack.String(), err)
	}
	return &resp, nil
}
