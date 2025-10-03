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

// AccessToken represents a Pulumi Service access token.
type AccessToken struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	TokenValue  string `json:"tokenValue"`
	Description string `json:"description"`
	Admin       bool   `json:"admin"`
}

type createTokenResponse struct {
	ID         string `json:"id"`
	TokenValue string `json:"tokenValue"`
}

type createTokenRequest struct {
	Description string `json:"description"`
}

type accessTokenResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LastUsed    int    `json:"lastUsed"`
	Admin       bool   `json:"admin"`
}

type listTokenResponse struct {
	Tokens []accessTokenResponse `json:"tokens"`
}

// CreateAccessToken creates a new access token with the specified description.
func (c *Client) CreateAccessToken(ctx context.Context, description string) (*AccessToken, error) {
	apiPath := path.Join("user", "tokens")

	createReq := createTokenRequest{
		Description: description,
	}

	var createRes createTokenResponse

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

// DeleteAccessToken deletes an access token by its ID.
func (c *Client) DeleteAccessToken(ctx context.Context, tokenID string) error {
	if len(tokenID) == 0 {
		return errors.New("tokenid length must be greater than zero")
	}

	apiPath := path.Join("user", "tokens", tokenID)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete access token %q: %w", tokenID, err)
	}

	return nil
}

// GetAccessToken retrieves an access token by its ID.
func (c *Client) GetAccessToken(ctx context.Context, id string) (*AccessToken, error) {
	apiPath := path.Join("user", "tokens")

	var listRes listTokenResponse

	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &listRes)

	if err != nil {
		return nil, fmt.Errorf("failed to list access tokens: %w", err)
	}

	for i := 0; i < len(listRes.Tokens); i++ {
		token := listRes.Tokens[i]
		if token.ID == id {
			return &AccessToken{
				ID:          token.ID,
				Description: token.Description,
			}, nil
		}
	}

	return nil, nil
}
