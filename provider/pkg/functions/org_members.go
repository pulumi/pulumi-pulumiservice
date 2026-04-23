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

// GetOrganizationMembersFunction lists the members of a Pulumi Cloud
// organization. Useful for driving access-control declarations from the live
// org roster without hardcoding usernames.
type GetOrganizationMembersFunction struct{}

type GetOrganizationMembersInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

type OrganizationMemberInfo struct {
	Username      string  `pulumi:"username"`
	Name          string  `pulumi:"name"`
	Email         string  `pulumi:"email"`
	GithubLogin   string  `pulumi:"githubLogin"`
	Role          string  `pulumi:"role"`
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
	a.Describe(&m.Name, "The member's display name.")
	a.Describe(&m.Email, "The member's email address.")
	a.Describe(&m.GithubLogin, "The member's GitHub login.")
	a.Describe(&m.Role, "The member's built-in role (member, admin, billing-manager).")
	a.Describe(&m.RoleId, "The custom role ID assigned to this member, if any.")
	a.Describe(&m.RoleName, "The custom role name assigned to this member, if any.")
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
		info := OrganizationMemberInfo{
			Username:      m.User.GithubLogin,
			Name:          m.User.Name,
			Email:         m.User.Email,
			GithubLogin:   m.User.GithubLogin,
			Role:          m.Role,
			KnownToPulumi: m.KnownToPulumi,
			VirtualAdmin:  m.VirtualAdmin,
		}
		if m.FGARole != nil {
			id, name := m.FGARole.ID, m.FGARole.Name
			info.RoleId = &id
			info.RoleName = &name
		}
		out = append(out, info)
	}

	return infer.FunctionResponse[GetOrganizationMembersOutput]{
		Output: GetOrganizationMembersOutput{Members: out},
	}, nil
}
