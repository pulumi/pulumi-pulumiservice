// Copyright 2016-2025, Pulumi Corporation.
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
	"fmt"
	"net/http"
	"path"
)

type RoleClient interface {
	CreateRole(ctx context.Context, orgName string, req CreateRoleRequest) (*Role, error)
	GetRole(ctx context.Context, orgName string, roleID string) (*Role, error)
	UpdateRole(ctx context.Context, orgName string, roleID string, req UpdateRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, orgName string, roleID string) error
}

// Role represents a role response from the API
type Role struct {
	ID                string                 `json:"id"`
	OrgID             string                 `json:"orgId"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	ResourceType      string                 `json:"resourceType"`
	UXPurpose         string                 `json:"uxPurpose"`
	Details           map[string]interface{} `json:"details"`
	DefaultIdentifier string                 `json:"defaultIdentifier"`
	Version           int                    `json:"version"`
	IsOrgDefault      bool                   `json:"isOrgDefault"`
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ResourceType string                 `json:"resourceType,omitempty"`
	UXPurpose    string                 `json:"uxPurpose,omitempty"`
	Details      map[string]interface{} `json:"details"`
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	Name        string                 `json:"Name"`
	Description string                 `json:"Description"`
	Details     map[string]interface{} `json:"Details"`
}

func (c *Client) CreateRole(ctx context.Context, orgName string, req CreateRoleRequest) (*Role, error) {
	apiPath := path.Join("orgs", orgName, "roles")

	var role Role
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &role, nil
}

func (c *Client) GetRole(ctx context.Context, orgName string, roleID string) (*Role, error) {
	apiPath := path.Join("orgs", orgName, "roles", roleID)

	var role Role
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &role)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (c *Client) UpdateRole(ctx context.Context, orgName string, roleID string, req UpdateRoleRequest) (*Role, error) {
	apiPath := path.Join("orgs", orgName, "roles", roleID)

	var role Role
	_, err := c.do(ctx, http.MethodPatch, apiPath, req, &role)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return &role, nil
}

func (c *Client) DeleteRole(ctx context.Context, orgName string, roleID string) error {
	apiPath := path.Join("orgs", orgName, "roles", roleID)

	result, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		if result.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}
