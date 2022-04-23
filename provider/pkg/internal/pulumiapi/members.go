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

type addMemberToOrgReq struct {
	Role string `json:"role"`
}

func (c *Client) AddMemberToOrg(ctx context.Context, userName string, orgName string, role string) error {

	if len(userName) == 0 {
		return errors.New("username should not be empty")
	}
	if len(orgName) == 0 {
		return errors.New("organization name should not be empty")
	}

	roleList := []string{"admin", "member"}

	if !contains(roleList, role) {
		return fmt.Errorf("role must be one of: %v", roleList)
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

func (c *Client) ListOrgMembers(ctx context.Context, orgName string) (*Members, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	apiPath := path.Join("orgs", orgName, "members")

	req, err := c.createRequest(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.URL.RawQuery = "type=backend"

	var members Members
	_, err = c.sendRequest(req, &members)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization members: %w", err)
	}

	return &members, nil

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
