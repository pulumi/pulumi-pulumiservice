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

type Members struct {
	Members []Member
}

type Member struct {
	Role          string
	User          User
	KnownToPulumi bool
	VirtualAdmin  bool
}

type User struct {
	Name        string
	GithubLogin string
	AvatarUrl   string
	Email       string
}

func (c *Client) AddMemberToOrg(userName string, orgName string, role string) error {

	if len(userName) == 0 {
		return errors.New("username should not be empty")
	}
	if len(orgName) == 0 {
		return errors.New("organisation name should not be empty")
	}

	roleList := []string{"admin", "member"}

	if !Contains(roleList, role) {
		return errors.New("role must be either an admin or a member")
	}

	path := fmt.Sprintf("orgs/%s/members/%s", orgName, userName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	values := map[string]string{"role": role}
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", endpt.String(), bytes.NewBuffer(data))
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
		var members Members
		err = json.NewDecoder(res.Body).Decode(&members)
		if err != nil {
			return err
		}

		return nil
	case 400, 401, 403, 404, 500:
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

func (c *Client) ListOrgMembers(orgName string) (*[]Member, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	path := fmt.Sprintf("orgs/%s/members", orgName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequest("GET", endpt.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)

	req.URL.RawQuery = "type=backend"

	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		var memberArray Members
		err = json.NewDecoder(res.Body).Decode(&memberArray)
		if err != nil {
			return nil, err
		}

		return &memberArray.Members, nil
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

func (c *Client) DeleteMemberFromOrg(orgName string, userName string) error {
	if len(orgName) == 0 {
		return errors.New("orgName must not be empty")
	}

	if len(userName) == 0 {
		return errors.New("userName must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/members/%s", orgName, userName)
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
