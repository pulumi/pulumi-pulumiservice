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
	Role          string   `json:"role"`
	User          User     `json:"user"`
	KnownToPulumi bool     `json:"knownToPulumi"`
	VirtualAdmin  bool     `json:"virtualAdmin"`
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
	Name        string `json:"name"`
	GithubLogin string `json:"githubLogin"`
	AvatarURL   string `json:"avatarUrl"`
	Email       string `json:"email"`
}

type addMemberToOrgReq struct {
	Role string `json:"role"`
}

type updateMemberRoleReq struct {
	Role      string  `json:"role,omitempty"`
	FGARoleID *string `json:"fgaRoleId,omitempty"`
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
//
// The underlying member-PATCH endpoint rejects built-in role names in the
// `role` field — it only accepts `fgaRoleId`. When called with a built-in
// role name, this helper resolves the FGA ID for that built-in first and
// sends the PATCH with `fgaRoleId` set.
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

	// Translate built-in role names to their per-org FGA IDs. Custom roles
	// (fgaRoleID already non-nil) pass through unchanged.
	effectiveFGA := fgaRoleID
	if effectiveFGA == nil || *effectiveFGA == "" {
		id, err := c.ResolveBuiltInRoleID(ctx, orgName, role)
		if err != nil {
			return fmt.Errorf("failed to resolve built-in role %q: %w", role, err)
		}
		effectiveFGA = &id
	}

	apiPath := path.Join("orgs", orgName, "members", userName)
	req := updateMemberRoleReq{FGARoleID: effectiveFGA}
	if _, err := c.do(ctx, http.MethodPatch, apiPath, req, nil); err != nil {
		return fmt.Errorf("failed to update org member role: %w", err)
	}
	return nil
}

// ListOrgMembers returns every visible member of the organization by
// merging the two rosters Pulumi Cloud exposes and deduping by username:
//
//   - `?type=backend`: identity-provider roster
//     (GitHub/GitLab/Bitbucket/SAML; paginated; includes users who haven't
//     logged into Pulumi yet, with KnownToPulumi=false).
//   - `?type=frontend`: Pulumi DB seat-count roster (no pagination, ≤50).
//
// Either roster alone can miss members for orgs whose IdP and DB
// memberships have drifted (commonly: SAML-provisioned orgs with legacy
// non-SAML accounts that the IdP no longer enumerates). Calling both is
// the closest the API exposes to "every member regardless of provenance".
//
// Backend wins on dedup conflict (it carries the IdP-of-record fields
// like KnownToPulumi). A frontend failure is non-fatal: the backend
// result is returned with a logged warning so a transient frontend
// error doesn't gate the more-inclusive backend path.
func (c *Client) ListOrgMembers(ctx context.Context, orgName string) (*Members, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	backend, err := c.listOrgMembersByType(ctx, orgName, "backend")
	if err != nil {
		return nil, fmt.Errorf("failed to list organization members: %w", err)
	}
	frontend, frontendErr := c.listOrgMembersByType(ctx, orgName, "frontend")
	if frontendErr != nil {
		// Non-fatal: backend is the more-inclusive roster, return it
		// alone rather than failing the whole call.
		frontend = nil
	}

	seen := make(map[string]bool, len(backend))
	out := make([]Member, 0, len(backend)+len(frontend))
	add := func(m Member) {
		key := memberKey(m)
		// Skip the synthetic org-as-user entry the backend roster emits
		// for some orgs (key matching orgName, epoch-zero `created`,
		// empty fgaRole). Not a real member; should never appear in
		// `getOrganizationMembers` output.
		if key == orgName {
			return
		}
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, m)
	}
	for _, m := range backend {
		add(m)
	}
	for _, m := range frontend {
		add(m)
	}

	return &Members{Members: out}, nil
}

// memberKey is the dedup key for merging the two rosters. Prefer the
// canonical Pulumi username; fall back to GithubLogin for IdP-only users
// who haven't signed in yet (Name can be empty for them).
func memberKey(m Member) string {
	if m.User.Name != "" {
		return m.User.Name
	}
	return m.User.GithubLogin
}

func (c *Client) listOrgMembersByType(
	ctx context.Context, orgName, rosterType string,
) ([]Member, error) {
	apiPath := path.Join("orgs", orgName, "members")
	var all []Member
	token := ""
	for {
		q := url.Values{"type": []string{rosterType}}
		if token != "" {
			q.Set("continuationToken", token)
		}
		var page Members
		if _, err := c.doWithQuery(ctx, http.MethodGet, apiPath, q, nil, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Members...)
		if page.ContinuationToken == nil || *page.ContinuationToken == "" {
			break
		}
		token = *page.ContinuationToken
	}
	return all, nil
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
