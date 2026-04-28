// Copyright 2016-2026, Pulumi Corporation.
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

package functions

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// GetEnvironmentFunction looks up an existing ESC environment by
// (organization, project, name) and returns its UUID. Useful when the
// environment is managed outside Pulumi (or in a different stack) and the
// caller needs its `environmentId` to pin a custom RBAC role to it via
// a `literalEnvironment` expression's `identity` field.
type GetEnvironmentFunction struct{}

type GetEnvironmentInput struct {
	OrganizationName string `pulumi:"organizationName"`
	ProjectName      string `pulumi:"projectName,optional"`
	Name             string `pulumi:"name"`
}

type GetEnvironmentOutput struct {
	OrganizationName string `pulumi:"organizationName"`
	ProjectName      string `pulumi:"projectName"`
	Name             string `pulumi:"name"`
	EnvironmentID    string `pulumi:"environmentId"`
}

func (GetEnvironmentFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetEnvironmentFunction{},
		"Looks up an existing ESC environment by name and returns its UUID. Use this to pin a custom "+
			"RBAC role to a specific environment via a `literalEnvironment` expression when the "+
			"environment is not managed by the current Pulumi program. Errors when the environment is not found.",
	)
	a.SetToken("index", "getEnvironment")
}

func (i *GetEnvironmentInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization that owns the environment.")
	a.Describe(&i.ProjectName, "The ESC project name. Defaults to `default`.")
	a.Describe(&i.Name, "The environment name.")
}

func (o *GetEnvironmentOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.OrganizationName, "The Pulumi Cloud organization that owns the environment.")
	a.Describe(&o.ProjectName, "The ESC project the environment lives in.")
	a.Describe(&o.Name, "The environment name.")
	a.Describe(
		&o.EnvironmentID,
		"The environment's UUID. Use this as the `identity` value when pinning a custom RBAC role to this "+
			"environment via a `literalEnvironment` expression.",
	)
}

func (GetEnvironmentFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetEnvironmentInput],
) (infer.FunctionResponse[GetEnvironmentOutput], error) {
	in := req.Input
	if in.OrganizationName == "" {
		return infer.FunctionResponse[GetEnvironmentOutput]{}, fmt.Errorf("`organizationName` must not be empty")
	}
	if in.Name == "" {
		return infer.FunctionResponse[GetEnvironmentOutput]{}, fmt.Errorf("`name` must not be empty")
	}
	project := in.ProjectName
	if project == "" {
		project = "default"
	}

	meta, err := config.GetClient(ctx).GetEnvironmentMetadata(ctx, in.OrganizationName, project, in.Name)
	if err != nil {
		return infer.FunctionResponse[GetEnvironmentOutput]{}, fmt.Errorf(
			"failed to look up environment %s/%s/%s: %w", in.OrganizationName, project, in.Name, err,
		)
	}
	if meta == nil {
		return infer.FunctionResponse[GetEnvironmentOutput]{}, fmt.Errorf(
			"environment %s/%s/%s not found", in.OrganizationName, project, in.Name,
		)
	}

	return infer.FunctionResponse[GetEnvironmentOutput]{
		Output: GetEnvironmentOutput{
			OrganizationName: in.OrganizationName,
			ProjectName:      project,
			Name:             in.Name,
			EnvironmentID:    meta.ID,
		},
	}, nil
}
