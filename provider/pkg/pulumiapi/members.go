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
package pulumiapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type MemberClient interface {
	AddMemberToOrg(ctx context.Context, userName, orgName, role string) error
	UpdateOrgMemberRole(ctx context.Context, orgName, userName, role string, fgaRoleID *string) error
	ListOrgMembers(ctx context.Context, orgName string) (*Members, error)
	GetOrgMember(ctx context.Context, orgName, userName string) (*Member, error)
	DeleteMemberFromOrg(ctx context.Context, orgName, userName string) error
}

type Members struct {
	Members           []Member
	ContinuationToken *string `json:"continuationToken,omitempty"`
}

type Member struct {
	Role          string
	User          User
	KnownToPulumi bool
	VirtualAdmin  bool
	FGARole       *FGARole `json:"fgaRole,omitempty"`
}

// FGARole is the fine-grained role assigned to a member. For built-in roles,
// the ID/name reflect member/admin/billing-manager; for custom roles, they
// reflect the custom role defined on the organization.
type FGARole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Name        string
	GithubLogin string
	AvatarURL   string
	Email       string
}

type addMemberToOrgReq struct {
	Role string `json:"role"`
}

type updateMemberRoleReq struct {
	Role       string  `json:"role,omitempty"`
	FGARoleID  *string `json:"fgaRoleId,omitempty"`
}

func validBuiltinOrgRole(role string) bool {
	switch role {
	case "admin", "member", "billing-manager":
		return true
	}
	return false
}

func (c *Client) AddMemberToOrg(ctx context.Context, userName string, orgName string, role string) error {
	if len(userName) == 0 {
		return errors.New("username should not be empty")
	}
	if len(orgName) == 0 {
		return errors.New("organization name should not be empty")
	}
	if !validBuiltinOrgRole(role) {
		return fmt.Errorf("role must be one of: admin, member, billing-manager")
	}

	apiPath := path.Join("orgs", orgName, "members", userName)

	req := addMemberToOrgReq{
		Role: role,
	}
	_, err := c.do(ctx, http.MethodPost, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to add member to org: %w", err)
	}
	return nil
}

// UpdateOrgMemberRole updates a member's role on the organization. Pass role
// for a built-in role (member, admin, billing-manager); pass fgaRoleID to
// assign a custom role (takes precedence). At least one must be non-empty.
func (c *Client) UpdateOrgMemberRole(
	ctx context.Context,
	orgName, userName, role string,
	fgaRoleID *string,
) error {
	if len(orgName) == 0 {
		return errors.New("organization name should not be empty")
	}
	if len(userName) == 0 {
		return errors.New("username should not be empty")
	}
	if role == "" && (fgaRoleID == nil || *fgaRoleID == "") {
		return errors.New("one of role or fgaRoleID must be set")
	}
	if role != "" && !validBuiltinOrgRole(role) {
		return fmt.Errorf("role must be one of: admin, member, billing-manager")
	}

	apiPath := path.Join("orgs", orgName, "members", userName)
	req := updateMemberRoleReq{Role: role, FGARoleID: fgaRoleID}
	if _, err := c.do(ctx, http.MethodPatch, apiPath, req, nil); err != nil {
		return fmt.Errorf("failed to update org member role: %w", err)
	}
	return nil
}

// ListOrgMembers returns every member of the organization, following
// continuationToken pagination.
func (c *Client) ListOrgMembers(ctx context.Context, orgName string) (*Members, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	apiPath := path.Join("orgs", orgName, "members")
	all := Members{Members: []Member{}}

	token := ""
	for {
		q := url.Values{"type": []string{"backend"}}
		if token != "" {
			q.Set("continuationToken", token)
		}

		var page Members
		if _, err := c.doWithQuery(ctx, http.MethodGet, apiPath, q, nil, &page); err != nil {
			return nil, fmt.Errorf("failed to list organization members: %w", err)
		}
		all.Members = append(all.Members, page.Members...)

		if page.ContinuationToken == nil || *page.ContinuationToken == "" {
			break
		}
		token = *page.ContinuationToken
	}

	return &all, nil
}

// GetOrgMember looks up a single member by username using the list endpoint.
// Returns (nil, nil) when not found.
func (c *Client) GetOrgMember(ctx context.Context, orgName, userName string) (*Member, error) {
	members, err := c.ListOrgMembers(ctx, orgName)
	if err != nil {
		return nil, err
	}
	for i, m := range members.Members {
		if m.User.GithubLogin == userName || m.User.Name == userName {
			return &members.Members[i], nil
		}
	}
	return nil, nil
}

func (c *Client) DeleteMemberFromOrg(ctx context.Context, orgName string, userName string) error {
	if len(orgName) == 0 {
		return errors.New("orgName must not be empty")
	}

	if len(userName) == 0 {
		return errors.New("userName must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "members", userName)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)

	if err != nil {
		return fmt.Errorf("failed to delete member from org: %w", err)
	}
	return nil
}
