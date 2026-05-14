// Copyright 2026, Pulumi Corporation.
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

package resources

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestDeploymentSettingsResourceID(t *testing.T) {
	assert.Equal(t, "org/proj/stack", deploymentSettingsResourceID("org", "proj", "stack"))
}

func TestNormalizeDurationString(t *testing.T) {
	t.Run("valid duration", func(t *testing.T) {
		got, err := normalizeDurationString("1h")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "1h0m0s", *got)
	})

	t.Run("empty input", func(t *testing.T) {
		_, err := normalizeDurationString("")
		require.Error(t, err)
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := normalizeDurationString("not-a-duration")
		require.Error(t, err)
	})
}

// TestDeploymentSettingsSourceContextRoundTrip guards against refresh drift on
// `sourceContext`: a Create stores the inputs verbatim, while a refresh rebuilds
// them from the API response via deploymentSettingsInputFromAPI. The two must
// agree, otherwise `pulumi refresh` reports a spurious update. Mirrors the
// `ts-deployment-settings` example (git source with sshAuth secrets).
func TestDeploymentSettingsSourceContextRoundTrip(t *testing.T) {
	password := "my_password"
	input := DeploymentSettingsInput{
		Organization: "org",
		Project:      "proj",
		Stack:        "stack",
		SourceContext: &DeploymentSettingsSourceContext{
			Git: &DeploymentSettingsGitSource{
				RepoURL: "https://github.com/pulumi/deploy-demos.git",
				Branch:  "refs/heads/main",
				RepoDir: "pulumi-programs/simple-resource",
				GitAuth: &DeploymentSettingsGitSourceGitAuth{
					SSHAuth: &DeploymentSettingsGitAuthSSHAuth{
						SSHPrivateKey: "key",
						Password:      &password,
					},
				},
			},
		},
	}

	stack := pulumiapi.StackIdentifier{OrgName: "org", ProjectName: "proj", StackName: "stack"}

	t.Run("cloud echoes plaintext secrets", func(t *testing.T) {
		api, err := input.toAPIDeploymentSettings()
		require.NoError(t, err)
		// Simulate the wire round-trip the cloud performs on GET.
		var returned pulumiapi.DeploymentSettings
		b, err := json.Marshal(api)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(b, &returned))

		got := deploymentSettingsInputFromAPI(stack, &returned, input)
		assert.Equal(t, input.SourceContext, got.SourceContext)
	})

	t.Run("cloud returns ciphertext secrets", func(t *testing.T) {
		api, err := input.toAPIDeploymentSettings()
		require.NoError(t, err)
		// The cloud stores secrets encrypted and returns ciphertext (no plaintext).
		ssh := api.SourceContext.Git.GitAuth.SSHAuth
		ssh.SSHPrivateKey = pulumiapi.SecretValue{Ciphertext: []byte("enc-key"), Secret: true}
		ssh.Password = &pulumiapi.SecretValue{Ciphertext: []byte("enc-pw"), Secret: true}

		got := deploymentSettingsInputFromAPI(stack, &api, input)
		assert.Equal(t, input.SourceContext, got.SourceContext)
	})

	t.Run("cloud omits write-only git auth on read", func(t *testing.T) {
		// Pulumi Cloud does not echo write-only git-auth credentials on GET.
		// The reconstruction must fall back to the previously-declared gitAuth
		// instead of dropping it, otherwise refresh reports a spurious diff.
		api, err := input.toAPIDeploymentSettings()
		require.NoError(t, err)
		api.SourceContext.Git.GitAuth = nil

		got := deploymentSettingsInputFromAPI(stack, &api, input)
		assert.Equal(t, input.SourceContext, got.SourceContext)
	})
}
