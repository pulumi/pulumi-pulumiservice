package pulumiapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
)

type CreateStackRequest struct {
	StackName string `json:"stackName"`
}

type GetStackResponse struct {
	OrgName     string `json:"orgName"`
	ProjectName string `json:"projectName"`
	StackName   string `json:"stackName"`
}

func (c *Client) CreateStack(ctx context.Context, stack StackName) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName)
	_, err := c.do(ctx, http.MethodPost, apiPath, &CreateStackRequest{StackName: stack.StackName}, nil)
	if err != nil {
		return fmt.Errorf("failed to create stack (%v): %w", stack, err)
	}
	return nil
}

func (c *Client) GetStack(ctx context.Context, stack StackName) (*GetStackResponse, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName)
	var resp GetStackResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &resp)
	if err != nil {
		var errResp *errorResponse
		if errors.As(err, &errResp) && errResp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get stack (%v): %w", stack, err)
	}
	return &resp, nil
}

func (c *Client) DeleteStack(ctx context.Context, stack StackName) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete stack (%v): %w", stack, err)
	}
	return nil
}
