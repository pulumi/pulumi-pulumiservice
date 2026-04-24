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
	"fmt"
	"net/http"
)

type UserClient interface {
	GetCurrentUser(ctx context.Context) (*CurrentUser, error)
}

// CurrentUser describes the Pulumi Cloud user that the provider's configured
// access token belongs to. Mirrors the `/api/user` endpoint response; fields
// not exercised by the provider are intentionally omitted.
type CurrentUser struct {
	ID          string `json:"id"`
	GithubLogin string `json:"githubLogin"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatarUrl"`
}

func (c *Client) GetCurrentUser(ctx context.Context) (*CurrentUser, error) {
	var user CurrentUser
	if _, err := c.do(ctx, http.MethodGet, "user", nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return &user, nil
}
