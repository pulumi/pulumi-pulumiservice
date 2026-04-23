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
	"errors"
	"fmt"
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
		out = append(out, memberInfoFrom(m))
	}

	return infer.FunctionResponse[GetOrganizationMembersOutput]{
		Output: GetOrganizationMembersOutput{Members: out},
	}, nil
}

// GetOrganizationMemberFunction looks up a single Pulumi Cloud organization
// member by username or email. Exactly one of the two lookup fields must be
// provided. Returns an error when the member is not found.
type GetOrganizationMemberFunction struct{}

type GetOrganizationMemberInput struct {
	OrganizationName string  `pulumi:"organizationName"`
	Username         *string `pulumi:"username,optional"`
	Email            *string `pulumi:"email,optional"`
}

type GetOrganizationMemberOutput struct {
	OrganizationMemberInfo
}

func (GetOrganizationMemberFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetOrganizationMemberFunction{},
		"Looks up a single member of a Pulumi Cloud organization by username or "+
			"email. Exactly one of `username` or `email` must be set. Returns an "+
			"error when the member is not found.",
	)
	a.SetToken("index", "getOrganizationMember")
}

func (i *GetOrganizationMemberInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
	a.Describe(&i.Username, "The Pulumi Cloud username to look up. Mutually exclusive with `email`.")
	a.Describe(&i.Email, "The email address to look up. Matching is case-insensitive. Mutually exclusive with `username`.")
}

func (GetOrganizationMemberFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrganizationMemberInput],
) (infer.FunctionResponse[GetOrganizationMemberOutput], error) {
	username := strings.TrimSpace(deref(req.Input.Username))
	email := strings.TrimSpace(deref(req.Input.Email))

	switch {
	case username == "" && email == "":
		return infer.FunctionResponse[GetOrganizationMemberOutput]{}, errors.New(
			"exactly one of `username` or `email` must be set",
		)
	case username != "" && email != "":
		return infer.FunctionResponse[GetOrganizationMemberOutput]{}, errors.New(
			"`username` and `email` are mutually exclusive; set only one",
		)
	}

	client := config.GetClient(ctx)

	var (
		member *pulumiapi.Member
		err    error
		lookup string
	)
	if username != "" {
		lookup = fmt.Sprintf("username %q", username)
		member, err = client.GetOrgMember(ctx, req.Input.OrganizationName, username)
	} else {
		lookup = fmt.Sprintf("email %q", email)
		member, err = client.GetOrgMemberByEmail(ctx, req.Input.OrganizationName, email)
	}
	if err != nil {
		return infer.FunctionResponse[GetOrganizationMemberOutput]{}, fmt.Errorf(
			"failed to look up organization member by %s: %w",
			lookup,
			err,
		)
	}
	if member == nil {
		return infer.FunctionResponse[GetOrganizationMemberOutput]{}, fmt.Errorf(
			"organization %q has no member matching %s",
			req.Input.OrganizationName,
			lookup,
		)
	}

	return infer.FunctionResponse[GetOrganizationMemberOutput]{
		Output: GetOrganizationMemberOutput{OrganizationMemberInfo: memberInfoFrom(*member)},
	}, nil
}

func memberInfoFrom(m pulumiapi.Member) OrganizationMemberInfo {
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
	return info
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
