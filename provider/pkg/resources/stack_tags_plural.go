package resources

import (
	"context"
	"fmt"
	"maps"
	"path"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil/rpcerror"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
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

func (i *PulumiServiceStackTagsInput) ToPropertyMap() resource.PropertyMap {
	return util.ToPropertyMap(*i)
}

func (i *PulumiServiceStackTagsInput) ToRPC() (*structpb.Struct, error) {
	return plugin.MarshalProperties(i.ToPropertyMap(), plugin.MarshalOptions{
		KeepOutputValues: true,
	})
}

func (st *PulumiServiceStackTagsResource) Name() string {
	return "pulumiservice:index:StackTags"
}

func (st *PulumiServiceStackTagsResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (st *PulumiServiceStackTagsResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(
		req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
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
		if k == "organization" || k == "project" || k == "stack" {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec // safe conversion from plugin.DiffKind
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
	err := util.FromProperties(req.GetProperties(), &input)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     input.Organization,
		ProjectName: input.Project,
		StackName:   input.Stack,
	}
	id := path.Join(input.Organization, input.Project, input.Stack, "tags")

	// Create tags in sorted order so partial failures are deterministic.
	created := map[string]string{}
	tagNames := sortedKeys(input.Tags)
	for _, name := range tagNames {
		err := st.Client.CreateStackTag(ctx, stackName, pulumiapi.StackTag{Name: name, Value: input.Tags[name]})
		if err != nil {
			// Record the subset of tags that were successfully created so Pulumi
			// can track them in state and the next `up` resumes from the right place.
			// partial := input copies the value struct, then partial.Tags = created
			// swaps in the subset without mutating the caller's input.
			partial := input
			partial.Tags = created
			return nil, partialErrorStackTags(id, fmt.Errorf("failed to create tag %q: %w", name, err), partial, input)
		}
		created[name] = input.Tags[name]
	}

	return &pulumirpc.CreateResponse{
		Id:         id,
		Properties: req.GetProperties(),
	}, nil
}

func (st *PulumiServiceStackTagsResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	organization, project, stack, err := parseStackTagsID(req.Id)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     organization,
		ProjectName: project,
		StackName:   stack,
	}

	allTags, err := st.Client.GetStackTags(ctx, stackName)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack tags (%q): %w", req.Id, err)
	}

	// Parse current inputs to know which tags we manage.
	var currentInput PulumiServiceStackTagsInput
	if req.Inputs != nil {
		inputMap, err := plugin.UnmarshalProperties(req.Inputs, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
		}
		if err := util.FromPropertyMap(inputMap, &currentInput); err != nil {
			// If we can't decode the previous inputs, don't silently fall into
			// the import/adopt-all-tags path — that would let us take ownership
			// of (and later delete) tags we never managed.
			return nil, fmt.Errorf("failed to decode stored inputs for %q: %w", req.Id, err)
		}
	}

	managedTags := make(map[string]string)
	if len(currentInput.Tags) > 0 {
		// We have previous state, only return tags we previously managed.
		for tagName := range currentInput.Tags {
			if value, exists := allTags[tagName]; exists {
				managedTags[tagName] = value
			}
		}
	} else {
		// No previous state (import scenario): adopt every tag currently on the stack.
		// The user is expected to declare the `tags` map explicitly after import so
		// subsequent updates match their intent — see the schema description.
		// Clone so we don't alias the API-client-owned map.
		managedTags = maps.Clone(allTags)
	}

	state := PulumiServiceStackTagsInput{
		Organization: organization,
		Project:      project,
		Stack:        stack,
		Tags:         managedTags,
	}

	props, err := util.ToProperties(state)
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
	err := util.FromProperties(req.GetOlds(), &oldInput)
	if err != nil {
		return nil, err
	}

	var newInput PulumiServiceStackTagsInput
	err = util.FromProperties(req.GetNews(), &newInput)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     newInput.Organization,
		ProjectName: newInput.Project,
		StackName:   newInput.Stack,
	}
	id := path.Join(newInput.Organization, newInput.Project, newInput.Stack, "tags")

	// Compute the add/remove/modify sets. Pulumi Cloud tags are immutable per
	// key, so a value change is a delete + create.
	tagsToDelete := []string{}
	tagsToCreate := map[string]string{}

	for oldName, oldValue := range oldInput.Tags {
		if newValue, exists := newInput.Tags[oldName]; !exists {
			tagsToDelete = append(tagsToDelete, oldName)
		} else if newValue != oldValue {
			tagsToDelete = append(tagsToDelete, oldName)
			tagsToCreate[oldName] = newValue
		}
	}
	for newName, newValue := range newInput.Tags {
		if _, exists := oldInput.Tags[newName]; !exists {
			tagsToCreate[newName] = newValue
		}
	}

	// Track the live tag set as we mutate it. If any API call fails we return
	// this as the resource state so Pulumi knows exactly which tags exist.
	currentTags := map[string]string{}
	for k, v := range oldInput.Tags {
		currentTags[k] = v
	}

	sort.Strings(tagsToDelete)
	for _, tagName := range tagsToDelete {
		err := st.Client.DeleteStackTag(ctx, stackName, tagName)
		if err != nil {
			state := newInput
			state.Tags = currentTags
			return nil, partialErrorStackTags(id, fmt.Errorf("failed to delete tag %q: %w", tagName, err), state, newInput)
		}
		delete(currentTags, tagName)
	}

	createNames := sortedKeys(tagsToCreate)
	for _, name := range createNames {
		value := tagsToCreate[name]
		err := st.Client.CreateStackTag(ctx, stackName, pulumiapi.StackTag{Name: name, Value: value})
		if err != nil {
			state := newInput
			state.Tags = currentTags
			return nil, partialErrorStackTags(id, fmt.Errorf("failed to create tag %q: %w", name, err), state, newInput)
		}
		currentTags[name] = value
	}

	return &pulumirpc.UpdateResponse{
		Properties: req.GetNews(),
	}, nil
}

func (st *PulumiServiceStackTagsResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var input PulumiServiceStackTagsInput
	err := util.FromProperties(req.GetProperties(), &input)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     input.Organization,
		ProjectName: input.Project,
		StackName:   input.Stack,
	}

	tagNames := sortedKeys(input.Tags)
	for _, tagName := range tagNames {
		err = st.Client.DeleteStackTag(ctx, stackName, tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to delete tag %q: %w", tagName, err)
		}
	}

	return &pbempty.Empty{}, nil
}

// parseStackTagsID parses an ID of the form `organization/project/stack/tags`.
func parseStackTagsID(id string) (organization, project, stack string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 || parts[3] != "tags" {
		return "", "", "", fmt.Errorf("%q is invalid, must be in organization/project/stack/tags format", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// partialErrorStackTags wraps err so Pulumi records `state` as the resource's
// last known state before failure. Without this, a Create that creates 2 of 5
// tags and then fails would leave those 2 untracked and re-create them on retry.
func partialErrorStackTags(id string, err error, state, inputs PulumiServiceStackTagsInput) error {
	stateRPC, stateSerErr := state.ToRPC()
	inputRPC, inputSerErr := inputs.ToRPC()
	if stateSerErr != nil {
		err = fmt.Errorf("err serializing state: %v (src error: %v)", stateSerErr, err)
	}
	if inputSerErr != nil {
		err = fmt.Errorf("err serializing inputs: %v (src error: %v)", inputSerErr, err)
	}
	detail := pulumirpc.ErrorResourceInitFailed{
		Id:         id,
		Properties: stateRPC,
		Reasons:    []string{err.Error()},
		Inputs:     inputRPC,
	}
	return rpcerror.WithDetails(rpcerror.New(codes.Unknown, err.Error()), &detail)
}
