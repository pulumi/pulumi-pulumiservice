package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type StackClient interface {
	CreateStack(ctx context.Context, stack StackIdentifier) error
	StackExists(ctx context.Context, stack StackIdentifier) (bool, error)
	DeleteStack(ctx context.Context, stack StackIdentifier, forceDestroy bool) error
}

type CreateStackRequest struct {
	StackName string `json:"stackName"`
}

func (c *Client) CreateStack(ctx context.Context, stack StackIdentifier) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName)
	_, err := c.do(ctx, http.MethodPost, apiPath, CreateStackRequest{
		StackName: stack.StackName,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create stack '%s': %w", stack, err)
	}
	return nil
}

func (c *Client) StackExists(ctx context.Context, stackName StackIdentifier) (bool, error) {
	if stackName.OrgName == "" || stackName.ProjectName == "" || stackName.StackName == "" {
		return false, fmt.Errorf("invalid stack identifier: %v", stackName)
	}
	apiPath := path.Join("stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName)
	var s stack
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &s)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			return false, nil
		}

		return false, fmt.Errorf("failed to get stack: %w", err)
	}
	return true, nil
}

func (c *Client) DeleteStack(ctx context.Context, stackName StackIdentifier, forceDestroy bool) error {
	apiPath := path.Join(
		"stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName,
	)

	var err error
	if forceDestroy {
		_, err = c.doWithQuery(ctx, http.MethodDelete, apiPath, url.Values{"forceDestroy": []string{trueValue}}, nil, nil)
	} else {
		_, err = c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	}
	if err != nil {
		return fmt.Errorf("failed to delete stack: %w", err)
	}

	return nil
}
