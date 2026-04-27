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
	"errors"
	"fmt"
	"net/http"
	"path"
)

// EnvironmentMetadataClient is the slice of the Pulumi Cloud API needed to
// resolve an ESC environment's UUID. Kept as its own interface so the
// Environment resource can take just this surface (the rest of the env
// CRUD is served by the upstream `esc` client) and tests can stub it
// without faking the rest of the HTTP client.
type EnvironmentMetadataClient interface {
	GetEnvironmentMetadata(ctx context.Context, orgName, projectName, envName string) (*EnvironmentMetadata, error)
}

// EnvironmentMetadata mirrors the read-only metadata block returned by
// `GET /api/esc/environments/{org}/{project}/{env}/metadata`. We only
// surface fields the provider needs today; the wire shape carries more
// (ownedBy, gatedActions, …) which we ignore.
type EnvironmentMetadata struct {
	// ID is the environment's UUID, used as the `Identity` value when
	// pinning a custom RBAC role to a specific environment via a
	// PermissionLiteralExpressionEnvironment expression.
	ID string `json:"id"`
}

// GetEnvironmentMetadata fetches the metadata block (notably the env's UUID)
// for an existing ESC environment. Returns (nil, nil) when the env does not
// exist so callers can distinguish "not found" from a transport error.
func (c *Client) GetEnvironmentMetadata(
	ctx context.Context,
	orgName, projectName, envName string,
) (*EnvironmentMetadata, error) {
	if orgName == "" {
		return nil, errors.New("organization name must not be empty")
	}
	if projectName == "" {
		return nil, errors.New("project name must not be empty")
	}
	if envName == "" {
		return nil, errors.New("environment name must not be empty")
	}

	apiPath := path.Join("esc", "environments", orgName, projectName, envName, "metadata")
	var meta EnvironmentMetadata
	if _, err := c.do(ctx, http.MethodGet, apiPath, nil, &meta); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get environment metadata for %s/%s/%s: %w", orgName, projectName, envName, err)
	}
	return &meta, nil
}
