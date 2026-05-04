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

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/apitype"
)

type RoleClient interface {
	CreateRole(
		ctx context.Context, orgName string, req apitype.PermissionDescriptorBase,
	) (*apitype.PermissionDescriptorRecord, error)
	GetRole(ctx context.Context, orgName, roleID string) (*apitype.PermissionDescriptorRecord, error)
	UpdateRole(
		ctx context.Context, orgName, roleID string,
		name, description *string, details apitype.PermissionDescriptor,
	) (*apitype.PermissionDescriptorRecord, error)
	DeleteRole(ctx context.Context, orgName, roleID string, force bool) error
	ListAvailableRoleScopes(ctx context.Context, orgName string) (map[string][]RoleScopeGroup, error)
	ListOrgRoles(ctx context.Context, orgName, uxPurpose string) ([]apitype.PermissionDescriptorRecord, error)
	ResolveBuiltInRoleID(ctx context.Context, orgName, builtInRole string) (string, error)
}

// RoleScope is a single permission scope (e.g. "stack:read") plus its
// description. Flatten these out of the grouped response for consumer code.
type RoleScope struct {
	Name        string `json:"-"`
	Description string `json:"-"`
}

// rbacScope matches the RbacScope JSON shape: {name: "...", metadata: {description: "..."}}.
// RbacScope.name is an enum on the wire but serialises as a plain string; we keep
// it as a string here because the enum drifts and we don't want customers to
// hit a hard wall when the service adds a new scope.
type rbacScope struct {
	Name     string            `json:"name"`
	Metadata rbacScopeMetadata `json:"metadata"`
}

type rbacScopeMetadata struct {
	Description string `json:"description"`
}

// RoleScopeGroup is a bucket of related scopes (e.g. "Stacks",
// "Stack deployments"). The bucketing matches what the Pulumi Cloud console
// shows when building a custom role.
type RoleScopeGroup struct {
	Name   string      `json:"-"`
	Scopes []RoleScope `json:"-"`
}

type rbacScopeGroup struct {
	Name   string      `json:"name"`
	Scopes []rbacScope `json:"scopes"`
}

// updateRoleReqWire is the legacy hand-rolled PATCH body. Stays until the
// upstream OpenAPI spec for `apitype.UpdateRoleRequest` is fixed — the
// generated type currently emits Pascal-case JSON keys (`Name`, `Description`,
// `Details`), which the API rejects. See UpdateRole below.
type updateRoleReqWire struct {
	Name        *string                    `json:"name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Details     apitype.PermissionDescriptor `json:"details,omitempty"`
}

// CreateRole creates a new permission descriptor on the organization. The
// caller chooses whether the entry is a role, policy, or other kind via
// `req.UxPurpose`; resource-layer policy on which kinds are valid is enforced
// there, not here.
func (c *Client) CreateRole(
	ctx context.Context,
	orgName string,
	req apitype.PermissionDescriptorBase,
) (*apitype.PermissionDescriptorRecord, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	if req.Name == "" {
		return nil, errors.New("role name must not be empty")
	}
	if req.Details == nil {
		return nil, errors.New("role permissions details must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles")
	var role apitype.PermissionDescriptorRecord
	if _, err := c.do(ctx, http.MethodPost, apiPath, req, &role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return &role, nil
}

// GetRole fetches a role by ID. Returns (nil, nil) if the role does not exist.
func (c *Client) GetRole(
	ctx context.Context, orgName, roleID string,
) (*apitype.PermissionDescriptorRecord, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	if len(roleID) == 0 {
		return nil, errors.New("role id must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles", roleID)
	var role apitype.PermissionDescriptorRecord
	if _, err := c.do(ctx, http.MethodGet, apiPath, nil, &role); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

// UpdateRole updates a role's name, description, and/or permission details.
// Any of the three may be nil to leave unchanged; details as nil preserves the
// existing permission tree.
//
// Transport stays on the hand-rolled PATCH because `apitype.UpdateRoleRequest`
// is currently mis-specified (Pascal-case JSON tags). Migrate when the spec
// upstream is corrected.
func (c *Client) UpdateRole(
	ctx context.Context,
	orgName, roleID string,
	name, description *string,
	details apitype.PermissionDescriptor,
) (*apitype.PermissionDescriptorRecord, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	if len(roleID) == 0 {
		return nil, errors.New("role id must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles", roleID)
	req := updateRoleReqWire{Name: name, Description: description, Details: details}
	var role apitype.PermissionDescriptorRecord
	if _, err := c.do(ctx, http.MethodPatch, apiPath, req, &role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	return &role, nil
}

// ListAvailableRoleScopes returns the permission scope catalogue, bucketed by
// resource type. The top-level map keys (e.g. "stack", "team") come from the
// service and may change; scope group names ("Stacks", "Environments", …)
// are the console labels.
func (c *Client) ListAvailableRoleScopes(
	ctx context.Context,
	orgName string,
) (map[string][]RoleScopeGroup, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles", "scopes")
	raw := map[string][]rbacScopeGroup{}
	if _, err := c.do(ctx, http.MethodGet, apiPath, nil, &raw); err != nil {
		return nil, fmt.Errorf("failed to list available role scopes: %w", err)
	}

	out := make(map[string][]RoleScopeGroup, len(raw))
	for bucket, groups := range raw {
		converted := make([]RoleScopeGroup, 0, len(groups))
		for _, g := range groups {
			scopes := make([]RoleScope, 0, len(g.Scopes))
			for _, s := range g.Scopes {
				scopes = append(scopes, RoleScope{Name: s.Name, Description: s.Metadata.Description})
			}
			converted = append(converted, RoleScopeGroup{Name: g.Name, Scopes: scopes})
		}
		out[bucket] = converted
	}
	return out, nil
}

// ListOrgRoles returns the role catalogue for an organization filtered by
// uxPurpose (e.g. "role", "policy"). The service requires uxPurpose — calling
// the endpoint without it returns 400 Bad Request.
func (c *Client) ListOrgRoles(
	ctx context.Context, orgName, uxPurpose string,
) ([]apitype.PermissionDescriptorRecord, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	if len(uxPurpose) == 0 {
		return nil, errors.New("uxPurpose must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles")
	q := url.Values{"uxPurpose": []string{uxPurpose}}
	var body struct {
		Roles []apitype.PermissionDescriptorRecord `json:"roles"`
	}
	if _, err := c.doWithQuery(ctx, http.MethodGet, apiPath, q, nil, &body); err != nil {
		return nil, fmt.Errorf("failed to list organization roles: %w", err)
	}
	return body.Roles, nil
}

// ResolveBuiltInRoleID returns the FGA role ID for a Pulumi Cloud built-in
// role (admin, member, billing-manager). The member PATCH endpoint rejects
// built-in role names in the `role` field — callers must translate to
// `fgaRoleId` first. Each org has its own set of built-in role IDs.
func (c *Client) ResolveBuiltInRoleID(ctx context.Context, orgName, builtInRole string) (string, error) {
	roles, err := c.ListOrgRoles(ctx, orgName, "role")
	if err != nil {
		return "", err
	}
	for _, r := range roles {
		if r.DefaultIdentifier == builtInRole {
			return r.ID, nil
		}
	}
	return "", fmt.Errorf("built-in role %q not found in organization %q", builtInRole, orgName)
}

// DeleteRole deletes a custom role. When force is true the role is removed
// even if still assigned to members or teams.
func (c *Client) DeleteRole(ctx context.Context, orgName, roleID string, force bool) error {
	if len(orgName) == 0 {
		return errors.New("organization name must not be empty")
	}
	if len(roleID) == 0 {
		return errors.New("role id must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles", roleID)
	q := url.Values{}
	if force {
		q.Set("force", "true")
	}
	if _, err := c.doWithQuery(ctx, http.MethodDelete, apiPath, q, nil, nil); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

