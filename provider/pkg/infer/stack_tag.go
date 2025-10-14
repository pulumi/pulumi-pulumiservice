// Copyright 2016-2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package infer

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// StackTag is a Pulumi resource that represents a tag on a stack.
// Tags are key-value pairs that can be used to organize and categorize stacks.
type StackTag struct{}

// StackTagArgs defines the inputs for creating a StackTag.
type StackTagArgs struct {
	// Organization is the name of the Pulumi organization.
	Organization string `pulumi:"organization"`
	// Project is the name of the Pulumi project.
	Project string `pulumi:"project"`
	// Stack is the name of the Pulumi stack.
	Stack string `pulumi:"stack"`
	// Name is the tag key/name.
	Name string `pulumi:"name"`
	// Value is the tag value.
	Value string `pulumi:"value"`
}

// StackTagState represents the state of a StackTag resource.
type StackTagState struct {
	StackTagArgs
}

// Create creates a new stack tag using the v1.1.2 infer API.
func (*StackTag) Create(
	ctx context.Context,
	req infer.CreateRequest[StackTagArgs],
) (infer.CreateResponse[StackTagState], error) {
	input := req.Inputs

	if req.DryRun {
		// During preview, return a placeholder ID
		id := path.Join(input.Organization, input.Project, input.Stack, input.Name)
		return infer.CreateResponse[StackTagState]{
			ID:     id,
			Output: StackTagState{StackTagArgs: input},
		}, nil
	}

	// Get the API client from context
	// TODO: Implement proper client context handling once we set up configuration
	client := getClientFromContext(ctx)
	if client == nil {
		return infer.CreateResponse[StackTagState]{}, fmt.Errorf("API client not configured")
	}

	stackIdentifier := pulumiapi.StackIdentifier{
		OrgName:     input.Organization,
		ProjectName: input.Project,
		StackName:   input.Stack,
	}

	stackTag := pulumiapi.StackTag{
		Name:  input.Name,
		Value: input.Value,
	}

	err := client.CreateTag(ctx, stackIdentifier, stackTag)
	if err != nil {
		return infer.CreateResponse[StackTagState]{}, fmt.Errorf("failed to create stack tag: %w", err)
	}

	id := path.Join(input.Organization, input.Project, input.Stack, input.Name)
	state := StackTagState{StackTagArgs: input}

	return infer.CreateResponse[StackTagState]{
		ID:     id,
		Output: state,
	}, nil
}

// Read reads the current state of a stack tag using the v1.1.2 infer API.
func (*StackTag) Read(
	ctx context.Context,
	id string,
	inputs StackTagArgs,
	state StackTagState,
) (
	canonicalID string,
	normalizedInputs StackTagArgs,
	normalizedState StackTagState,
	err error,
) {
	// Parse the ID
	organization, project, stack, tagName, err := splitStackTagID(id)
	if err != nil {
		return "", inputs, state, err
	}

	client := getClientFromContext(ctx)
	if client == nil {
		return "", inputs, state, fmt.Errorf("API client not configured")
	}

	stackIdentifier := pulumiapi.StackIdentifier{
		OrgName:     organization,
		ProjectName: project,
		StackName:   stack,
	}

	tag, err := client.GetStackTag(ctx, stackIdentifier, tagName)
	if err != nil {
		return "", inputs, state, fmt.Errorf("failed to read stack tag: %w", err)
	}

	// If tag doesn't exist, return empty ID to signal deletion
	if tag == nil {
		return "", inputs, state, nil
	}

	// Update state with current values from API
	normalizedState = StackTagState{
		StackTagArgs: StackTagArgs{
			Organization: organization,
			Project:      project,
			Stack:        stack,
			Name:         tag.Name,
			Value:        tag.Value,
		},
	}

	return id, normalizedState.StackTagArgs, normalizedState, nil
}

// Update is not implemented because all stack tag updates are destructive (delete + recreate).
// By not implementing Update, the infer framework will automatically trigger a replace operation
// for any changes, which matches the behavior of the manual implementation.
//
// The manual implementation explicitly returns an error:
// "unexpected call to update, expected create to be called instead"
//
// The infer framework handles this automatically by detecting no Update method and triggering
// a replace (Delete + Create) instead.

// Delete deletes a stack tag.
func (*StackTag) Delete(ctx context.Context, id string, state StackTagState) error {
	// Parse the ID
	organization, project, stack, tagName, err := splitStackTagID(id)
	if err != nil {
		return err
	}

	client := getClientFromContext(ctx)
	if client == nil {
		return fmt.Errorf("API client not configured")
	}

	stackIdentifier := pulumiapi.StackIdentifier{
		OrgName:     organization,
		ProjectName: project,
		StackName:   stack,
	}

	err = client.DeleteStackTag(ctx, stackIdentifier, tagName)
	if err != nil {
		return fmt.Errorf("failed to delete stack tag: %w", err)
	}

	return nil
}

// Annotate provides descriptions for the StackTag resource and its fields.
func (*StackTag) Annotate(a infer.Annotator) {
	a.Describe(new(StackTag), "A Stack Tag associates metadata with a Pulumi stack. "+
		"Tags are key-value pairs that can be used to organize and categorize stacks.")
}

// splitStackTagID parses a stack tag ID into its components.
// ID format: organization/project/stack/tagName
func splitStackTagID(id string) (organization, project, stack, tagName string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid stack tag ID %q: must be in format organization/project/stack/tagName", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

// getClientFromContext retrieves the Pulumi Service API client from the context.
func getClientFromContext(ctx context.Context) *pulumiapi.Client {
	return GetClient(ctx)
}
