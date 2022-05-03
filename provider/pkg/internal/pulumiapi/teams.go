// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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

type Teams struct {
	Teams []Team
}

type Team struct {
	Type        string `json:"kind"`
	Name        string
	DisplayName string
	Description string
	Members     []TeamMember
}

type TeamMember struct {
	Name        string
	GithubLogin string
	AvatarUrl   string
	Role        string
}

type createTeamRequest struct {
	Organization string `json:"organization"`
	TeamType     string `json:"teamType"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
}

type updateTeamRequest struct {
	NewDisplayName string `json:"newDisplayName"`
	NewDescription string `json:"newDescription"`
}

type updateTeamMembershipRequest struct {
	MemberAction string `json:"memberAction"`
	Member       string `json:"member"`
}

func (c *Client) ListTeams(ctx context.Context, orgName string) ([]Team, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	apiUrl := path.Join("orgs", orgName, "teams")

	var teamArray Teams
	_, err := c.do(ctx, http.MethodGet, apiUrl, nil, &teamArray)

	if err != nil {
		return nil, fmt.Errorf("failed to list teams for %q: %w", orgName, err)
	}
	return teamArray.Teams, nil
}

func (c *Client) GetTeam(ctx context.Context, orgName string, teamName string) (*Team, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(teamName) == 0 {
		return nil, errors.New("empty teamName")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName)

	var team Team
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &team)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &team, nil
}

func (c *Client) CreateTeam(ctx context.Context, orgName, teamName, teamType, displayName, description string) (*Team, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return nil, errors.New("teamname must not be empty")
	}

	if len(teamType) == 0 {
		return nil, errors.New("teamtype must not be empty")
	}

	teamtypeList := []string{"github", "pulumi"}
	if !contains(teamtypeList, teamType) {
		return nil, fmt.Errorf("teamtype must be one of %v, got %q", teamtypeList, teamType)
	}

	apiPath := path.Join("orgs", orgName, "teams", teamType)

	createReq := createTeamRequest{
		Organization: orgName,
		TeamType:     teamType,
		Name:         teamName,
		DisplayName:  displayName,
		Description:  description,
	}

	var team Team
	_, err := c.do(ctx, http.MethodPost, apiPath, createReq, &team)

	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return &team, nil
}

func (c *Client) UpdateTeam(ctx context.Context, orgName, teamName, displayName, description string) error {
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName)

	updateReq := updateTeamRequest{
		NewDisplayName: displayName,
		NewDescription: description,
	}

	_, err := c.do(ctx, "PATCH", apiPath, updateReq, nil)

	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}
	return nil
}

func (c *Client) DeleteTeam(ctx context.Context, orgName, teamName string) error {

	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}

func (c *Client) updateTeamMembership(ctx context.Context, orgName, teamName, userName, addOrRemove string) error {
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	if len(userName) == 0 {
		return errors.New("username must not be empty")
	}

	addOrRemoveValues := []string{"add", "remove"}
	if !contains(addOrRemoveValues, addOrRemove) {
		return errors.New("value must be `add` or `remove`")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamName)

	updateMembershipReq := updateTeamMembershipRequest{
		MemberAction: addOrRemove,
		Member:       userName,
	}

	_, err := c.do(ctx, http.MethodPatch, apiPath, updateMembershipReq, nil)

	if err != nil {
		return fmt.Errorf("failed to update team membership: %w", err)
	}
	return nil
}

func (c *Client) AddMemberToTeam(ctx context.Context, orgName, teamName, userName string) error {
	err := c.updateTeamMembership(ctx, orgName, teamName, userName, "add")
	if err != nil {
		var errResp *errorResponse
		if errors.As(err, &errResp) && errResp.StatusCode == http.StatusConflict {
			// ignore 409 since that means the team member is already added
			return nil
		}
		return err
	} else {
		return nil
	}
}

func (c *Client) DeleteMemberFromTeam(ctx context.Context, orgName, teamName, userName string) error {
	err := c.updateTeamMembership(ctx, orgName, teamName, userName, "remove")
	if err != nil {
		return err
	} else {
		return nil
	}
}
