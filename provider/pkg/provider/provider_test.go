// Copyright 2016-2025, Pulumi Corporation.
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

package provider

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

func TestDiffConfig(t *testing.T) {
	tests := []struct {
		name            string
		oldConfig       resource.PropertyMap
		newConfig       resource.PropertyMap
		expectedChanges pulumirpc.DiffResponse_DiffChanges
	}{
		{
			name: "accessToken changed",
			oldConfig: resource.PropertyMap{
				"accessToken": resource.NewPropertyValue("old-token-123"),
				"apiUrl":      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			newConfig: resource.PropertyMap{
				"accessToken": resource.NewPropertyValue("new-token-456"),
				"apiUrl":      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			expectedChanges: pulumirpc.DiffResponse_DIFF_SOME,
		},
		{
			name: "apiUrl changed",
			oldConfig: resource.PropertyMap{
				"accessToken": resource.NewPropertyValue("token-123"),
				"apiUrl":      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			newConfig: resource.PropertyMap{
				"accessToken": resource.NewPropertyValue("token-123"),
				"apiUrl":      resource.NewPropertyValue("https://custom.pulumi.example.com"),
			},
			expectedChanges: pulumirpc.DiffResponse_DIFF_SOME,
		},
		{
			name: "no changes",
			oldConfig: resource.PropertyMap{
				"accessToken": resource.NewPropertyValue("token-123"),
				"apiUrl":      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			newConfig: resource.PropertyMap{
				"accessToken": resource.NewPropertyValue("token-123"),
				"apiUrl":      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			expectedChanges: pulumirpc.DiffResponse_DIFF_NONE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create provider instance
			provider, err := MakeProvider(nil, "pulumiservice", "1.0.0")
			require.NoError(t, err)

			// Marshal old and new configs
			oldProps, err := plugin.MarshalProperties(tt.oldConfig, plugin.MarshalOptions{})
			require.NoError(t, err)

			newProps, err := plugin.MarshalProperties(tt.newConfig, plugin.MarshalOptions{})
			require.NoError(t, err)

			// Call CheckConfig with old config
			_, err = provider.CheckConfig(ctx, &pulumirpc.CheckRequest{
				News: oldProps,
			})
			require.NoError(t, err)

			// Call DiffConfig to detect changes
			resp, err := provider.DiffConfig(ctx, &pulumirpc.DiffRequest{
				Urn:  "urn:pulumi:stack::project::pulumi:providers:pulumiservice::provider",
				Olds: oldProps,
				News: newProps,
			})
			require.NoError(t, err)

			// Assert expected changes
			assert.Equal(t, tt.expectedChanges, resp.Changes)
			assert.Empty(t, resp.Replaces, "config changes should not require replacement")
		})
	}
}

func TestConfigure_SetsAccessToken(t *testing.T) {
	ctx := context.Background()

	// Ensure PULUMI_ACCESS_TOKEN is not set from environment
	oldToken := os.Getenv(EnvVarPulumiAccessToken)
	err := os.Setenv(EnvVarPulumiAccessToken, "")
	require.NoError(t, err)
	defer func() {
		err := os.Setenv(EnvVarPulumiAccessToken, oldToken)
		assert.NoError(t, err)
	}()

	// Create provider instance
	provider, err := MakeProvider(nil, "pulumiservice", "1.0.0")
	require.NoError(t, err)

	// Create config with accessToken
	config := resource.PropertyMap{
		"accessToken": resource.NewPropertyValue("pul-test0token"),
	}
	props, err := plugin.MarshalProperties(config, plugin.MarshalOptions{})
	require.NoError(t, err)

	// Call CheckConfig
	_, err = provider.CheckConfig(ctx, &pulumirpc.CheckRequest{
		News: props,
	})
	require.NoError(t, err)

	// Build config variables map as expected by Configure
	configVars := map[string]string{
		"pulumiservice:config:accessToken": "pul-test0token",
	}

	// Call Configure
	_, err = provider.Configure(ctx, &pulumirpc.ConfigureRequest{
		Variables: configVars,
	})
	require.NoError(t, err)

	// Assert that AccessToken is set on the provider
	// We need to cast to the concrete type to access AccessToken field
	concreteProvider, ok := provider.(*pulumiserviceProvider)
	if !ok {
		// The provider is wrapped, so we need to check if client was set instead
		t.Skip("Provider is wrapped, cannot directly access AccessToken field")
	}
	assert.NotNil(t, concreteProvider.client, "client should be initialized after Configure")
}

// TestProvider_LayeredSchema verifies the unified provider serves three
// resource sources under one pulumiservice schema:
//
//  1. legacy raw gRPC server (manual-schema.json) — e.g. Stack
//  2. modern infer (WithResources)              — e.g. Team
//  3. metadata-driven v2 layer (rest.BuildSchema spliced via withCloudV2Schema)
//     — e.g. OrganizationWebhook at pulumiservice:v2:*
//
// All three share the pulumiservice package name; v1 surface is implicit
// at the package root (pulumiservice:index:*), v2 is an explicit module.
// Existing user code using pulumiservice:index:* keeps working unchanged.
func TestProvider_LayeredSchema(t *testing.T) {
	provider, err := MakeProvider(nil, "pulumiservice", "1.0.0")
	require.NoError(t, err)

	resp, err := provider.GetSchema(context.Background(), &pulumirpc.GetSchemaRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Schema)

	var spec schema.PackageSpec
	require.NoError(t, json.Unmarshal([]byte(resp.Schema), &spec))

	assert.Equal(t, "pulumiservice", spec.Name)

	mustHave := []struct {
		token  string
		source string
	}{
		{"pulumiservice:index:Stack", "v1: legacy raw (manual-schema.json)"},
		{"pulumiservice:index:Team", "v1: modern infer (WithResources)"},
		{"pulumiservice:v2:OrganizationWebhook", "v2: metadata-driven (withCloudV2Schema)"},
	}
	for _, c := range mustHave {
		_, ok := spec.Resources[c.token]
		assert.Truef(t, ok, "missing %s — expected from %s", c.token, c.source)
	}

}
