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
	GetStackID(ctx context.Context, stack StackIdentifier) (string, error)
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
	id, err := c.GetStackID(ctx, stackName)
	return id != "", err
}

// GetStackID returns the stack's unique ID (the program ID that RBAC entity
// rules reference), or "" if the stack does not exist.
func (c *Client) GetStackID(ctx context.Context, stackName StackIdentifier) (string, error) {
	if stackName.OrgName == "" || stackName.ProjectName == "" || stackName.StackName == "" {
		return "", fmt.Errorf("invalid stack identifier: %v", stackName)
	}
	apiPath := path.Join("stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName)
	var s struct {
		ID string `json:"id"`
	}
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &s)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			return "", nil
		}

		return "", fmt.Errorf("failed to get stack: %w", err)
	}
	if s.ID == "" {
		return "", fmt.Errorf("stack %s returned no id", stackName)
	}
	return s.ID, nil
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
