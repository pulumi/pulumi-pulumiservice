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

package functions

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// GetStackFunction looks up an existing stack, chiefly to obtain the
// `stackId` that role permission descriptors reference.
type GetStackFunction struct{}

type GetStackInput struct {
	OrganizationName string `pulumi:"organizationName"`
	ProjectName      string `pulumi:"projectName"`
	StackName        string `pulumi:"stackName"`
}

type GetStackOutput struct {
	OrganizationName string            `pulumi:"organizationName"`
	ProjectName      string            `pulumi:"projectName"`
	StackName        string            `pulumi:"stackName"`
	StackId          string            `pulumi:"stackId"` //nolint:revive // matches the SDK property name
	Tags             map[string]string `pulumi:"tags"`
}

func (GetStackFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetStackFunction{},
		"Looks up an existing stack in Pulumi Cloud. Use its `stackId` to scope role permissions to the "+
			"stack, for example with an `id` entity rule in `buildRolePermissions` or with "+
			"`buildStackScopedPermissions`.",
	)
	a.SetToken("index", "getStack")
}

func (i *GetStackInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&i.ProjectName, "The project name.")
	a.Describe(&i.StackName, "The stack name.")
}

func (o *GetStackOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&o.ProjectName, "The project name.")
	a.Describe(&o.StackName, "The stack name.")
	a.Describe(&o.StackId, "The stack's unique Pulumi Cloud identifier, which is distinct from the "+
		"`organization/project/stack` name.")
	a.Describe(&o.Tags, "The stack's tags.")
}

func (GetStackFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetStackInput],
) (infer.FunctionResponse[GetStackOutput], error) {
	in := req.Input
	id := pulumiapi.StackIdentifier{
		OrgName:     in.OrganizationName,
		ProjectName: in.ProjectName,
		StackName:   in.StackName,
	}
	s, err := config.GetClient(ctx).GetStack(ctx, id)
	if err != nil {
		return infer.FunctionResponse[GetStackOutput]{}, err
	}
	if s == nil {
		return infer.FunctionResponse[GetStackOutput]{}, fmt.Errorf(
			"stack %s/%s/%s not found", in.OrganizationName, in.ProjectName, in.StackName)
	}
	tags := s.Tags
	if tags == nil {
		tags = map[string]string{}
	}
	return infer.FunctionResponse[GetStackOutput]{Output: GetStackOutput{
		OrganizationName: in.OrganizationName,
		ProjectName:      in.ProjectName,
		StackName:        in.StackName,
		StackId:          s.ID,
		Tags:             tags,
	}}, nil
}
