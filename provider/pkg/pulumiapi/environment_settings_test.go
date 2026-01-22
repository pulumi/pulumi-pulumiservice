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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvironmentSettings(t *testing.T) {
	t.Run("Happy Path - DeletionProtected true", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/settings",
			ResponseCode:      200,
			ResponseBody: EnvironmentSettings{
				DeletionProtected: true,
			},
		})
		settings, err := c.GetEnvironmentSettings(ctx, "org", "project", "env")
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.True(t, settings.DeletionProtected)
	})

	t.Run("Happy Path - DeletionProtected false", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/settings",
			ResponseCode:      200,
			ResponseBody: EnvironmentSettings{
				DeletionProtected: false,
			},
		})
		settings, err := c.GetEnvironmentSettings(ctx, "org", "project", "env")
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.False(t, settings.DeletionProtected)
	})

	t.Run("404 returns default settings", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/settings",
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})
		settings, err := c.GetEnvironmentSettings(ctx, "org", "project", "env")
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.False(t, settings.DeletionProtected)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/settings",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		settings, err := c.GetEnvironmentSettings(ctx, "org", "project", "env")
		assert.Nil(t, settings)
		assert.EqualError(t, err, "failed to get environment settings for org/project/env: 401 API error: unauthorized")
	})
}

func TestUpdateEnvironmentSettings(t *testing.T) {
	deletionProtectedTrue := true
	deletionProtectedFalse := false

	t.Run("Happy Path - Enable deletion protection", func(t *testing.T) {
		req := UpdateEnvironmentSettingsRequest{
			DeletionProtected: &deletionProtectedTrue,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/settings",
			ExpectedReqBody:   req,
			ResponseCode:      200,
		})
		err := c.UpdateEnvironmentSettings(ctx, "org", "project", "env", req)
		assert.NoError(t, err)
	})

	t.Run("Happy Path - Disable deletion protection", func(t *testing.T) {
		req := UpdateEnvironmentSettingsRequest{
			DeletionProtected: &deletionProtectedFalse,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/settings",
			ExpectedReqBody:   req,
			ResponseCode:      200,
		})
		err := c.UpdateEnvironmentSettings(ctx, "org", "project", "env", req)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		req := UpdateEnvironmentSettingsRequest{
			DeletionProtected: &deletionProtectedTrue,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/settings",
			ExpectedReqBody:   req,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.UpdateEnvironmentSettings(ctx, "org", "project", "env", req)
		assert.EqualError(t, err, "failed to update environment settings for org/project/env: 401 API error: unauthorized")
	})
}
