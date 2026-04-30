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

// TestConfigure_FallsBackToEnvVarsWhenInputsAreEmpty pins the round-trip
// for `pulumi import`-persisted state. The Pulumi engine persists the
// default-provider's `inputs` from the original Configure call; on
// `pulumi up` of a new resource, that snapshot is the SetDefault-resolved
// config (including apiUrl from PULUMI_BACKEND_URL). On `pulumi import`,
// the engine instead persists `inputs: {}` — both `accessToken` and
// `apiUrl` arrive at this Configure call as empty strings.
//
// AccessToken's manual env-var fallback (the one that's been in this
// function since before this test was written) covers it; the apiUrl
// fallback added alongside this test mirrors that. Without both
// fallbacks, destroy of an imported resource against any non-prod
// Pulumi Cloud (review stack, self-hosted) silently dials
// https://api.pulumi.com per NewClient's default and 401s with a
// non-prod token.
//
// Symptom this guards against:
//
//	failed to delete role: 401 API error: Unauthorized:
//	No credentials provided or are invalid
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

// TestConfigure_ExplicitConfigBeatsEnvVar pins the precedence: a
// non-empty value already set in c.AccessToken / c.APIURL (from stack
// config or the engine's Configure RPC) is not overwritten by the
// env-var fallback. Without this, a user who sets
// `pulumi config set pulumiservice:apiUrl https://stack-config.example`
// would have it silently overridden by whatever PULUMI_BACKEND_URL was
// in the calling shell.
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

// TestConfigure_EmptyInputsAndNoEnvVarsErrors documents the no-creds
// path: empty config, no env vars, no stored credentials → returns
// ErrAccessTokenNotFound. APIURL stays empty, but that's fine because
// the call returns before NewClient is reached (the AccessToken check
// comes first).
func TestConfigure_EmptyInputsAndNoEnvVarsErrors(t *testing.T) {
	t.Setenv(EnvVarPulumiAccessToken, "")
	t.Setenv(EnvVarPulumiBackendURL, "")
	// Note: workspace.GetStoredCredentials reads ~/.pulumi/credentials.json
	// which may be present on dev machines. We can't reliably stub it from
	// here without changing the production code, so this test only runs
	// when the file isn't there or the current backend has no token.
	// Skipping the assertion when credentials are present keeps the test
	// honest in CI (clean environment) without fighting the developer's
	// local login state.
	c := &Config{}
	err := c.Configure(context.Background())
	if err == nil {
		t.Skipf("workspace.GetStoredCredentials returned a token; can't test the "+
			"no-creds path on this machine. AccessToken=%q", c.AccessToken)
	}
	assert.ErrorIs(t, err, ErrAccessTokenNotFound,
		"no creds anywhere must return ErrAccessTokenNotFound")
}
