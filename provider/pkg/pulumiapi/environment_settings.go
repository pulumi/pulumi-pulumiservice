// Copyright 2016-2026, Pulumi Corporation.
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
	"fmt"
	"net/http"
	"path"
)

type EnvironmentSettingsClient interface {
	GetEnvironmentSettings(ctx context.Context, orgName, projectName, envName string) (*EnvironmentSettings, error)
	UpdateEnvironmentSettings(ctx context.Context, orgName, projectName, envName string, req UpdateEnvironmentSettingsRequest) error
}

type EnvironmentSettings struct {
	DeletionProtected bool `json:"deletionProtected"`
}

type UpdateEnvironmentSettingsRequest struct {
	DeletionProtected *bool `json:"deletionProtected,omitempty"`
}

func (c *Client) GetEnvironmentSettings(ctx context.Context, orgName, projectName, envName string) (*EnvironmentSettings, error) {
	apiPath := path.Join("esc", "environments", orgName, projectName, envName, "settings")
	var settings EnvironmentSettings
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &settings)
	if err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			// Return default settings if endpoint returns 404
			return &EnvironmentSettings{DeletionProtected: false}, nil
		}
		return nil, fmt.Errorf("failed to get environment settings for %s/%s/%s: %w", orgName, projectName, envName, err)
	}
	return &settings, nil
}

func (c *Client) UpdateEnvironmentSettings(ctx context.Context, orgName, projectName, envName string, req UpdateEnvironmentSettingsRequest) error {
	apiPath := path.Join("esc", "environments", orgName, projectName, envName, "settings")
	_, err := c.do(ctx, http.MethodPatch, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to update environment settings for %s/%s/%s: %w", orgName, projectName, envName, err)
	}
	return nil
}
