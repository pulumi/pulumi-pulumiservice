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
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type TeamStackPermissionScope int

const (
	TeamStackPermissionRead  TeamStackPermissionScope = 101
	TeamStackPermissionEdit  TeamStackPermissionScope = 102
	TeamStackPermissionAdmin TeamStackPermissionScope = 103
)

func (TeamStackPermissionScope) Values() []infer.EnumValue[TeamStackPermissionScope] {
	return []infer.EnumValue[TeamStackPermissionScope]{
		{Name: "read", Value: TeamStackPermissionRead, Description: "Grants read permissions to stack."},
		{Name: "edit", Value: TeamStackPermissionEdit, Description: "Grants edit permissions to stack."},
		{Name: gcAdmin, Value: TeamStackPermissionAdmin, Description: "Grants admin permissions to stack."},
	}
}

type TeamStackPermission struct{}

var (
	_ infer.CustomCreate[TeamStackPermissionInput, TeamStackPermissionState] = &TeamStackPermission{}
	_ infer.CustomDelete[TeamStackPermissionState]                           = &TeamStackPermission{}
	_ infer.CustomRead[TeamStackPermissionInput, TeamStackPermissionState]   = &TeamStackPermission{}
)

func (*TeamStackPermission) Annotate(a infer.Annotator) {
	a.Describe(&TeamStackPermission{}, "Grants a team permissions to the specified stack.")
	a.SetToken("index", "TeamStackPermission")
}

type TeamStackPermissionInput struct {
	Organization string                   `pulumi:"organization" provider:"replaceOnChanges"`
	Project      string                   `pulumi:"project"      provider:"replaceOnChanges"`
	Stack        string                   `pulumi:"stack"        provider:"replaceOnChanges"`
	Team         string                   `pulumi:"team"         provider:"replaceOnChanges"`
	Permission   TeamStackPermissionScope `pulumi:"permission"   provider:"replaceOnChanges"`
}

func (i *TeamStackPermissionInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "The organization or the personal account name of the stack.")
	a.Describe(&i.Project, "The project name for this stack.")
	a.Describe(&i.Stack, "The name of the stack that the team will be granted permissions to.")
	a.Describe(&i.Team, "The name of the team to grant this stack permissions to. This is not the display name.")
	a.Describe(&i.Permission, "Sets the permission level that this team will be granted to the stack.")
}

type TeamStackPermissionState struct {
	TeamStackPermissionInput
}

func (*TeamStackPermission) Create(
	ctx context.Context,
	req infer.CreateRequest[TeamStackPermissionInput],
) (infer.CreateResponse[TeamStackPermissionState], error) {
	if req.DryRun {
		return infer.CreateResponse[TeamStackPermissionState]{
			Output: TeamStackPermissionState{TeamStackPermissionInput: req.Inputs},
		}, nil
	}
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.Inputs.Organization,
		ProjectName: req.Inputs.Project,
		StackName:   req.Inputs.Stack,
	}
	err := config.GetClient(ctx).AddStackPermission(ctx, stack, req.Inputs.Team, int(req.Inputs.Permission))
	if err != nil {
		return infer.CreateResponse[TeamStackPermissionState]{}, fmt.Errorf(
			"error granting team stack permission: %w", err,
		)
	}
	return infer.CreateResponse[TeamStackPermissionState]{
		ID:     teamStackPermissionResourceID(stack, req.Inputs.Team),
		Output: TeamStackPermissionState{TeamStackPermissionInput: req.Inputs},
	}, nil
}

func (*TeamStackPermission) Delete(
	ctx context.Context,
	req infer.DeleteRequest[TeamStackPermissionState],
) (infer.DeleteResponse, error) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.State.Organization,
		ProjectName: req.State.Project,
		StackName:   req.State.Stack,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).RemoveStackPermission(ctx, stack, req.State.Team)
}

func (*TeamStackPermission) Read(
	ctx context.Context,
	req infer.ReadRequest[TeamStackPermissionInput, TeamStackPermissionState],
) (infer.ReadResponse[TeamStackPermissionInput, TeamStackPermissionState], error) {
	permID, err := splitTeamStackPermissionID(req.ID)
	if err != nil {
		if strings.Contains(err.Error(), "expected 4 parts") {
			return infer.ReadResponse[TeamStackPermissionInput, TeamStackPermissionState]{}, fmt.Errorf(
				"TeamStackPermission resources created before v0.17.0 do not support refresh. " +
					"You will need to destroy and recreate this resource with >v0.17.0 to successfully refresh",
			)
		}
		return infer.ReadResponse[TeamStackPermissionInput, TeamStackPermissionState]{}, err
	}

	stack := pulumiapi.StackIdentifier{
		OrgName:     permID.Organization,
		ProjectName: permID.Project,
		StackName:   permID.Stack,
	}
	permission, err := config.GetClient(ctx).GetTeamStackPermission(ctx, stack, permID.Team)
	if err != nil {
		return infer.ReadResponse[TeamStackPermissionInput, TeamStackPermissionState]{}, fmt.Errorf(
			"failed to get team stack permission: %w", err,
		)
	}
	if permission == nil {
		return infer.ReadResponse[TeamStackPermissionInput, TeamStackPermissionState]{}, nil
	}

	inputs := TeamStackPermissionInput{
		Organization: permID.Organization,
		Project:      permID.Project,
		Stack:        permID.Stack,
		Team:         permID.Team,
		Permission:   TeamStackPermissionScope(*permission),
	}
	return infer.ReadResponse[TeamStackPermissionInput, TeamStackPermissionState]{
		ID:     req.ID,
		Inputs: inputs,
		State:  TeamStackPermissionState{TeamStackPermissionInput: inputs},
	}, nil
}

func teamStackPermissionResourceID(stack pulumiapi.StackIdentifier, team string) string {
	return fmt.Sprintf("%s/%s/%s/%s", stack.OrgName, stack.ProjectName, stack.StackName, team)
}

type teamStackPermissionID struct {
	Organization string
	Project      string
	Stack        string
	Team         string
}

func splitTeamStackPermissionID(id string) (teamStackPermissionID, error) {
	split := strings.Split(id, "/")
	if len(split) != 4 {
		return teamStackPermissionID{}, fmt.Errorf("invalid id %q, expected 4 parts", id)
	}
	return teamStackPermissionID{
		Organization: split[0],
		Project:      split[1],
		Stack:        split[2],
		Team:         split[3],
	}, nil
}
