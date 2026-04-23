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
	"path"
)

type TeamRoleClient interface {
	EnableTeamRoles(ctx context.Context, orgName, teamName string) error
	AssignRoleToTeam(ctx context.Context, orgName, teamName, roleID string) error
	RemoveRoleFromTeam(ctx context.Context, orgName, teamName, roleID string) error
	ListTeamRoles(ctx context.Context, orgName, teamName string) ([]TeamRoleRef, error)
	GetTeamRole(ctx context.Context, orgName, teamName, roleID string) (*TeamRoleRef, error)
}

type TeamRoleRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listTeamRolesResponse struct {
	Roles []TeamRoleRef `json:"roles"`
}

// EnableTeamRoles flips the per-team custom-roles feature on. The endpoint is
// idempotent at the caller level: if it's already enabled, we swallow the
// 400/409 the service returns.
func (c *Client) EnableTeamRoles(ctx context.Context, orgName, teamName string) error {
	if len(orgName) == 0 {
		return errors.New("organization name should not be empty")
	}
	if len(teamName) == 0 {
		return errors.New("team name should not be empty")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName, "enable-team-roles")
	if _, err := c.do(ctx, http.MethodPost, apiPath, nil, nil); err != nil {
		switch GetErrorStatusCode(err) {
		case http.StatusConflict, http.StatusBadRequest:
			// Already enabled — best-effort idempotency.
			return nil
		}
		return fmt.Errorf("failed to enable team roles: %w", err)
	}
	return nil
}

// AssignRoleToTeam assigns a custom role to a team. Because the team's custom-
// roles feature is a one-time opt-in and there's no reliable read-side signal,
// we call EnableTeamRoles first and treat 400/409 as already-enabled.
func (c *Client) AssignRoleToTeam(ctx context.Context, orgName, teamName, roleID string) error {
	if len(orgName) == 0 {
		return errors.New("organization name should not be empty")
	}
	if len(teamName) == 0 {
		return errors.New("team name should not be empty")
	}
	if len(roleID) == 0 {
		return errors.New("role id should not be empty")
	}

	if err := c.EnableTeamRoles(ctx, orgName, teamName); err != nil {
		return err
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName, "roles", roleID)
	// Service registers this as POST (despite the Java spec showing PUT).
	if _, err := c.do(ctx, http.MethodPost, apiPath, nil, nil); err != nil {
		return fmt.Errorf("failed to assign role to team: %w", err)
	}
	return nil
}

func (c *Client) RemoveRoleFromTeam(ctx context.Context, orgName, teamName, roleID string) error {
	if len(orgName) == 0 {
		return errors.New("organization name should not be empty")
	}
	if len(teamName) == 0 {
		return errors.New("team name should not be empty")
	}
	if len(roleID) == 0 {
		return errors.New("role id should not be empty")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName, "roles", roleID)
	if _, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("failed to remove role from team: %w", err)
	}
	return nil
}

func (c *Client) ListTeamRoles(ctx context.Context, orgName, teamName string) ([]TeamRoleRef, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name should not be empty")
	}
	if len(teamName) == 0 {
		return nil, errors.New("team name should not be empty")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName, "roles")
	var resp listTeamRolesResponse
	if _, err := c.do(ctx, http.MethodGet, apiPath, nil, &resp); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list team roles: %w", err)
	}
	return resp.Roles, nil
}

func (c *Client) GetTeamRole(ctx context.Context, orgName, teamName, roleID string) (*TeamRoleRef, error) {
	roles, err := c.ListTeamRoles(ctx, orgName, teamName)
	if err != nil {
		return nil, err
	}
	for i, r := range roles {
		if r.ID == roleID {
			return &roles[i], nil
		}
	}
	return nil, nil
}
