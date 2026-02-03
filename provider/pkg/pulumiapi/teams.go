// Copyright 2016-2022, Pulumi Corporation.
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

type TeamClient interface {
	ListTeams(ctx context.Context, orgName string) ([]Team, error)
	GetTeam(ctx context.Context, orgName string, teamName string) (*Team, error)
	CreateTeam(
		ctx context.Context,
		orgName, teamName, teamType, displayName, description string,
		teamID int64,
	) (*Team, error)
	UpdateTeam(ctx context.Context, orgName, teamName, displayName, description string) error
	DeleteTeam(ctx context.Context, orgName, teamName string) error
	AddMemberToTeam(ctx context.Context, orgName, teamName, userName string) error
	DeleteMemberFromTeam(ctx context.Context, orgName, teamName, userName string) error
	AddStackPermission(ctx context.Context, stack StackIdentifier, teamName string, permission int) error
	RemoveStackPermission(ctx context.Context, stack StackIdentifier, teamName string) error
	GetTeamStackPermission(ctx context.Context, stack StackIdentifier, teamName string) (*int, error)
	AddEnvironmentSettings(ctx context.Context, req CreateTeamEnvironmentSettingsRequest) error
	RemoveEnvironmentSettings(ctx context.Context, req TeamEnvironmentSettingsRequest) error
	GetTeamEnvironmentSettings(ctx context.Context, req TeamEnvironmentSettingsRequest) (*string, *Duration, error)
}

type Teams struct {
	Teams []Team
}

type Team struct {
	Type         string `json:"kind"`
	Name         string
	DisplayName  string
	Description  string
	Members      []TeamMember
	Stacks       []TeamStackPermission
	Environments []TeamEnvironmentSettings
}

type TeamMember struct {
	Name        string
	GithubLogin string
	AvatarURL   string
	Role        string
}

type TeamStackPermission struct {
	ProjectName string `json:"projectName"`
	StackName   string `json:"stackName"`
	Permission  int    `json:"permission"`
}

type TeamEnvironmentSettings struct {
	EnvName         string    `json:"envName"`
	ProjectName     string    `json:"projectName"`
	Permission      string    `json:"permission"`
	MaxOpenDuration *Duration `json:"maxOpenDuration,omitempty"`
}

type createTeamRequest struct {
	Organization string `json:"organization"`
	TeamType     string `json:"teamType"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	GitHubTeamID int64  `json:"githubTeamID,omitempty"`
}

type updateTeamRequest struct {
	NewDisplayName string `json:"newDisplayName"`
	NewDescription string `json:"newDescription"`
}

type updateTeamMembershipRequest struct {
	MemberAction string `json:"memberAction"`
	Member       string `json:"member"`
}

type AddStackPermission struct {
	ProjectName string `json:"projectName"`
	StackName   string `json:"stackName"`
	Permission  int    `json:"permission"`
}

type addStackPermissionRequest struct {
	AddStackPermission AddStackPermission `json:"addStackPermission"`
}

type RemoveStackPermission struct {
	ProjectName string `json:"projectName"`
	StackName   string `json:"stackName"`
}

type removeStackPermissionRequest struct {
	RemoveStackPermission RemoveStackPermission `json:"removeStack"`
}

type CreateTeamEnvironmentSettingsRequest struct {
	TeamEnvironmentSettingsRequest
	Permission      string    `json:"permission,omitempty"`
	MaxOpenDuration *Duration `json:"maxOpenDuration,omitempty"`
}

type TeamEnvironmentSettingsRequest struct {
	Organization string `json:"organization,omitempty"`
	Team         string `json:"team,omitempty"`
	Environment  string `json:"environment,omitempty"`
	Project      string `json:"project,omitempty"`
}

type AddEnvironmentPermission struct {
	EnvName         string    `json:"envName"`
	ProjectName     string    `json:"projectName"`
	Permission      string    `json:"permission"`
	MaxOpenDuration *Duration `json:"maxOpenDuration,omitempty"`
}

type addEnvironmentSettingsRequest struct {
	AddEnvironmentPermission AddEnvironmentPermission `json:"addEnvironmentPermission"`
}

type RemoveEnvironmentPermission struct {
	EnvName     string `json:"envName"`
	ProjectName string `json:"projectName"`
}

type removeEnvironmentPermissionRequest struct {
	RemoveEnvironment RemoveEnvironmentPermission `json:"removeEnvironment"`
}

func (c *Client) ListTeams(ctx context.Context, orgName string) ([]Team, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	apiURL := path.Join("orgs", orgName, "teams")

	var teamArray Teams
	_, err := c.do(ctx, http.MethodGet, apiURL, nil, &teamArray)
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
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			// Important: we return nil here to hint it was not found
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &team, nil
}

func (c *Client) CreateTeam(
	ctx context.Context,
	orgName, teamName, teamType, displayName, description string,
	teamID int64,
) (*Team, error) {
	teamtypeList := []string{"github", "pulumi"}
	if !contains(teamtypeList, teamType) {
		return nil, fmt.Errorf("teamtype must be one of %v, got %q", teamtypeList, teamType)
	}

	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 && teamType != "github" {
		return nil, errors.New("teamname must not be empty")
	}

	if teamType == "github" && teamID == 0 {
		return nil, errors.New("github teams require a githubTeamId")
	}

	apiPath := path.Join("orgs", orgName, "teams", teamType)

	createReq := createTeamRequest{
		Organization: orgName,
		TeamType:     teamType,
		Name:         teamName,
		DisplayName:  displayName,
		Description:  description,
		GitHubTeamID: teamID,
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
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusConflict {
			// ignore 409 since that means the team member is already added
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) DeleteMemberFromTeam(ctx context.Context, orgName, teamName, userName string) error {
	err := c.updateTeamMembership(ctx, orgName, teamName, userName, "remove")
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AddStackPermission(ctx context.Context, stack StackIdentifier, teamName string, permission int) error {
	if len(stack.OrgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	apiPath := path.Join("orgs", stack.OrgName, "teams", teamName)

	addStackPermissionRequest := addStackPermissionRequest{
		AddStackPermission: AddStackPermission{
			ProjectName: stack.ProjectName,
			StackName:   stack.StackName,
			Permission:  permission,
		},
	}

	_, err := c.do(ctx, http.MethodPatch, apiPath, addStackPermissionRequest, nil)
	if err != nil {
		return fmt.Errorf("failed to add stack permission for team: %w", err)
	}
	return nil
}

func (c *Client) RemoveStackPermission(ctx context.Context, stack StackIdentifier, teamName string) error {
	if len(stack.OrgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	apiPath := path.Join("orgs", stack.OrgName, "teams", teamName)

	removeStackPermissionRequest := removeStackPermissionRequest{
		RemoveStackPermission: RemoveStackPermission{ProjectName: stack.ProjectName, StackName: stack.StackName},
	}

	_, err := c.do(ctx, http.MethodPatch, apiPath, removeStackPermissionRequest, nil)
	if err != nil {
		return fmt.Errorf("failed to remove stack permission for team: %w", err)
	}
	return nil
}

func (c *Client) GetTeamStackPermission(ctx context.Context, stack StackIdentifier, teamName string) (*int, error) {
	if len(stack.OrgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return nil, errors.New("teamname must not be empty")
	}

	apiPath := path.Join("orgs", stack.OrgName, "teams", teamName)

	var team Team
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &team)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	for _, stackPermission := range team.Stacks {
		if stackPermission.ProjectName == stack.ProjectName && stackPermission.StackName == stack.StackName {
			return &stackPermission.Permission, nil
		}
	}

	return nil, nil
}

func (c *Client) AddEnvironmentSettings(ctx context.Context, req CreateTeamEnvironmentSettingsRequest) error {
	if len(req.Organization) == 0 {
		return errors.New("organization name must not be empty")
	}
	if len(req.Team) == 0 {
		return errors.New("team name must not be empty")
	}
	if len(req.Environment) == 0 {
		return errors.New("environment name must not be empty")
	}

	apiPath := path.Join("orgs", req.Organization, "teams", req.Team)

	addEnvironmentSettingsRequest := addEnvironmentSettingsRequest{
		AddEnvironmentPermission: AddEnvironmentPermission{
			ProjectName:     req.Project,
			EnvName:         req.Environment,
			Permission:      req.Permission,
			MaxOpenDuration: req.MaxOpenDuration,
		},
	}

	_, err := c.do(ctx, http.MethodPatch, apiPath, addEnvironmentSettingsRequest, nil)
	if err != nil {
		return fmt.Errorf(
			"failed to add team settings for environment %s to team %s due to error: %w",
			req.Environment,
			req.Team,
			err,
		)
	}
	return nil
}

func (c *Client) RemoveEnvironmentSettings(ctx context.Context, req TeamEnvironmentSettingsRequest) error {
	if len(req.Organization) == 0 {
		return errors.New("organization name must not be empty")
	}
	if len(req.Team) == 0 {
		return errors.New("team name must not be empty")
	}
	if len(req.Environment) == 0 {
		return errors.New("environment name must not be empty")
	}

	apiPath := path.Join("orgs", req.Organization, "teams", req.Team)

	removeEnvironmentSettingsRequest := removeEnvironmentPermissionRequest{
		RemoveEnvironment: RemoveEnvironmentPermission{
			ProjectName: req.Project,
			EnvName:     req.Environment,
		},
	}

	_, err := c.do(ctx, http.MethodPatch, apiPath, removeEnvironmentSettingsRequest, nil)
	if err != nil {
		return fmt.Errorf(
			"failed to remove permissions for environment %s from team %s due to error: %w",
			req.Environment,
			req.Team,
			err,
		)
	}
	return nil
}

func (c *Client) GetTeamEnvironmentSettings(
	ctx context.Context,
	req TeamEnvironmentSettingsRequest,
) (*string, *Duration, error) {
	if len(req.Organization) == 0 {
		return nil, nil, errors.New("organization name must not be empty")
	}
	if len(req.Team) == 0 {
		return nil, nil, errors.New("team name must not be empty")
	}
	if len(req.Environment) == 0 {
		return nil, nil, errors.New("environment name must not be empty")
	}

	apiPath := path.Join("orgs", req.Organization, "teams", req.Team)

	var team Team
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &team)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get team environment permission: %w", err)
	}

	for _, settings := range team.Environments {
		if settings.EnvName == req.Environment {
			return &settings.Permission, settings.MaxOpenDuration, nil
		}
	}

	return nil, nil, nil
}
