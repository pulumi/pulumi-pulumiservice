package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
)

type StackIdentifier struct {
	OrgName     string `json:"orgName"`
	ProjectName string `json:"projectName"`
	StackName   string `json:"stackName"`
}

func (s StackIdentifier) String() string {
	return fmt.Sprintf("%s/%s/%s", s.OrgName, s.ProjectName, s.StackName)
}

func NewStackIdentifier(id string) (StackIdentifier, error) {
	splitID := strings.Split(id, "/")
	if len(splitID) != 3 {
		return StackIdentifier{}, fmt.Errorf("invalid stack id: %s", id)
	}
	return StackIdentifier{
		OrgName:     splitID[0],
		ProjectName: splitID[1],
		StackName:   splitID[2],
	}, nil
}

type StackTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// in order to retrieve stack tags, we have to get the entire stack. we only need to unmarshal the tags property
type stack struct {
	Tags map[string]string `json:"tags"`
}

func (c *Client) CreateTag(ctx context.Context, stack StackIdentifier, tag StackTag) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "tags")
	_, err := c.do(ctx, http.MethodPost, apiPath, tag, nil)
	if err != nil {
		return fmt.Errorf("failed to create tag (%s=%s): %w", tag.Name, tag.Value, err)
	}
	return nil
}

func (c *Client) GetStackTag(ctx context.Context, stackName StackIdentifier, tagName string) (*StackTag, error) {
	apiPath := path.Join("stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName)
	var s stack
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &s)
	if err != nil {
		return nil, fmt.Errorf("failed to get stack tag: %w", err)
	}
	tagValue, ok := s.Tags[tagName]
	if !ok {
		return nil, nil
	}
	return &StackTag{
		Name:  tagName,
		Value: tagValue,
	}, nil
}

func (c *Client) DeleteStackTag(ctx context.Context, stackName StackIdentifier, tagName string) error {
	apiPath := path.Join(
		"stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName, "tags", tagName,
	)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	return nil
}
