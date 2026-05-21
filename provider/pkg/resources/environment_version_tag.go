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
)

type EnvironmentVersionTag struct{}

var (
	_ infer.CustomCreate[EnvironmentVersionTagInput, EnvironmentVersionTagState] = &EnvironmentVersionTag{}
	_ infer.CustomUpdate[EnvironmentVersionTagInput, EnvironmentVersionTagState] = &EnvironmentVersionTag{}
	_ infer.CustomDelete[EnvironmentVersionTagState]                             = &EnvironmentVersionTag{}
	_ infer.CustomRead[EnvironmentVersionTagInput, EnvironmentVersionTagState]   = &EnvironmentVersionTag{}
)

func (*EnvironmentVersionTag) Annotate(a infer.Annotator) {
	a.Describe(&EnvironmentVersionTag{}, "A tag on a specific revision of an environment.")
	a.SetToken("index", "EnvironmentVersionTag")
}

type EnvironmentVersionTagInput struct {
	Organization string `pulumi:"organization"      provider:"replaceOnChanges"`
	Project      string `pulumi:"project,optional"  provider:"replaceOnChanges"`
	Environment  string `pulumi:"environment"       provider:"replaceOnChanges"`
	TagName      string `pulumi:"tagName"           provider:"replaceOnChanges"`
	Revision     int    `pulumi:"revision"`
}

func (i *EnvironmentVersionTagInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.SetDefault(&i.Project, defaultProject)
	a.Describe(&i.Environment, "Environment name.")
	a.Describe(&i.TagName, "Tag name.")
	a.Describe(&i.Revision, "Revision number.")
}

type EnvironmentVersionTagState struct {
	EnvironmentVersionTagInput
}

func (*EnvironmentVersionTag) Create(
	ctx context.Context,
	req infer.CreateRequest[EnvironmentVersionTagInput],
) (infer.CreateResponse[EnvironmentVersionTagState], error) {
	if req.DryRun {
		return infer.CreateResponse[EnvironmentVersionTagState]{
			Output: EnvironmentVersionTagState{EnvironmentVersionTagInput: req.Inputs},
		}, nil
	}
	revision := req.Inputs.Revision
	err := config.GetEscClient(ctx).CreateEnvironmentRevisionTag(
		ctx,
		req.Inputs.Organization,
		req.Inputs.Project,
		req.Inputs.Environment,
		req.Inputs.TagName,
		&revision,
	)
	if err != nil {
		return infer.CreateResponse[EnvironmentVersionTagState]{}, fmt.Errorf(
			"error creating environment version tag %q: %w", req.Inputs.TagName, err,
		)
	}
	return infer.CreateResponse[EnvironmentVersionTagState]{
		ID: environmentVersionTagID(
			req.Inputs.Organization, req.Inputs.Project, req.Inputs.Environment, req.Inputs.TagName,
		),
		Output: EnvironmentVersionTagState{EnvironmentVersionTagInput: req.Inputs},
	}, nil
}

func (*EnvironmentVersionTag) Update(
	ctx context.Context,
	req infer.UpdateRequest[EnvironmentVersionTagInput, EnvironmentVersionTagState],
) (infer.UpdateResponse[EnvironmentVersionTagState], error) {
	if req.DryRun {
		return infer.UpdateResponse[EnvironmentVersionTagState]{
			Output: EnvironmentVersionTagState{EnvironmentVersionTagInput: req.Inputs},
		}, nil
	}
	revision := req.Inputs.Revision
	err := config.GetEscClient(ctx).UpdateEnvironmentRevisionTag(
		ctx,
		req.Inputs.Organization,
		req.Inputs.Project,
		req.Inputs.Environment,
		req.Inputs.TagName,
		&revision,
	)
	if err != nil {
		return infer.UpdateResponse[EnvironmentVersionTagState]{}, fmt.Errorf(
			"error updating environment version tag %q: %w", req.Inputs.TagName, err,
		)
	}
	return infer.UpdateResponse[EnvironmentVersionTagState]{
		Output: EnvironmentVersionTagState{EnvironmentVersionTagInput: req.Inputs},
	}, nil
}

func (*EnvironmentVersionTag) Delete(
	ctx context.Context,
	req infer.DeleteRequest[EnvironmentVersionTagState],
) (infer.DeleteResponse, error) {
	err := config.GetEscClient(ctx).DeleteEnvironmentRevisionTag(
		ctx,
		req.State.Organization,
		req.State.Project,
		req.State.Environment,
		req.State.TagName,
	)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf(
			"error deleting environment version tag %q: %w", req.State.TagName, err,
		)
	}
	return infer.DeleteResponse{}, nil
}

func (*EnvironmentVersionTag) Read(
	ctx context.Context,
	req infer.ReadRequest[EnvironmentVersionTagInput, EnvironmentVersionTagState],
) (infer.ReadResponse[EnvironmentVersionTagInput, EnvironmentVersionTagState], error) {
	orgName, projectName, environmentName, tagName, err := splitEnvironmentVersionTagID(req.ID)
	if err != nil {
		return infer.ReadResponse[EnvironmentVersionTagInput, EnvironmentVersionTagState]{}, err
	}

	tag, err := config.GetEscClient(ctx).GetEnvironmentRevisionTag(ctx, orgName, projectName, environmentName, tagName)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return infer.ReadResponse[EnvironmentVersionTagInput, EnvironmentVersionTagState]{},
			fmt.Errorf("failed to read EnvironmentVersionTag (%q): %w", req.ID, err)
	}
	if tag == nil {
		return infer.ReadResponse[EnvironmentVersionTagInput, EnvironmentVersionTagState]{}, nil
	}
	inputs := EnvironmentVersionTagInput{
		Organization: orgName,
		Project:      projectName,
		Environment:  environmentName,
		TagName:      tagName,
		Revision:     tag.Revision,
	}
	return infer.ReadResponse[EnvironmentVersionTagInput, EnvironmentVersionTagState]{
		ID:     environmentVersionTagID(orgName, projectName, environmentName, tagName),
		Inputs: inputs,
		State:  EnvironmentVersionTagState{EnvironmentVersionTagInput: inputs},
	}, nil
}

func environmentVersionTagID(organization, project, environment, tagName string) string {
	return path.Join(organization, project, environment, tagName)
}

func splitEnvironmentVersionTagID(id string) (string, string, string, string, error) {
	// Format:
	//   organization/project/environment/tag or
	//   organization/environment/tag (legacy)
	s := strings.Split(id, "/")
	switch len(s) {
	case 4:
		return s[0], s[1], s[2], s[3], nil
	case 3:
		// Legacy pattern. Assume "default" project.
		return s[0], defaultProject, s[1], s[2], nil
	default:
		return "", "", "", "", fmt.Errorf(
			"%q is invalid, must be in the format: organization/project/environment/tag", id,
		)
	}
}
