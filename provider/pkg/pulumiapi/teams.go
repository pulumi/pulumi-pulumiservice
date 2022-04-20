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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

func (c *Client) ListTeams(orgName string) (*[]Team, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	path := fmt.Sprintf("orgs/%s/teams", orgName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequest("GET", endpt.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		var teamArray Teams
		err = json.NewDecoder(res.Body).Decode(&teamArray)
		if err != nil {
			return nil, err
		}

		return &teamArray.Teams, nil
	case 400, 401, 403, 404, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes

	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) GetTeam(orgName string, teamName string) (*Team, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(teamName) == 0 {
		return nil, errors.New("empty orgName")
	}

	path := fmt.Sprintf("orgs/%s/teams/%s", orgName, teamName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequest("GET", endpt.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		var team Team
		err = json.NewDecoder(res.Body).Decode(&team)
		if err != nil {
			return nil, err
		}

		return &team, nil
	case 400, 401, 403, 404, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes

	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) CreateTeam(orgName string, teamName string, teamType string, displayName string, description string) (*Team, error) {
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
	if !Contains(teamtypeList, teamType) {
		return nil, errors.New("teamtype must be either `pulumi` or `github`")
	}

	path := fmt.Sprintf("orgs/%s/teams/%s", orgName, teamType)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	values := map[string]string{"organization": orgName, "teamType": teamType, "name": teamName, "displayName": displayName, "description": description}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpt.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200, 201:
		var team Team
		err = json.NewDecoder(res.Body).Decode(&team)
		if err != nil {
			return nil, err
		}

		return &team, nil
	case 400, 401, 403, 404, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes

	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) UpdateTeam(orgName string, teamName string, displayName string, description string) (*Team, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return nil, errors.New("teamname must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/teams/%s", orgName, teamName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	values := map[string]string{
		"newDisplayName": displayName,
		"newDescription": description,
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", endpt.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)

	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 204:
		var team Team
		team.Description = description
		team.DisplayName = displayName
		return &team, nil
	case 400, 401, 403, 404, 405, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes
	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) DeleteTeam(orgName string, teamName string) error {

	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/teams/%s", orgName, teamName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequest("DELETE", endpt.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)

	res, err := c.c.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 204:
		return nil
	case 400, 401, 403, 404, 405, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return &errRes
	default:
		return fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) updateTeamMembership(orgName string, teamName string, userName string, addOrRemove string) error {
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
	if !Contains(addOrRemoveValues, addOrRemove) {
		return errors.New("value must be `add` or `remove`")
	}

	path := fmt.Sprintf("orgs/%s/teams/%s", orgName, teamName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	values := map[string]string{"memberAction": addOrRemove, "member": userName}
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", endpt.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)

	res, err := c.c.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200, 204:
		return nil
	case 400, 401, 403, 404, 405, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return &errRes
	default:
		return fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) AddMemberToTeam(orgName string, teamName string, userName string) error {
	err := c.updateTeamMembership(orgName, teamName, userName, "add")
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (c *Client) DeleteMemberFromTeam(orgName string, teamName string, userName string) error {
	err := c.updateTeamMembership(orgName, teamName, userName, "remove")
	if err != nil {
		return err
	} else {
		return nil
	}
}
