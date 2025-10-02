package resources

import (
	"context"
	"fmt"
	"path"
	"sort"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceStackTagsResource struct {
	Client *pulumiapi.Client
}

type PulumiServiceStackTagsInput struct {
	Organization string            `pulumi:"organization"`
	Project      string            `pulumi:"project"`
	Stack        string            `pulumi:"stack"`
	Tags         map[string]string `pulumi:"tags"`
}

const stackTagsStructTagKey = "pulumi" // could also be "json"

func (i *PulumiServiceStackTagsInput) ToPropertyMap() resource.PropertyMap {
	return util.ToPropertyMap(*i, stackTagsStructTagKey)
}

func (st *PulumiServiceStackTagsResource) ToPulumiServiceStackTagsInput(inputMap resource.PropertyMap) PulumiServiceStackTagsInput {
	input := PulumiServiceStackTagsInput{}
	_ = util.FromPropertyMap(inputMap, stackTagsStructTagKey, &input)
	return input
}

func (st *PulumiServiceStackTagsResource) Name() string {
	return "pulumiservice:index:StackTags"
}

func (st *PulumiServiceStackTagsResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (st *PulumiServiceStackTagsResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaces := []string{}

	for k, v := range dd {
		// Organization, project, stack changes require replacement
		if k == "organization" || k == "project" || k == "stack" {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		// Tag changes can be updated in place
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_SOME
	if len(detailedDiffs) == 0 {
		changes = pulumirpc.DiffResponse_DIFF_NONE
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            replaces,
		DetailedDiff:        detailedDiffs,
		DeleteBeforeReplace: len(replaces) > 0,
		HasDetailedDiff:     true,
	}, nil
}

func (st *PulumiServiceStackTagsResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	var input PulumiServiceStackTagsInput
	err := util.FromProperties(req.GetProperties(), stackTagsStructTagKey, &input)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     input.Organization,
		ProjectName: input.Project,
		StackName:   input.Stack,
	}

	// Create each tag in the map
	for name, value := range input.Tags {
		stackTag := pulumiapi.StackTag{
			Name:  name,
			Value: value,
		}
		err = st.Client.CreateTag(ctx, stackName, stackTag)
		if err != nil {
			return nil, fmt.Errorf("failed to create tag %q: %w", name, err)
		}
	}

	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.Organization, input.Project, input.Stack, "tags"),
		Properties: req.GetProperties(),
	}, nil
}

func (st *PulumiServiceStackTagsResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	// Parse ID: organization/project/stack/tags
	organization, project, stack, _, err := splitStackTagId(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     organization,
		ProjectName: project,
		StackName:   stack,
	}

	// Get all tags from the stack
	allTags, err := st.Client.GetStackTags(ctx, stackName)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack tags (%q): %w", req.Id, err)
	}

	// Parse current inputs to know which tags we manage
	var currentInput PulumiServiceStackTagsInput
	if req.Inputs != nil {
		inputMap, err := plugin.UnmarshalProperties(req.Inputs, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
		}
		_ = util.FromPropertyMap(inputMap, stackTagsStructTagKey, &currentInput)
	}

	// Filter to only the tags we manage
	managedTags := make(map[string]string)
	if len(currentInput.Tags) > 0 {
		// We have previous state, only return tags we previously managed
		for tagName := range currentInput.Tags {
			if value, exists := allTags[tagName]; exists {
				managedTags[tagName] = value
			}
		}
	} else {
		// No previous state (import scenario), return all tags
		managedTags = allTags
	}

	input := PulumiServiceStackTagsInput{
		Organization: organization,
		Project:      project,
		Stack:        stack,
		Tags:         managedTags,
	}

	props, err := util.ToProperties(input, stackTagsStructTagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}

func (st *PulumiServiceStackTagsResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()

	var oldInput PulumiServiceStackTagsInput
	err := util.FromProperties(req.GetOlds(), stackTagsStructTagKey, &oldInput)
	if err != nil {
		return nil, err
	}

	var newInput PulumiServiceStackTagsInput
	err = util.FromProperties(req.GetNews(), stackTagsStructTagKey, &newInput)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     newInput.Organization,
		ProjectName: newInput.Project,
		StackName:   newInput.Stack,
	}

	// Determine which tags to add, update, and delete
	tagsToDelete := []string{}
	tagsToCreate := map[string]string{}

	// Find tags that were removed or changed
	for oldName, oldValue := range oldInput.Tags {
		if newValue, exists := newInput.Tags[oldName]; !exists {
			// Tag was removed
			tagsToDelete = append(tagsToDelete, oldName)
		} else if newValue != oldValue {
			// Tag value changed - delete old and create new (tags are immutable)
			tagsToDelete = append(tagsToDelete, oldName)
			tagsToCreate[oldName] = newValue
		}
	}

	// Find tags that were added
	for newName, newValue := range newInput.Tags {
		if _, exists := oldInput.Tags[newName]; !exists {
			tagsToCreate[newName] = newValue
		}
	}

	// Delete tags (in sorted order for deterministic behavior)
	sort.Strings(tagsToDelete)
	for _, tagName := range tagsToDelete {
		err = st.Client.DeleteStackTag(ctx, stackName, tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to delete tag %q: %w", tagName, err)
		}
	}

	// Create/update tags (in sorted order for deterministic behavior)
	tagNames := make([]string, 0, len(tagsToCreate))
	for name := range tagsToCreate {
		tagNames = append(tagNames, name)
	}
	sort.Strings(tagNames)

	for _, name := range tagNames {
		value := tagsToCreate[name]
		stackTag := pulumiapi.StackTag{
			Name:  name,
			Value: value,
		}
		err = st.Client.CreateTag(ctx, stackName, stackTag)
		if err != nil {
			return nil, fmt.Errorf("failed to create/update tag %q: %w", name, err)
		}
	}

	return &pulumirpc.UpdateResponse{
		Properties: req.GetNews(),
	}, nil
}

func (st *PulumiServiceStackTagsResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var input PulumiServiceStackTagsInput
	err := util.FromProperties(req.GetProperties(), stackTagsStructTagKey, &input)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     input.Organization,
		ProjectName: input.Project,
		StackName:   input.Stack,
	}

	// Delete all managed tags (in sorted order for deterministic behavior)
	tagNames := make([]string, 0, len(input.Tags))
	for name := range input.Tags {
		tagNames = append(tagNames, name)
	}
	sort.Strings(tagNames)

	for _, tagName := range tagNames {
		err = st.Client.DeleteStackTag(ctx, stackName, tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to delete tag %q: %w", tagName, err)
		}
	}

	return &pbempty.Empty{}, nil
}
