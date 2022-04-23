package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
)

type Stack struct {
	OrgName     string            `json:"orgName"`
	ProjectName string            `json:"projectName"`
	StackName   string            `json:"stackName"`
	Tags        map[string]string `json:"tags"`
}

type StackTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) SetTags(ctx context.Context, stack Stack) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "tags")
	for name, value := range stack.Tags {
		tag := StackTag{
			Name:  name,
			Value: value,
		}
		_, err := c.do(ctx, http.MethodPost, apiPath, tag, nil)
		if err != nil {
			return fmt.Errorf("failed to create tag (%s=%s): %w", name, value, err)
		}

	}
	return nil
}

func (c *Client) DeleteTag(ctx context.Context, stack Stack, tagName string) error {
	apiPath := path.Join(
		"stacks", stack.OrgName, stack.ProjectName, stack.StackName, "tags", tagName,
	)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	return nil
}
