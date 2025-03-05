package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceEnvironmentVersionTagResource struct {
	Client esc_client.Client
}

type PulumiServiceEnvironmentVersionTagInput struct {
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Environment  string `pulumi:"environment"`
	TagName      string `pulumi:"tagName"`
	Revision     int    `pulumi:"revision"`
}

func (i *PulumiServiceEnvironmentVersionTagInput) ToPropertyMap() resource.PropertyMap {
	return util.ToPropertyMap(*i, structTagKey)
}

func (evt *PulumiServiceEnvironmentVersionTagResource) ToPulumiServiceEnvironmentVersionTagInput(properties *structpb.Struct) PulumiServiceEnvironmentVersionTagInput {
	input := PulumiServiceEnvironmentVersionTagInput{}
	_ = util.FromProperties(properties, structTagKey, &input)

	if input.Project == "" {
		input.Project = defaultProject
	}

	return input
}

func (evt *PulumiServiceEnvironmentVersionTagResource) Name() string {
	return "pulumiservice:index:EnvironmentVersionTag"
}

func (evt *PulumiServiceEnvironmentVersionTagResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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
	replaces := []string(nil)
	replaceProperties := map[string]bool{
		"organization": true,
		"project":      true,
		"environment":  true,
		"tagName":      true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		Replaces:            replaces,
		DetailedDiff:        detailedDiffs,
		HasDetailedDiff:     true,
		DeleteBeforeReplace: len(replaces) > 0,
	}, nil
}

func (evt *PulumiServiceEnvironmentVersionTagResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	input := evt.ToPulumiServiceEnvironmentVersionTagInput(req.GetProperties())

	err := evt.Client.DeleteEnvironmentRevisionTag(ctx, input.Organization, input.Project, input.Environment, input.TagName)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (evt *PulumiServiceEnvironmentVersionTagResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()

	input := evt.ToPulumiServiceEnvironmentVersionTagInput(req.GetProperties())

	err := evt.Client.CreateEnvironmentRevisionTag(ctx, input.Organization, input.Project, input.Environment, input.TagName, &input.Revision)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.Organization, input.Project, input.Environment, input.TagName),
		Properties: req.GetProperties(),
	}, nil
}

func (evt *PulumiServiceEnvironmentVersionTagResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "environment", "tagName", "revision"} {
		if !inputMap[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.GetNews(), Failures: failures}, nil
}

func (evt *PulumiServiceEnvironmentVersionTagResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	var input PulumiServiceEnvironmentVersionTagInput
	err := util.FromProperties(req.GetNews(), structTagKey, &input)
	if err != nil {
		return nil, err
	}

	err = evt.Client.UpdateEnvironmentRevisionTag(ctx, input.Organization, input.Project, input.Environment, input.TagName, &input.Revision)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: req.GetNews(),
	}, nil
}

func (evt *PulumiServiceEnvironmentVersionTagResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	orgName, projectName, environmentName, tagName, err := splitEnvironmentTagId(req.Id)
	if err != nil {
		return nil, err
	}

	tag, err := evt.Client.GetEnvironmentRevisionTag(ctx, orgName, projectName, environmentName, tagName)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return nil, fmt.Errorf("failed to read EnvironmentVersionTag (%q): %w", req.Id, err)
	}
	if tag == nil {
		// if the tag doesn't exist, then return empty response
		return &pulumirpc.ReadResponse{}, nil
	}

	input := PulumiServiceEnvironmentVersionTagInput{
		Organization: orgName,
		Project:      projectName,
		Environment:  environmentName,
		TagName:      tagName,
		Revision:     tag.Revision,
	}

	props, err := util.ToProperties(input, structTagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}

func splitEnvironmentTagId(id string) (string, string, string, string, error) {
	// format:
	//   organization/project/environment/tag or
	//   organization/environment/tag (legacy)

	var org, project, environment, tag string

	s := strings.Split(id, "/")
	switch len(s) {
	case 4:
		org = s[0]
		project = s[1]
		environment = s[2]
		tag = s[3]
	case 3:
		// Legacy pattern. Assume "default" project
		org = s[0]
		project = defaultProject
		environment = s[1]
		tag = s[2]
	default:
		return "", "", "", "", fmt.Errorf("%q is invalid, must be in the format: organization/environment/tag", id)
	}

	return org, project, environment, tag, nil
}
