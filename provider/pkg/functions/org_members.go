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
	"slices"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// GetOrganizationMembersFunction lists the members of a Pulumi Cloud
// organization. Useful for driving access-control declarations from the live
// org roster without hardcoding usernames.
type GetOrganizationMembersFunction struct{}

type GetOrganizationMembersInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

// OrganizationMemberInfo carries the fields returned by the backend member
// roster for Pulumi Cloud organizations. Display name and email are not
// included: the backing identity provider (GitHub/GitLab/Bitbucket) does not
// populate them in the member list, and Pulumi only knows those fields for
// members who have already signed in (surfaced via `knownToPulumi`).
type OrganizationMemberInfo struct {
	Username      string  `pulumi:"username"`
	Role          *string `pulumi:"role,optional"`
	RoleId        *string `pulumi:"roleId,optional"`
	RoleName      *string `pulumi:"roleName,optional"`
	KnownToPulumi bool    `pulumi:"knownToPulumi"`
	VirtualAdmin  bool    `pulumi:"virtualAdmin"`
}

type GetOrganizationMembersOutput struct {
	Members []OrganizationMemberInfo `pulumi:"members"`
}

func (GetOrganizationMembersFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetOrganizationMembersFunction{},
		"Lists all members of a Pulumi Cloud organization, including their role assignments.",
	)
	a.SetToken("index", "getOrganizationMembers")
}

func (i *GetOrganizationMembersInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
}

func (m *OrganizationMemberInfo) Annotate(a infer.Annotator) {
	a.Describe(&m.Username, "The member's Pulumi Cloud username.")
	a.Describe(
		&m.Role,
		"The member's built-in role (member, admin, billing-manager). Absent when a custom role is assigned "+
			"— check `roleId` in that case.",
	)
	a.Describe(&m.RoleId, "The custom role ID assigned to this member, if any.")
	a.Describe(&m.RoleName, "The name of the currently assigned role (custom role name, or built-in role).")
	a.Describe(&m.KnownToPulumi, "Whether this member has a Pulumi Cloud account.")
	a.Describe(
		&m.VirtualAdmin,
		"Whether this member is an admin in Pulumi Cloud without admin access on the backing identity provider.",
	)
}

func (GetOrganizationMembersFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrganizationMembersInput],
) (infer.FunctionResponse[GetOrganizationMembersOutput], error) {
	client := config.GetClient(ctx)

	members, err := client.ListOrgMembers(ctx, req.Input.OrganizationName)
	if err != nil {
		return infer.FunctionResponse[GetOrganizationMembersOutput]{}, fmt.Errorf(
			"failed to list organization members: %w",
			err,
		)
	}

	out := make([]OrganizationMemberInfo, 0, len(members.Members))
	for _, m := range members.Members {
		out = append(out, memberInfoFrom(m))
	}

	return infer.FunctionResponse[GetOrganizationMembersOutput]{
		Output: GetOrganizationMembersOutput{Members: out},
	}, nil
}

// GetOrganizationMemberFunction looks up a single Pulumi Cloud organization
// member by username. Returns an error when the member is not found.
type GetOrganizationMemberFunction struct{}

type GetOrganizationMemberInput struct {
	OrganizationName string `pulumi:"organizationName"`
	Username         string `pulumi:"username"`
}

type GetOrganizationMemberOutput struct {
	OrganizationMemberInfo
}

func (GetOrganizationMemberFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetOrganizationMemberFunction{},
		"Looks up a single member of a Pulumi Cloud organization by username (the backing identity-provider "+
			"login, e.g. GitHub login). Returns an error when the member is not found.",
	)
	a.SetToken("index", "getOrganizationMember")
}

func (i *GetOrganizationMemberInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
	a.Describe(&i.Username, "The Pulumi Cloud username (backing identity-provider login) to look up.")
}

func (GetOrganizationMemberFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrganizationMemberInput],
) (infer.FunctionResponse[GetOrganizationMemberOutput], error) {
	username := strings.TrimSpace(req.Input.Username)
	if username == "" {
		return infer.FunctionResponse[GetOrganizationMemberOutput]{}, fmt.Errorf("`username` must not be empty")
	}

	client := config.GetClient(ctx)
	member, err := client.GetOrgMember(ctx, req.Input.OrganizationName, username)
	if err != nil {
		return infer.FunctionResponse[GetOrganizationMemberOutput]{}, fmt.Errorf(
			"failed to look up organization member %q: %w",
			username,
			err,
		)
	}
	if member == nil {
		return infer.FunctionResponse[GetOrganizationMemberOutput]{}, fmt.Errorf(
			"organization %q has no member with username %q",
			req.Input.OrganizationName,
			username,
		)
	}

	return infer.FunctionResponse[GetOrganizationMemberOutput]{
		Output: GetOrganizationMemberOutput{OrganizationMemberInfo: memberInfoFrom(*member)},
	}, nil
}

// builtinOrgRoles mirrors the built-in role set in the OrganizationMember
// resource; kept in sync so the data source and resource surface the same
// shape for the same underlying member (built-in → role set, custom → roleId
// set, never both).
var builtinOrgRoles = []string{"member", "admin", "billing-manager"}

func memberInfoFrom(m pulumiapi.Member) OrganizationMemberInfo {
	info := OrganizationMemberInfo{
		Username:      m.User.GithubLogin,
		KnownToPulumi: m.KnownToPulumi,
		VirtualAdmin:  m.VirtualAdmin,
	}
	if m.FGARole != nil {
		name := m.FGARole.Name
		info.RoleName = &name
		if slices.Contains(builtinOrgRoles, m.FGARole.Name) {
			info.Role = &name
			return info
		}
		id := m.FGARole.ID
		info.RoleId = &id
		return info
	}
	role := m.Role
	info.Role = &role
	info.RoleName = &role
	return info
}
