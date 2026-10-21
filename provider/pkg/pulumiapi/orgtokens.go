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

type OrgAccessTokenClient interface {
	CreateOrgAccessToken(ctx context.Context, orgName string, req CreateOrgAccessTokenRequest) (*AccessToken, error)
	DeleteOrgAccessToken(ctx context.Context, tokenID, orgName string) error
	GetOrgAccessToken(ctx context.Context, tokenID, orgName string) (*AccessToken, error)
}

type createOrgTokenResponse struct {
	ID         string `json:"id"`
	TokenValue string `json:"tokenValue"`
}

type CreateOrgAccessTokenRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Admin       bool   `json:"admin"`
	Expires     int64  `json:"expires,omitempty"`
	RoleID      string `json:"roleID,omitempty"`
}

func (c *Client) CreateOrgAccessToken(
	ctx context.Context,
	orgName string,
	createReq CreateOrgAccessTokenRequest,
) (*AccessToken, error) {

	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(createReq.Name) == 0 {
		return nil, errors.New("empty name")
	}

	apiPath := path.Join("orgs", orgName, "tokens")

	var createRes createOrgTokenResponse

	_, err := c.do(ctx, http.MethodPost, apiPath, createReq, &createRes)

	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	return &AccessToken{
		ID:          createRes.ID,
		TokenValue:  createRes.TokenValue,
		Description: createReq.Description,
	}, nil

}

func (c *Client) DeleteOrgAccessToken(ctx context.Context, tokenID, orgName string) error {
	if len(tokenID) == 0 {
		return errors.New("tokenid length must be greater than zero")
	}

	if len(orgName) == 0 {
		return errors.New("orgname length must be greater than zero")
	}

	apiPath := path.Join("orgs", orgName, "tokens", tokenID)

	fmt.Println(apiPath)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete access token %q: %w", tokenID, err)
	}

	return nil
}

func (c *Client) GetOrgAccessToken(ctx context.Context, tokenID, orgName string) (*AccessToken, error) {
	apiPath := path.Join("orgs", orgName, "tokens")

	var listRes listTokenResponse

	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &listRes)

	if err != nil {
		return nil, fmt.Errorf("failed to list org access tokens: %w", err)
	}

	for i := 0; i < len(listRes.Tokens); i++ {
		token := listRes.Tokens[i]
		if token.ID == tokenID {
			accessToken := &AccessToken{
				ID:          token.ID,
				Name:        token.Name,
				Description: token.Description,
				Admin:       token.Admin,
				Expires:     token.Expires,
			}
			if token.Role != nil {
				accessToken.RoleID = token.Role.ID
			}
			return accessToken, nil
		}
	}

	return nil, nil
}
