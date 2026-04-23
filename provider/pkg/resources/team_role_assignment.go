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

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

type TeamRoleAssignment struct{}

var (
	_ infer.CustomCreate[TeamRoleAssignmentInput, TeamRoleAssignmentState] = &TeamRoleAssignment{}
	_ infer.CustomDelete[TeamRoleAssignmentState]                          = &TeamRoleAssignment{}
	_ infer.CustomRead[TeamRoleAssignmentInput, TeamRoleAssignmentState]   = &TeamRoleAssignment{}
)

func (*TeamRoleAssignment) Annotate(a infer.Annotator) {
	a.Describe(
		&TeamRoleAssignment{},
		"Assigns a custom (fine-grained) role to a Pulumi Cloud team. The Pulumi Cloud API currently "+
			"supports one role per team; creating a second assignment replaces the first. Automatically "+
			"enables the team's custom-roles feature on first use.",
	)
}

type TeamRoleAssignmentInput struct {
	OrganizationName string `pulumi:"organizationName" provider:"replaceOnChanges"`
	TeamName         string `pulumi:"teamName"         provider:"replaceOnChanges"`
	RoleId           string `pulumi:"roleId"           provider:"replaceOnChanges"`
}

func (i *TeamRoleAssignmentInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&i.TeamName, "The team name.")
	a.Describe(&i.RoleId, "The ID of the custom role to assign.")
}

type TeamRoleAssignmentState struct {
	TeamRoleAssignmentInput
	RoleName string `pulumi:"roleName"`
}

func (s *TeamRoleAssignmentState) Annotate(a infer.Annotator) {
	a.Describe(&s.RoleName, "The name of the assigned role at the time of last refresh.")
}

func (*TeamRoleAssignment) Create(
	ctx context.Context,
	req infer.CreateRequest[TeamRoleAssignmentInput],
) (infer.CreateResponse[TeamRoleAssignmentState], error) {
	id := teamRoleAssignmentID(req.Inputs.OrganizationName, req.Inputs.TeamName, req.Inputs.RoleId)
	if req.DryRun {
		return infer.CreateResponse[TeamRoleAssignmentState]{
			ID:     id,
			Output: TeamRoleAssignmentState{TeamRoleAssignmentInput: req.Inputs},
		}, nil
	}

	client := config.GetClient(ctx)
	if err := client.AssignRoleToTeam(
		ctx,
		req.Inputs.OrganizationName,
		req.Inputs.TeamName,
		req.Inputs.RoleId,
	); err != nil {
		return infer.CreateResponse[TeamRoleAssignmentState]{}, err
	}

	ref, err := client.GetTeamRole(ctx, req.Inputs.OrganizationName, req.Inputs.TeamName, req.Inputs.RoleId)
	if err != nil {
		return infer.CreateResponse[TeamRoleAssignmentState]{
			ID:     id,
			Output: TeamRoleAssignmentState{TeamRoleAssignmentInput: req.Inputs},
		}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
	}
	out := TeamRoleAssignmentState{TeamRoleAssignmentInput: req.Inputs}
	if ref != nil {
		out.RoleName = ref.Name
	}
	return infer.CreateResponse[TeamRoleAssignmentState]{ID: id, Output: out}, nil
}

func (*TeamRoleAssignment) Delete(
	ctx context.Context,
	req infer.DeleteRequest[TeamRoleAssignmentState],
) (infer.DeleteResponse, error) {
	client := config.GetClient(ctx)
	return infer.DeleteResponse{}, client.RemoveRoleFromTeam(
		ctx,
		req.State.OrganizationName,
		req.State.TeamName,
		req.State.RoleId,
	)
}

func (*TeamRoleAssignment) Read(
	ctx context.Context,
	req infer.ReadRequest[TeamRoleAssignmentInput, TeamRoleAssignmentState],
) (infer.ReadResponse[TeamRoleAssignmentInput, TeamRoleAssignmentState], error) {
	orgName, teamName, roleID, err := splitTeamRoleAssignmentID(req.ID)
	if err != nil {
		return infer.ReadResponse[TeamRoleAssignmentInput, TeamRoleAssignmentState]{}, err
	}

	client := config.GetClient(ctx)
	ref, err := client.GetTeamRole(ctx, orgName, teamName, roleID)
	if err != nil {
		return infer.ReadResponse[TeamRoleAssignmentInput, TeamRoleAssignmentState]{}, fmt.Errorf(
			"failed to read team-role assignment (%q): %w",
			req.ID,
			err,
		)
	}
	if ref == nil {
		return infer.ReadResponse[TeamRoleAssignmentInput, TeamRoleAssignmentState]{}, nil
	}

	in := TeamRoleAssignmentInput{
		OrganizationName: orgName,
		TeamName:         teamName,
		RoleId:           roleID,
	}
	return infer.ReadResponse[TeamRoleAssignmentInput, TeamRoleAssignmentState]{
		ID:     req.ID,
		Inputs: in,
		State: TeamRoleAssignmentState{
			TeamRoleAssignmentInput: in,
			RoleName:                ref.Name,
		},
	}, nil
}

func teamRoleAssignmentID(org, team, role string) string {
	return fmt.Sprintf("%s/%s/%s", org, team, role)
}

func splitTeamRoleAssignmentID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf(
			"%q is invalid, must be in the format: organization/team/roleId",
			id,
		)
	}
	return parts[0], parts[1], parts[2], nil
}
