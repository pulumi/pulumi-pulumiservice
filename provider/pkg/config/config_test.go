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

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// `pulumi import` persists default-provider inputs as `{}`. On later
// operations, Configure receives empty AccessToken and empty APIURL.
// Both must fall back to env vars; otherwise destroy against a non-prod
// backend dials api.pulumi.com and 401s.
func TestConfigure_FallsBackToEnvVarsWhenInputsAreEmpty(t *testing.T) {
	t.Setenv(EnvVarPulumiAccessToken, "pul-test-token")
	t.Setenv(EnvVarPulumiBackendURL, "https://test-backend.example/")

	c := &Config{} // Empty inputs — what `pulumi import` persists.
	require.NoError(t, c.Configure(context.Background()))

	assert.Equal(t, "pul-test-token", c.AccessToken,
		"AccessToken must fall back to PULUMI_ACCESS_TOKEN when input is empty")
	assert.Equal(t, "https://test-backend.example/", c.APIURL,
		"APIURL must fall back to PULUMI_BACKEND_URL when input is empty — "+
			"otherwise NewClient defaults to api.pulumi.com and a non-prod "+
			"token will 401 on destroy of imported resources")
}

// Explicit config (from stack config or RPC) wins over env-var fallback.
func TestConfigure_ExplicitConfigBeatsEnvVar(t *testing.T) {
	t.Setenv(EnvVarPulumiAccessToken, "pul-env-token")
	t.Setenv(EnvVarPulumiBackendURL, "https://env-backend.example/")

	c := &Config{
		AccessToken: "pul-explicit-token",
		APIURL:      "https://explicit-backend.example/",
	}
	require.NoError(t, c.Configure(context.Background()))

	assert.Equal(t, "pul-explicit-token", c.AccessToken,
		"explicit AccessToken must not be overridden by env var fallback")
	assert.Equal(t, "https://explicit-backend.example/", c.APIURL,
		"explicit APIURL must not be overridden by env var fallback")
}

// No config, no env, no stored credentials → ErrAccessTokenNotFound.
// Skips when ~/.pulumi/credentials.json is populated (dev machines).
func TestConfigure_EmptyInputsAndNoEnvVarsErrors(t *testing.T) {
	t.Setenv(EnvVarPulumiAccessToken, "")
	t.Setenv(EnvVarPulumiBackendURL, "")
	c := &Config{}
	err := c.Configure(context.Background())
	if err == nil {
		t.Skipf("workspace.GetStoredCredentials returned a token; skipping. AccessToken=%q", c.AccessToken)
	}
	assert.ErrorIs(t, err, ErrAccessTokenNotFound)
}
