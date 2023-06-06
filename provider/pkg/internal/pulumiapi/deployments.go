package pulumiapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"
)

type CreateDeploymentRequest struct {
	DeploymentSettings
	Op              string `json:"operation,omitempty"`
	InheritSettings bool   `json:"inheritSettings,omitempty"`
}

type CreateDeploymentResponse struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type StepRun struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Started     *time.Time `json:"started,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
}

type DeploymentJob struct {
	Status      string     `json:"status"`
	Started     *time.Time `json:"started,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
	Steps       []StepRun  `json:"steps"`
}

type GetDeploymentResponse struct {
	ID            string          `json:"id"`
	Created       string          `json:"created"`
	Modified      string          `json:"modified"`
	Status        string          `json:"status"`
	Version       int             `json:"version"`
	Jobs          []DeploymentJob `json:"jobs"`
	LatestVersion int             `json:"latestVersion"`
}

type DeploymentLogLine struct {
	Header    string    `json:"header,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Line      string    `json:"line,omitempty"`
}

type DeploymentLogs struct {
	Lines     []DeploymentLogLine `json:"lines,omitempty"`
	NextToken string              `json:"nextToken,omitempty"`
}

type UpdateInfo struct {
	Version int `json:"version"`
}

func (c *Client) CreateDeployment(ctx context.Context, stack StackName, req CreateDeploymentRequest) (*CreateDeploymentResponse, error) {
	apiPath := path.Join("preview", stack.OrgName, stack.ProjectName, stack.StackName, "deployments")
	var resp CreateDeploymentResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &resp)
	if err != nil {
		var errResp *errorResponse
		if errors.As(err, &errResp) && errResp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("stack (%s) not found", stack.String())
		}
		return nil, fmt.Errorf("failed to create deployment for stack (%s): %w", stack.String(), err)
	}
	return &resp, nil
}

func (c *Client) GetDeployment(ctx context.Context, stack StackName, deploymentID string) (*GetDeploymentResponse, error) {
	apiPath := path.Join("preview", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", deploymentID)
	var resp GetDeploymentResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &resp)
	if err != nil {
		var errResp *errorResponse
		if errors.As(err, &errResp) && errResp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("deployment (%s) not found", stack.String())
		}
		return nil, fmt.Errorf("failed to get deployment (%s): %w", deploymentID, err)
	}
	return &resp, nil
}

func (c *Client) GetDeploymentLogs(ctx context.Context, stack StackName, deploymentID, continuationToken string) (*DeploymentLogs, error) {
	apiPath := path.Join("preview", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", deploymentID, "logs")
	if continuationToken != "" {
		apiPath += "?continuationToken=" + continuationToken
	}
	var resp DeploymentLogs
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &resp)
	if err != nil {
		var errResp *errorResponse
		if errors.As(err, &errResp) && errResp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("deployment (%s) not found", deploymentID)
		}
		return nil, fmt.Errorf("failed to get deployment logs (%s): %w", deploymentID, err)
	}
	return &resp, nil
}

func (c *Client) GetDeploymentUpdates(ctx context.Context, stack StackName, deploymentID string) ([]UpdateInfo, error) {
	apiPath := path.Join("preview", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", deploymentID, "updates")
	var resp []UpdateInfo
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &resp)
	if err != nil {
		var errResp *errorResponse
		if errors.As(err, &errResp) && errResp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("deployment (%s) not found", stack.String())
		}
		return nil, fmt.Errorf("failed to get deployment (%s): %w", deploymentID, err)
	}
	return resp, nil
}
