package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type CreateStackRequest struct {
	StackName string `json:"stackName"`
}

// AppStackLinks describes links to the stack in the Pulumi Console
type AppStackLinks struct {
	Self string `json:"self"`
}

// AppStackSummary describes the state of a stack, without including its specific resources
type AppStackSummary struct {
	ID            string         `json:"id"`
	OrgName       string         `json:"orgName"`
	ProjectName   string         `json:"projectName"`
	StackName     string         `json:"stackName"`
	LastUpdate    *int64         `json:"lastUpdate,omitempty"`
	ResourceCount *int           `json:"resourceCount,omitempty"`
	Links         *AppStackLinks `json:"links,omitempty"`
}

// ListStacksResponse returns a set of stack summaries
type ListStacksResponse struct {
	Stacks            []AppStackSummary `json:"stacks"`
	ContinuationToken *string           `json:"continuationToken,omitempty"`
}

// StackTeam is a team that has access to a particular stack
type StackTeam struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Permission  int    `json:"permission"`
	IsMember    bool   `json:"isMember"`
}

// ListTeamsByStackResponse lists all teams that have access to a stack
type ListTeamsByStackResponse struct {
	ProjectName string      `json:"projectName"`
	Teams       []StackTeam `json:"teams"`
}

// UserPermission describes a user's permission level for a stack
type UserPermission struct {
	User       UserInfo `json:"user"`
	Permission int      `json:"permission"`
}

// ListStackCollaboratorsResponse is the response when querying a stack's collaborators
type ListStackCollaboratorsResponse struct {
	Users                 []UserPermission `json:"users"`
	UserStackPermission   int              `json:"userStackPermission"`
	StackCreatorUserName  *string          `json:"stackCreatorUserName,omitempty"`
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
		_, err = c.doWithQuery(ctx, http.MethodDelete, apiPath, url.Values{"forceDestroy": []string{"true"}}, nil, nil)
	} else {
		_, err = c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	}
	if err != nil {
		return fmt.Errorf("failed to delete stack: %w", err)
	}

	return nil
}

// ListUserStacks lists all stacks accessible by the authenticated user
func (c *Client) ListUserStacks(ctx context.Context, maxResults *int) (*ListStacksResponse, error) {
	apiPath := "user/stacks"
	var response ListStacksResponse

	if maxResults != nil && *maxResults > 0 {
		query := url.Values{"maxResults": []string{fmt.Sprintf("%d", *maxResults)}}
		_, err := c.doWithQuery(ctx, http.MethodGet, apiPath, query, nil, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to list user stacks: %w", err)
		}
	} else {
		_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to list user stacks: %w", err)
		}
	}

	return &response, nil
}

// ListStackTeamPermissions lists all teams that have access to a stack
func (c *Client) ListStackTeamPermissions(ctx context.Context, stack StackIdentifier) (*ListTeamsByStackResponse, error) {
	if stack.OrgName == "" || stack.ProjectName == "" || stack.StackName == "" {
		return nil, fmt.Errorf("invalid stack identifier: %v", stack)
	}

	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "teams")
	var response ListTeamsByStackResponse

	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list stack team permissions: %w", err)
	}

	return &response, nil
}

// ListStackCollaborators lists all user collaborators for a stack
func (c *Client) ListStackCollaborators(ctx context.Context, stack StackIdentifier) (*ListStackCollaboratorsResponse, error) {
	if stack.OrgName == "" || stack.ProjectName == "" || stack.StackName == "" {
		return nil, fmt.Errorf("invalid stack identifier: %v", stack)
	}

	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "collaborators")
	var response ListStackCollaboratorsResponse

	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list stack collaborators: %w", err)
	}

	return &response, nil
}
