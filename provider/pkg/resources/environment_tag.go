// Copyright 2016-2026, Pulumi Corporation.

package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type PulumiServiceEnvironmentTagResource struct {
	Client esc_client.Client
}

type PulumiServiceEnvironmentTagInput struct {
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Environment  string `pulumi:"environment"`
	TagName      string `pulumi:"tagName"`
	Value        string `pulumi:"value"`
}

func (i *PulumiServiceEnvironmentTagInput) ToPropertyMap() resource.PropertyMap {
	return util.ToPropertyMap(*i)
}

func (et *PulumiServiceEnvironmentTagResource) ToPulumiServiceEnvironmentTagInput(
	properties *structpb.Struct,
) PulumiServiceEnvironmentTagInput {
	input := PulumiServiceEnvironmentTagInput{}
	_ = util.FromProperties(properties, &input)

	if input.Project == "" {
		input.Project = defaultProject
	}

	return input
}

func (et *PulumiServiceEnvironmentTagResource) Name() string {
	return "pulumiservice:index:EnvironmentTag"
}

func (et *PulumiServiceEnvironmentTagResource) Diff(
	req *pulumirpc.DiffRequest,
) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
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
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec // safe conversion from plugin.DiffKind
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

func (et *PulumiServiceEnvironmentTagResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	input := et.ToPulumiServiceEnvironmentTagInput(req.GetProperties())

	err := et.Client.DeleteEnvironmentTag(
		ctx,
		input.Organization,
		input.Project,
		input.Environment,
		input.TagName,
	)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (et *PulumiServiceEnvironmentTagResource) Create(
	req *pulumirpc.CreateRequest,
) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()

	input := et.ToPulumiServiceEnvironmentTagInput(req.GetProperties())

	_, err := et.Client.CreateEnvironmentTag(
		ctx,
		input.Organization,
		input.Project,
		input.Environment,
		input.TagName,
		input.Value,
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.Organization, input.Project, input.Environment, input.TagName),
		Properties: req.GetProperties(),
	}, nil
}

func (et *PulumiServiceEnvironmentTagResource) Check(
	req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "environment", "tagName", "value"} {
		if !inputMap[p].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.GetNews(), Failures: failures}, nil
}

func (et *PulumiServiceEnvironmentTagResource) Update(
	req *pulumirpc.UpdateRequest,
) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()

	var olds PulumiServiceEnvironmentTagInput
	if err := util.FromProperties(req.GetOlds(), &olds); err != nil {
		return nil, err
	}
	if olds.Project == "" {
		olds.Project = defaultProject
	}

	var news PulumiServiceEnvironmentTagInput
	if err := util.FromProperties(req.GetNews(), &news); err != nil {
		return nil, err
	}
	if news.Project == "" {
		news.Project = defaultProject
	}

	// Only `value` is updatable; other fields trigger replacement via Diff.
	// currentValue must come from old state so the API's concurrency check passes
	// even if the property has drifted server-side. newKey is left empty so the
	// ESC client preserves the existing tag name.
	_, err := et.Client.UpdateEnvironmentTag(
		ctx,
		news.Organization,
		news.Project,
		news.Environment,
		news.TagName,
		olds.Value,
		"",
		news.Value,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: req.GetNews(),
	}, nil
}

func (et *PulumiServiceEnvironmentTagResource) Read(
	req *pulumirpc.ReadRequest,
) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	orgName, projectName, environmentName, tagName, err := splitEnvironmentTagID(req.Id)
	if err != nil {
		return nil, err
	}

	tag, err := et.Client.GetEnvironmentTag(ctx, orgName, projectName, environmentName, tagName)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return nil, fmt.Errorf("failed to read EnvironmentTag (%q): %w", req.Id, err)
	}
	if tag == nil {
		// if the tag doesn't exist, then return empty response
		return &pulumirpc.ReadResponse{}, nil
	}

	input := PulumiServiceEnvironmentTagInput{
		Organization: orgName,
		Project:      projectName,
		Environment:  environmentName,
		TagName:      tag.Name,
		Value:        tag.Value,
	}

	props, err := util.ToProperties(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}
