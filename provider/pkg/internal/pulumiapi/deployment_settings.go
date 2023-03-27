package pulumiapi

//func (c *Client) CreateDeploymentSettings(ctx context.Context, stack StackName, tag StackTag) error {
//	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "tags")
//	_, err := c.do(ctx, http.MethodPost, apiPath, tag, nil)
//	if err != nil {
//		return fmt.Errorf("failed to create tag (%s=%s): %w", tag.Name, tag.Value, err)
//	}
//	return nil
//}
//
//func (c *Client) GetDeploymentSettings(ctx context.Context, stackName StackName, tagName string) (*StackTag, error) {
//	apiPath := path.Join(
//		"stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName, "deployment", "settings",
//	)
//	var ds deploymentSettings
//	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &s)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get stack tag: %w", err)
//	}
//	tagValue, ok := s.Tags[tagName]
//	if !ok {
//		return nil, nil
//	}
//	return &StackTag{
//		Name:  tagName,
//		Value: tagValue,
//	}, nil
//}
//
//func (c *Client) DeleteDeploymentSettings(ctx context.Context, stackName StackName) error {
//	apiPath := path.Join(
//		"stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName, "deployment", "settings",
//	)
//	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
//	if err != nil {
//		return fmt.Errorf("failed to make request: %w", err)
//	}
//	return nil
//}
