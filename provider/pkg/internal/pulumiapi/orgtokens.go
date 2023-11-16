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

type createOrgTokenResponse struct {
	ID         string `json:"id"`
	TokenValue string `json:"tokenValue"`
}

type createOrgTokenRequest struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Admin       bool   `json:"admin"`
}

func (c *Client) CreateOrgAccessToken(ctx context.Context, name, orgName, description string, admin bool) (*AccessToken, error) {

	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(name) == 0 {
		return nil, errors.New("empty name")
	}

	apiPath := path.Join("orgs", orgName, "tokens")

	createReq := createOrgTokenRequest{
		Name:        name,
		Description: description,
		Admin:       admin,
	}

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

func (c *Client) DeleteOrgAccessToken(ctx context.Context, tokenId, orgName string) error {
	if len(tokenId) == 0 {
		return errors.New("tokenid length must be greater than zero")
	}

	if len(orgName) == 0 {
		return errors.New("orgname length must be greater than zero")
	}

	apiPath := path.Join("orgs", orgName, "tokens", tokenId)

	fmt.Println(apiPath)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete access token %q: %w", tokenId, err)
	}

	return nil
}

func (c *Client) GetOrgAccessToken(ctx context.Context, tokenId, orgName string) (*AccessToken, error) {
	apiPath := path.Join("orgs", orgName, "tokens", tokenId)

	var listRes listTokenResponse

	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &listRes)

	if err != nil {
		return nil, fmt.Errorf("failed to list org access tokens: %w", err)
	}

	for i := 0; i < len(listRes.Tokens); i++ {
		token := listRes.Tokens[i]
		if token.ID == tokenId {
			return &AccessToken{
				ID:          token.ID,
				Description: token.Description,
			}, nil
		}
	}

	return nil, nil
}
