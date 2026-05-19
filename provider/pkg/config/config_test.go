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

// AccessToken is read from PULUMI_ACCESS_TOKEN when the explicit input
// is empty. APIURL fallback (PULUMI_BACKEND_URL / PULUMI_API) is handled
// by infer's SetDefault before Configure runs, so it isn't exercised here.
func TestConfigure_AccessTokenFallsBackToEnvVar(t *testing.T) {
	t.Setenv(EnvVarPulumiAccessToken, "pul-test-token")

	c := &Config{}
	require.NoError(t, c.Configure(context.Background()))

	assert.Equal(t, "pul-test-token", c.AccessToken,
		"AccessToken must fall back to PULUMI_ACCESS_TOKEN when input is empty")
}

// Explicit AccessToken wins over the env-var fallback.
func TestConfigure_ExplicitAccessTokenBeatsEnvVar(t *testing.T) {
	t.Setenv(EnvVarPulumiAccessToken, "pul-env-token")

	c := &Config{AccessToken: "pul-explicit-token"}
	require.NoError(t, c.Configure(context.Background()))

	assert.Equal(t, "pul-explicit-token", c.AccessToken,
		"explicit AccessToken must not be overridden by env var fallback")
}

// No config, no env, no stored credentials → ErrAccessTokenNotFound.
// Skips when ~/.pulumi/credentials.json is populated (dev machines).
func TestConfigure_EmptyInputsAndNoEnvVarsErrors(t *testing.T) {
	t.Setenv(EnvVarPulumiAccessToken, "")
	c := &Config{}
	err := c.Configure(context.Background())
	if err == nil {
		t.Skipf("workspace.GetStoredCredentials returned a token; skipping. AccessToken=%q", c.AccessToken)
	}
	assert.ErrorIs(t, err, ErrAccessTokenNotFound)
}
