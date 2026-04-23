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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type RoleClient interface {
	CreateRole(ctx context.Context, orgName string, req CreateRoleRequest) (*RoleDescriptor, error)
	GetRole(ctx context.Context, orgName, roleID string) (*RoleDescriptor, error)
	UpdateRole(ctx context.Context, orgName, roleID string, name, description *string, details json.RawMessage) (*RoleDescriptor, error)
	DeleteRole(ctx context.Context, orgName, roleID string, force bool) error
	ListAvailableRoleScopes(ctx context.Context, orgName string) (map[string][]RoleScopeGroup, error)
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

// RoleDescriptor mirrors the Pulumi Cloud PermissionDescriptorRecord. The
// `Details` field holds the opaque permission tree (allow/compose/condition/
// group/if-then-else/select) that the Cloud uses to evaluate permissions.
// We keep it as raw JSON so customers can supply the shape produced by the
// Cloud console without PSP having to model the entire tree.
type RoleDescriptor struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	ResourceType      string          `json:"resourceType"`
	UXPurpose         string          `json:"uxPurpose"`
	Details           json.RawMessage `json:"details,omitempty"`
	OrgID             string          `json:"orgId"`
	Version           int             `json:"version"`
	IsOrgDefault      bool            `json:"isOrgDefault"`
	DefaultIdentifier string          `json:"defaultIdentifier,omitempty"`
}

type CreateRoleRequest struct {
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	ResourceType string          `json:"resourceType"`
	UXPurpose    string          `json:"uxPurpose"`
	Details      json.RawMessage `json:"details"`
}

type updateRoleReq struct {
	Name        *string         `json:"Name"`
	Description *string         `json:"Description"`
	Details     json.RawMessage `json:"Details"`
}

// CreateRole creates a new custom role on the organization. resourceType and
// uxPurpose follow the service's PermissionDescriptorBase contract
// ("organization"/"role" is the common shape for a user-assignable fine-grained
// org role).
func (c *Client) CreateRole(
	ctx context.Context,
	orgName string,
	req CreateRoleRequest,
) (*RoleDescriptor, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name should not be empty")
	}
	if req.Name == "" {
		return nil, errors.New("role name should not be empty")
	}
	if req.ResourceType == "" {
		req.ResourceType = "organization"
	}
	if req.UXPurpose == "" {
		req.UXPurpose = "role"
	}
	if len(req.Details) == 0 {
		return nil, errors.New("role permissions details should not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles")
	var role RoleDescriptor
	if _, err := c.do(ctx, http.MethodPost, apiPath, req, &role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return &role, nil
}

// NewCreateRoleRequest is a small constructor for CreateRoleRequest. Exposed so
// that resource code can build requests without touching the unexported type.
func NewCreateRoleRequest(name, description, resourceType, uxPurpose string, details json.RawMessage) CreateRoleRequest {
	return CreateRoleRequest{
		Name:         name,
		Description:  description,
		ResourceType: resourceType,
		UXPurpose:    uxPurpose,
		Details:      details,
	}
}

// GetRole fetches a role by ID. Returns (nil, nil) if the role does not exist.
func (c *Client) GetRole(ctx context.Context, orgName, roleID string) (*RoleDescriptor, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name should not be empty")
	}
	if len(roleID) == 0 {
		return nil, errors.New("role id should not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles", roleID)
	var role RoleDescriptor
	if _, err := c.do(ctx, http.MethodGet, apiPath, nil, &role); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

// UpdateRole updates a role's name, description, and/or permission details.
// Any of the three may be nil to leave unchanged; Details as nil preserves the
// existing permission tree.
func (c *Client) UpdateRole(
	ctx context.Context,
	orgName, roleID string,
	name, description *string,
	details json.RawMessage,
) (*RoleDescriptor, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name should not be empty")
	}
	if len(roleID) == 0 {
		return nil, errors.New("role id should not be empty")
	}

	apiPath := path.Join("orgs", orgName, "roles", roleID)
	req := updateRoleReq{Name: name, Description: description, Details: details}
	var role RoleDescriptor
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
		return nil, errors.New("organization name should not be empty")
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

// DeleteRole deletes a custom role. When force is true the role is removed
// even if still assigned to members or teams.
func (c *Client) DeleteRole(ctx context.Context, orgName, roleID string, force bool) error {
	if len(orgName) == 0 {
		return errors.New("organization name should not be empty")
	}
	if len(roleID) == 0 {
		return errors.New("role id should not be empty")
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
