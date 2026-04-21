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
	"path"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type StackTag struct{}

var (
	_ infer.CustomCreate[StackTagInput, StackTagState] = &StackTag{}
	_ infer.CustomDelete[StackTagState]                = &StackTag{}
	_ infer.CustomRead[StackTagInput, StackTagState]   = &StackTag{}

	_ infer.Annotated = &StackTag{}
	_ infer.Annotated = &StackTagInput{}
)

func (s *StackTag) Annotate(a infer.Annotator) {
	a.Describe(s, "Stacks have associated metadata in the form of tags. Each tag consists of a name and value.")
}

type StackTagInput struct {
	Organization string `pulumi:"organization"  provider:"replaceOnChanges"`
	Project      string `pulumi:"project"       provider:"replaceOnChanges"`
	Stack        string `pulumi:"stack"         provider:"replaceOnChanges"`
	Name         string `pulumi:"name"          provider:"replaceOnChanges"`
	Value        string `pulumi:"value"         provider:"replaceOnChanges"`
}

func (i *StackTagInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.Describe(&i.Stack, "Stack name.")
	a.Describe(&i.Name, "Name of the tag. The 'key' part of the key=value pair")
	a.Describe(&i.Value, "Value of the tag. The 'value' part of the key=value pair")
}

type StackTagState = StackTagInput

func (*StackTag) Create(
	ctx context.Context,
	req infer.CreateRequest[StackTagInput],
) (infer.CreateResponse[StackTagState], error) {
	id := path.Join(req.Inputs.Organization, req.Inputs.Project, req.Inputs.Stack, req.Inputs.Name)

	if req.DryRun {
		return infer.CreateResponse[StackTagState]{
			ID:     id,
			Output: req.Inputs,
		}, nil
	}

	client := config.GetClient(ctx)
	err := client.CreateStackTag(ctx,
		pulumiapi.StackIdentifier{
			OrgName:     req.Inputs.Organization,
			ProjectName: req.Inputs.Project,
			StackName:   req.Inputs.Stack,
		},
		pulumiapi.StackTag{
			Name:  req.Inputs.Name,
			Value: req.Inputs.Value,
		},
	)
	if err != nil {
		return infer.CreateResponse[StackTagState]{}, err
	}

	return infer.CreateResponse[StackTagState]{
		ID:     id,
		Output: req.Inputs,
	}, nil
}

func (*StackTag) Delete(ctx context.Context, req infer.DeleteRequest[StackTagState]) (infer.DeleteResponse, error) {
	client := config.GetClient(ctx)
	return infer.DeleteResponse{}, client.DeleteStackTag(ctx,
		pulumiapi.StackIdentifier{
			OrgName:     req.State.Organization,
			ProjectName: req.State.Project,
			StackName:   req.State.Stack,
		},
		req.State.Name,
	)
}

func (*StackTag) Read(
	ctx context.Context,
	req infer.ReadRequest[StackTagInput, StackTagState],
) (infer.ReadResponse[StackTagInput, StackTagState], error) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 4 {
		return infer.ReadResponse[StackTagInput, StackTagState]{},
			fmt.Errorf("%q is invalid, must be in organization/project/stack/tagName format", req.ID)
	}
	organization, project, stack, tagName := parts[0], parts[1], parts[2], parts[3]

	client := config.GetClient(ctx)
	tag, err := client.GetStackTag(ctx,
		pulumiapi.StackIdentifier{
			OrgName:     organization,
			ProjectName: project,
			StackName:   stack,
		},
		tagName,
	)
	if err != nil {
		return infer.ReadResponse[StackTagInput, StackTagState]{},
			fmt.Errorf("failed to read StackTag (%q): %w", req.ID, err)
	}
	if tag == nil {
		return infer.ReadResponse[StackTagInput, StackTagState]{}, nil
	}

	state := StackTagState{
		Organization: organization,
		Project:      project,
		Stack:        stack,
		Name:         tag.Name,
		Value:        tag.Value,
	}

	return infer.ReadResponse[StackTagInput, StackTagState]{
		ID:     req.ID,
		Inputs: state,
		State:  state,
	}, nil
}
