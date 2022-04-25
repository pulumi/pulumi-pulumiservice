package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
)

type StackName struct {
	OrgName     string `json:"orgName"`
	ProjectName string `json:"projectName"`
	StackName   string `json:"stackName"`
}

type StackTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

<<<<<<< HEAD
<<<<<<< HEAD
func (c *Client) CreateTag(ctx context.Context, stack StackName, tag StackTag) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "tags")
	_, err := c.do(ctx, http.MethodPost, apiPath, tag, nil)
	if err != nil {
		return fmt.Errorf("failed to create tag (%s=%s): %w", tag.Name, tag.Value, err)
=======
func (c *Client) SetTags(ctx context.Context, stack StackName, tags map[string]string) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "tags")
	for name, value := range tags {
		tag := StackTag{
			Name:  name,
			Value: value,
		}
		_, err := c.do(ctx, http.MethodPost, apiPath, tag, nil)
		if err != nil {
			return fmt.Errorf("failed to create tag (%s=%s): %w", name, value, err)
		}

>>>>>>> d06708e (Add tests for api client library)
=======
func (c *Client) CreateTag(ctx context.Context, stack StackName, tag StackTag) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "tags")
	_, err := c.do(ctx, http.MethodPost, apiPath, tag, nil)
	if err != nil {
		return fmt.Errorf("failed to create tag (%s=%s): %w", tag.Name, tag.Value, err)
>>>>>>> 932b63e (rename SetStackTags to CreateStackTag to better match model)
	}
	return nil
}

func (c *Client) DeleteStackTag(ctx context.Context, stackName StackName, tagName string) error {
	apiPath := path.Join(
		"stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName, "tags", tagName,
	)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	return nil
}
