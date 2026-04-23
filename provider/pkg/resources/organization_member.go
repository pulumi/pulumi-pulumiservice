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
	"net/http"
	"slices"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type OrganizationMember struct{}

var (
	_ infer.CustomCreate[OrganizationMemberInput, OrganizationMemberState] = &OrganizationMember{}
	_ infer.CustomCheck[OrganizationMemberInput]                           = &OrganizationMember{}
	_ infer.CustomDelete[OrganizationMemberState]                          = &OrganizationMember{}
	_ infer.CustomRead[OrganizationMemberInput, OrganizationMemberState]   = &OrganizationMember{}
	_ infer.CustomUpdate[OrganizationMemberInput, OrganizationMemberState] = &OrganizationMember{}
)

func (*OrganizationMember) Annotate(a infer.Annotator) {
	a.Describe(
		&OrganizationMember{},
		"Manages a user's membership in a Pulumi Cloud organization and their assigned role. The user must "+
			"already have a Pulumi Cloud account before they can be added. Custom (fine-grained) roles are "+
			"assigned by setting `roleId`; built-in roles are assigned by setting `role`. When both are set, "+
			"`roleId` takes precedence.",
	)
}

type OrganizationMemberCore struct {
	OrganizationName string  `pulumi:"organizationName" provider:"replaceOnChanges"`
	Username         string  `pulumi:"username"         provider:"replaceOnChanges"`
	Role             *string `pulumi:"role,optional"`
	RoleId           *string `pulumi:"roleId,optional"`
}

func (c *OrganizationMemberCore) Annotate(a infer.Annotator) {
	a.Describe(&c.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&c.Username, "The Pulumi Cloud username of the member.")
	a.Describe(
		&c.Role,
		"The built-in organization role. One of `member`, `admin`, `billing-manager`. "+
			"Defaults to `member` on create. Ignored when `roleId` is set.",
	)
	a.Describe(&c.RoleId, "The ID of a custom (fine-grained) organization role to assign. Takes precedence over `role`.")
}

type OrganizationMemberInput struct {
	OrganizationMemberCore
}

type OrganizationMemberState struct {
	OrganizationMemberCore
	Name          string `pulumi:"name"`
	Email         string `pulumi:"email"`
	GithubLogin   string `pulumi:"githubLogin"`
	KnownToPulumi bool   `pulumi:"knownToPulumi"`
	RoleName      string `pulumi:"roleName"`
}

func (s *OrganizationMemberState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "The member's display name.")
	a.Describe(&s.Email, "The member's email address.")
	a.Describe(&s.GithubLogin, "The member's GitHub login.")
	a.Describe(&s.KnownToPulumi, "Whether the member has a Pulumi Cloud account.")
	a.Describe(&s.RoleName, "The name of the currently assigned role (custom role name, or built-in role).")
}

const defaultOrgMemberRole = "member"

var validOrgMemberRoles = []string{defaultOrgMemberRole, "admin", "billing-manager"}

func (*OrganizationMember) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[OrganizationMemberInput], error) {
	in, failures, err := infer.DefaultCheck[OrganizationMemberInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[OrganizationMemberInput]{}, err
	}
	if in.Role != nil && *in.Role != "" && !slices.Contains(validOrgMemberRoles, *in.Role) {
		failures = append(failures, p.CheckFailure{
			Property: "role",
			Reason:   fmt.Sprintf("role must be one of %v, got %q", validOrgMemberRoles, *in.Role),
		})
	}
	return infer.CheckResponse[OrganizationMemberInput]{Inputs: in, Failures: failures}, nil
}

func (*OrganizationMember) Create(
	ctx context.Context,
	req infer.CreateRequest[OrganizationMemberInput],
) (infer.CreateResponse[OrganizationMemberState], error) {
	id := fmt.Sprintf("%s/%s", req.Inputs.OrganizationName, req.Inputs.Username)

	core := req.Inputs.OrganizationMemberCore
	if core.Role == nil || *core.Role == "" {
		def := defaultOrgMemberRole
		core.Role = &def
	}

	if req.DryRun {
		return infer.CreateResponse[OrganizationMemberState]{
			ID:     id,
			Output: OrganizationMemberState{OrganizationMemberCore: core},
		}, nil
	}

	client := config.GetClient(ctx)

	// The add endpoint only takes built-in roles; custom roles are applied via
	// an update that follows the add. We always add with a built-in role first.
	addRole := util.OrZero(core.Role)
	if addRole == "" {
		addRole = defaultOrgMemberRole
	}
	// Adopt an existing membership: if the user is already in the org (409),
	// fall through and apply the desired role via UpdateOrgMemberRole below.
	// This makes Create behave as "ensure user is in org with role X".
	err := client.AddMemberToOrg(ctx, req.Inputs.Username, req.Inputs.OrganizationName, addRole)
	if err != nil && pulumiapi.GetErrorStatusCode(err) != http.StatusConflict {
		return infer.CreateResponse[OrganizationMemberState]{}, fmt.Errorf(
			"error adding user %q to org %q: %w",
			req.Inputs.Username,
			req.Inputs.OrganizationName,
			err,
		)
	}
	existingMember := err != nil // 409 path

	// Always patch the role when we adopted an existing member (the built-in
	// role from Add wasn't applied) or when a custom role was requested.
	needsRolePatch := existingMember || (core.RoleId != nil && *core.RoleId != "")
	if needsRolePatch {
		roleToSet := ""
		if core.RoleId == nil || *core.RoleId == "" {
			roleToSet = addRole
		}
		if err := client.UpdateOrgMemberRole(
			ctx,
			req.Inputs.OrganizationName,
			req.Inputs.Username,
			roleToSet,
			core.RoleId,
		); err != nil {
			return infer.CreateResponse[OrganizationMemberState]{
					ID: id,
					Output: OrganizationMemberState{
						OrganizationMemberCore: core,
					},
				}, infer.ResourceInitFailedError{Reasons: []string{
					fmt.Sprintf("user added but failed to assign role: %s", err.Error()),
				}}
		}
	}

	state, err := readOrgMemberState(ctx, req.Inputs.OrganizationName, req.Inputs.Username, core)
	if err != nil {
		return infer.CreateResponse[OrganizationMemberState]{
			ID:     id,
			Output: OrganizationMemberState{OrganizationMemberCore: core},
		}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
	}

	return infer.CreateResponse[OrganizationMemberState]{ID: id, Output: state}, nil
}

func (*OrganizationMember) Update(
	ctx context.Context,
	req infer.UpdateRequest[OrganizationMemberInput, OrganizationMemberState],
) (infer.UpdateResponse[OrganizationMemberState], error) {
	core := req.Inputs.OrganizationMemberCore

	if req.DryRun {
		return infer.UpdateResponse[OrganizationMemberState]{
			Output: OrganizationMemberState{OrganizationMemberCore: core},
		}, nil
	}

	client := config.GetClient(ctx)

	// Always call UpdateOrgMemberRole: the service treats role vs fgaRoleId
	// precedence correctly and this is simpler than diffing.
	role := ""
	if core.RoleId == nil || *core.RoleId == "" {
		role = util.OrZero(core.Role)
		if role == "" {
			role = defaultOrgMemberRole
		}
	}
	if err := client.UpdateOrgMemberRole(
		ctx,
		req.State.OrganizationName,
		req.State.Username,
		role,
		core.RoleId,
	); err != nil {
		return infer.UpdateResponse[OrganizationMemberState]{}, fmt.Errorf(
			"error updating org member role: %w",
			err,
		)
	}

	state, err := readOrgMemberState(ctx, req.State.OrganizationName, req.State.Username, core)
	if err != nil {
		return infer.UpdateResponse[OrganizationMemberState]{}, err
	}
	return infer.UpdateResponse[OrganizationMemberState]{Output: state}, nil
}

func (*OrganizationMember) Delete(
	ctx context.Context,
	req infer.DeleteRequest[OrganizationMemberState],
) (infer.DeleteResponse, error) {
	client := config.GetClient(ctx)
	return infer.DeleteResponse{}, client.DeleteMemberFromOrg(
		ctx,
		req.State.OrganizationName,
		req.State.Username,
	)
}

func (*OrganizationMember) Read(
	ctx context.Context,
	req infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState],
) (infer.ReadResponse[OrganizationMemberInput, OrganizationMemberState], error) {
	orgName, userName, err := splitOrgMemberID(req.ID)
	if err != nil {
		return infer.ReadResponse[OrganizationMemberInput, OrganizationMemberState]{}, err
	}

	core := req.Inputs.OrganizationMemberCore
	core.OrganizationName = orgName
	core.Username = userName

	state, err := readOrgMemberState(ctx, orgName, userName, core)
	if err != nil {
		return infer.ReadResponse[OrganizationMemberInput, OrganizationMemberState]{}, err
	}
	if state.Username == "" {
		// Member not found.
		return infer.ReadResponse[OrganizationMemberInput, OrganizationMemberState]{}, nil
	}

	return infer.ReadResponse[OrganizationMemberInput, OrganizationMemberState]{
		ID:     req.ID,
		Inputs: OrganizationMemberInput{OrganizationMemberCore: state.OrganizationMemberCore},
		State:  state,
	}, nil
}

func readOrgMemberState(
	ctx context.Context,
	orgName, userName string,
	core OrganizationMemberCore,
) (OrganizationMemberState, error) {
	client := config.GetClient(ctx)
	member, err := client.GetOrgMember(ctx, orgName, userName)
	if err != nil {
		return OrganizationMemberState{}, fmt.Errorf("failed to read org member: %w", err)
	}
	if member == nil {
		return OrganizationMemberState{}, nil
	}

	state := OrganizationMemberState{
		OrganizationMemberCore: OrganizationMemberCore{
			OrganizationName: orgName,
			Username:         userName,
			Role:             core.Role,
			RoleId:           core.RoleId,
		},
		Name:          member.User.Name,
		Email:         member.User.Email,
		GithubLogin:   member.User.GithubLogin,
		KnownToPulumi: member.KnownToPulumi,
	}

	applyMemberRoleToState(member, &state)
	return state, nil
}

// applyMemberRoleToState populates Role/RoleId/RoleName on state from the
// service response. fgaRole is authoritative: if it names a custom role
// (anything outside the built-in set) we surface roleId; otherwise we surface
// the built-in role.
func applyMemberRoleToState(member *pulumiapi.Member, state *OrganizationMemberState) {
	if member.FGARole != nil {
		state.RoleName = member.FGARole.Name
		if slices.Contains(validOrgMemberRoles, member.FGARole.Name) {
			name := member.FGARole.Name
			state.Role = &name
			state.RoleId = nil
			return
		}
		id := member.FGARole.ID
		state.RoleId = &id
		state.Role = nil
		return
	}
	name := member.Role
	state.Role = &name
	state.RoleName = name
}

func splitOrgMemberID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%q is invalid, must be in the format: organization/username", id)
	}
	return parts[0], parts[1], nil
}
