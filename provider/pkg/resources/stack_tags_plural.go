// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type StackTags struct{}

var (
	_ infer.CustomCreate[StackTagsInput, StackTagsState] = &StackTags{}
	_ infer.CustomUpdate[StackTagsInput, StackTagsState] = &StackTags{}
	_ infer.CustomDelete[StackTagsState]                 = &StackTags{}
	_ infer.CustomRead[StackTagsInput, StackTagsState]   = &StackTags{}
)

func (*StackTags) Annotate(a infer.Annotator) {
	a.Describe(&StackTags{},
		"Manages a set of stack tags as a single resource via a `tags` map, instead of one `StackTag` per key — "+
			"useful for YAML programs.\n\n"+
			"Only tags declared in `tags` are managed; tags added out-of-band (CLI, pulumibot, a singular `StackTag` "+
			"resource) are left alone. Tag values are immutable in Pulumi Cloud, so a value change is implemented as "+
			"delete-and-recreate.\n\n"+
			"Importing with ID `{organization}/{project}/{stack}/tags` adopts every tag currently on the stack; "+
			"declare `tags` explicitly after import so subsequent updates match your intent. See the "+
			"[registry docs](https://www.pulumi.com/registry/packages/pulumiservice/api-docs/stacktags/) for full "+
			"usage and examples.\n",
	)
	a.SetToken("index", "StackTags")
}

type StackTagsInput struct {
	Organization string            `pulumi:"organization" provider:"replaceOnChanges"`
	Project      string            `pulumi:"project"      provider:"replaceOnChanges"`
	Stack        string            `pulumi:"stack"        provider:"replaceOnChanges"`
	Tags         map[string]string `pulumi:"tags"`
}

func (i *StackTagsInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.Describe(&i.Stack, "Stack name.")
	a.Describe(&i.Tags, "Map of tag names to values. Each entry represents a stack tag.")
}

type StackTagsState = StackTagsInput

func stackTagsResourceID(organization, project, stack string) string {
	return path.Join(organization, project, stack, "tags")
}

// parseStackTagsResourceID parses an ID of the form `organization/project/stack/tags`.
func parseStackTagsResourceID(id string) (organization, project, stack string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 || parts[3] != "tags" {
		return "", "", "", fmt.Errorf("%q is invalid, must be in organization/project/stack/tags format", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func (*StackTags) Create(
	ctx context.Context,
	req infer.CreateRequest[StackTagsInput],
) (infer.CreateResponse[StackTagsState], error) {
	id := stackTagsResourceID(req.Inputs.Organization, req.Inputs.Project, req.Inputs.Stack)

	if req.DryRun {
		return infer.CreateResponse[StackTagsState]{
			ID:     id,
			Output: req.Inputs,
		}, nil
	}

	client := config.GetClient(ctx)
	stackIdentifier := pulumiapi.StackIdentifier{
		OrgName:     req.Inputs.Organization,
		ProjectName: req.Inputs.Project,
		StackName:   req.Inputs.Stack,
	}

	// Create tags in sorted order so partial failures are deterministic.
	created := map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(req.Inputs.Tags)) {
		value := req.Inputs.Tags[name]
		if err := client.CreateStackTag(ctx, stackIdentifier, pulumiapi.StackTag{Name: name, Value: value}); err != nil {
			partial := req.Inputs
			partial.Tags = created
			return infer.CreateResponse[StackTagsState]{
					ID:     id,
					Output: partial,
				}, infer.ResourceInitFailedError{
					Reasons: []string{fmt.Sprintf("failed to create tag %q: %s", name, err.Error())},
				}
		}
		created[name] = value
	}

	return infer.CreateResponse[StackTagsState]{
		ID:     id,
		Output: req.Inputs,
	}, nil
}

func (*StackTags) Update(
	ctx context.Context,
	req infer.UpdateRequest[StackTagsInput, StackTagsState],
) (infer.UpdateResponse[StackTagsState], error) {
	if req.DryRun {
		return infer.UpdateResponse[StackTagsState]{Output: req.Inputs}, nil
	}

	client := config.GetClient(ctx)
	stackIdentifier := pulumiapi.StackIdentifier{
		OrgName:     req.Inputs.Organization,
		ProjectName: req.Inputs.Project,
		StackName:   req.Inputs.Stack,
	}

	// Compute the add/remove/modify sets. Pulumi Cloud tags are immutable per
	// key, so a value change is a delete + create.
	tagsToDelete := []string{}
	tagsToCreate := map[string]string{}

	for oldName, oldValue := range req.State.Tags {
		if newValue, exists := req.Inputs.Tags[oldName]; !exists {
			tagsToDelete = append(tagsToDelete, oldName)
		} else if newValue != oldValue {
			tagsToDelete = append(tagsToDelete, oldName)
			tagsToCreate[oldName] = newValue
		}
	}
	for newName, newValue := range req.Inputs.Tags {
		if _, exists := req.State.Tags[newName]; !exists {
			tagsToCreate[newName] = newValue
		}
	}

	// Track the live tag set as we mutate it. If any API call fails we return
	// this as the resource state so Pulumi knows exactly which tags exist.
	currentTags := maps.Clone(req.State.Tags)
	if currentTags == nil {
		currentTags = map[string]string{}
	}

	partialOutput := func() StackTagsState {
		state := req.Inputs
		state.Tags = maps.Clone(currentTags)
		return state
	}

	sort.Strings(tagsToDelete)
	for _, tagName := range tagsToDelete {
		if err := client.DeleteStackTag(ctx, stackIdentifier, tagName); err != nil {
			return infer.UpdateResponse[StackTagsState]{Output: partialOutput()},
				infer.ResourceInitFailedError{
					Reasons: []string{fmt.Sprintf("failed to delete tag %q: %s", tagName, err.Error())},
				}
		}
		delete(currentTags, tagName)
	}

	for _, name := range slices.Sorted(maps.Keys(tagsToCreate)) {
		value := tagsToCreate[name]
		if err := client.CreateStackTag(ctx, stackIdentifier, pulumiapi.StackTag{Name: name, Value: value}); err != nil {
			return infer.UpdateResponse[StackTagsState]{Output: partialOutput()},
				infer.ResourceInitFailedError{
					Reasons: []string{fmt.Sprintf("failed to create tag %q: %s", name, err.Error())},
				}
		}
		currentTags[name] = value
	}

	return infer.UpdateResponse[StackTagsState]{Output: req.Inputs}, nil
}

func (*StackTags) Delete(
	ctx context.Context,
	req infer.DeleteRequest[StackTagsState],
) (infer.DeleteResponse, error) {
	client := config.GetClient(ctx)
	stackIdentifier := pulumiapi.StackIdentifier{
		OrgName:     req.State.Organization,
		ProjectName: req.State.Project,
		StackName:   req.State.Stack,
	}

	for _, tagName := range slices.Sorted(maps.Keys(req.State.Tags)) {
		if err := client.DeleteStackTag(ctx, stackIdentifier, tagName); err != nil {
			return infer.DeleteResponse{}, fmt.Errorf("failed to delete tag %q: %w", tagName, err)
		}
	}
	return infer.DeleteResponse{}, nil
}

func (*StackTags) Read(
	ctx context.Context,
	req infer.ReadRequest[StackTagsInput, StackTagsState],
) (infer.ReadResponse[StackTagsInput, StackTagsState], error) {
	organization, project, stack, err := parseStackTagsResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[StackTagsInput, StackTagsState]{}, err
	}

	client := config.GetClient(ctx)
	allTags, err := client.GetStackTags(ctx, pulumiapi.StackIdentifier{
		OrgName:     organization,
		ProjectName: project,
		StackName:   stack,
	})
	if err != nil {
		return infer.ReadResponse[StackTagsInput, StackTagsState]{},
			fmt.Errorf("failed to read stack tags (%q): %w", req.ID, err)
	}

	managedTags := map[string]string{}
	if len(req.Inputs.Tags) > 0 {
		// We have previous state; only return tags we previously managed.
		for tagName := range req.Inputs.Tags {
			if value, exists := allTags[tagName]; exists {
				managedTags[tagName] = value
			}
		}
	} else {
		// No previous state (import scenario): adopt every tag currently on
		// the stack. The user is expected to declare the `tags` map explicitly
		// after import so subsequent updates match their intent.
		managedTags = maps.Clone(allTags)
		if managedTags == nil {
			managedTags = map[string]string{}
		}
	}

	state := StackTagsState{
		Organization: organization,
		Project:      project,
		Stack:        stack,
		Tags:         managedTags,
	}

	return infer.ReadResponse[StackTagsInput, StackTagsState]{
		ID:     req.ID,
		Inputs: state,
		State:  state,
	}, nil
}
