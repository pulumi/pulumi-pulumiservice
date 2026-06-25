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
	"context"
	"encoding/json"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
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

// TestDeploymentSettingsEnvVarRefresh pins down how refreshed environment
// variables reconcile the API response with the previously-declared inputs:
// plaintext echoes are authoritative (out-of-band edits must surface as drift),
// while ciphertext echoes carry no usable value, so the known plaintext from
// state is kept to avoid spurious drift.
func TestDeploymentSettingsEnvVarRefresh(t *testing.T) {
	prev := &DeploymentSettingsOperationContext{
		EnvironmentVariables: map[string]string{
			"REGION": "us-west-2",
			"TOKEN":  "hunter2",
		},
	}

	t.Run("plaintext echo surfaces out-of-band changes", func(t *testing.T) {
		// A non-secret env var comes back as a bare JSON string holding the
		// *current* cloud-side value.
		var v pulumiapi.SecretValue
		require.NoError(t, json.Unmarshal([]byte(`"eu-central-1"`), &v))

		got := operationContextFromAPI(&pulumiapi.OperationContext{
			EnvironmentVariables: map[string]pulumiapi.SecretValue{"REGION": v},
		}, prev)
		assert.Equal(t, map[string]string{"REGION": "eu-central-1"}, got.EnvironmentVariables)
	})

	t.Run("ciphertext echo keeps known plaintext", func(t *testing.T) {
		// A secret env var comes back as {"secret": "<ciphertext>"} — no
		// plaintext, so the value previously recorded in state is authoritative.
		var v pulumiapi.SecretValue
		require.NoError(t, json.Unmarshal([]byte(`{"secret":"AAABBBccc=="}`), &v))

		got := operationContextFromAPI(&pulumiapi.OperationContext{
			EnvironmentVariables: map[string]pulumiapi.SecretValue{"TOKEN": v},
		}, prev)
		assert.Equal(t, map[string]string{"TOKEN": "hunter2"}, got.EnvironmentVariables)
	})
}

type deploymentSettingsClientMock struct {
	config.Client
	created *pulumiapi.DeploymentSettings
}

func (m *deploymentSettingsClientMock) CreateDeploymentSettings(
	_ context.Context, _ pulumiapi.StackIdentifier, ds pulumiapi.DeploymentSettings,
) (*pulumiapi.DeploymentSettings, error) {
	m.created = &ds
	return &ds, nil
}

// TestDeploymentSettingsSecretEnvironmentVariables asserts that env vars the
// user marked secret are sent to the Pulumi Cloud API with Secret: true, so the
// cloud stores them encrypted (the legacy implementation honored per-value
// secret markers this way).
//
// This test currently FAILS: infer decodes environmentVariables into
// map[string]string, stripping per-value secret markers before resource code
// runs, so every value is sent as non-secret plaintext. It is intentionally not
// skipped — it documents the desired behavior and will stay red until
// pulumi-go-provider can surface per-value secretness to typed inputs.
func TestDeploymentSettingsSecretEnvironmentVariables(t *testing.T) {
	mock := &deploymentSettingsClientMock{}
	ctx := config.WithMockClient(context.Background(), mock)

	prov, err := infer.NewProviderBuilder().
		WithNamespace("pulumi").
		WithResources(infer.Resource(&DeploymentSettings{})).
		Build()
	require.NoError(t, err)
	server, err := integration.NewServer(ctx, "pulumiservice", semver.MustParse("1.0.0"),
		integration.WithProvider(prov))
	require.NoError(t, err)

	urn := presource.URN("urn:pulumi:test::test::pulumiservice:index:DeploymentSettings::settings")
	inputs := property.NewMap(map[string]property.Value{
		"organization": property.New("org"),
		"project":      property.New("proj"),
		"stack":        property.New("stack"),
		"operationContext": property.New(property.NewMap(map[string]property.Value{
			"environmentVariables": property.New(property.NewMap(map[string]property.Value{
				"REGION": property.New("us-west-2"),
				"TOKEN":  property.New("hunter2").WithSecret(true),
			})),
		})),
	})

	checkResp, err := server.Check(p.CheckRequest{Urn: urn, Inputs: inputs})
	require.NoError(t, err)
	require.Empty(t, checkResp.Failures)

	_, err = server.Create(p.CreateRequest{Urn: urn, Properties: checkResp.Inputs})
	require.NoError(t, err)

	require.NotNil(t, mock.created)
	require.NotNil(t, mock.created.Operation)
	assert.Equal(t, map[string]pulumiapi.SecretValue{
		"REGION": {Value: "us-west-2"},
		"TOKEN":  {Value: "hunter2", Secret: true},
	}, mock.created.Operation.EnvironmentVariables)
}
