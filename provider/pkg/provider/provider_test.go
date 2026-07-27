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
				accessTokenKey: resource.NewPropertyValue("old-token-123"),
				apiURLKey:      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			newConfig: resource.PropertyMap{
				accessTokenKey: resource.NewPropertyValue("new-token-456"),
				apiURLKey:      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			expectedChanges: pulumirpc.DiffResponse_DIFF_SOME,
		},
		{
			name: "apiUrl changed",
			oldConfig: resource.PropertyMap{
				accessTokenKey: resource.NewPropertyValue("token-123"),
				apiURLKey:      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			newConfig: resource.PropertyMap{
				accessTokenKey: resource.NewPropertyValue("token-123"),
				apiURLKey:      resource.NewPropertyValue("https://custom.pulumi.example.com"),
			},
			expectedChanges: pulumirpc.DiffResponse_DIFF_SOME,
		},
		{
			name: "no changes",
			oldConfig: resource.PropertyMap{
				accessTokenKey: resource.NewPropertyValue("token-123"),
				apiURLKey:      resource.NewPropertyValue("https://api.pulumi.com"),
			},
			newConfig: resource.PropertyMap{
				accessTokenKey: resource.NewPropertyValue("token-123"),
				apiURLKey:      resource.NewPropertyValue("https://api.pulumi.com"),
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
		accessTokenKey: resource.NewPropertyValue("pul-test0token"),
	}
	props, err := plugin.MarshalProperties(config, plugin.MarshalOptions{})
	require.NoError(t, err)

	// Call CheckConfig
	_, err = provider.CheckConfig(ctx, &pulumirpc.CheckRequest{
		News: props,
	})
	require.NoError(t, err)

	// Build config args as the engine sends them: plain keys, secrets kept.
	configArgs, err := plugin.MarshalProperties(resource.PropertyMap{
		accessTokenKey: resource.MakeSecret(resource.NewStringProperty("pul-test0token")),
	}, plugin.MarshalOptions{KeepSecrets: true})
	require.NoError(t, err)

	// Configure the raw provider directly: MakeProvider wraps it, which
	// hides the client field this test asserts on.
	raw := &pulumiserviceProvider{}
	_, err = raw.Configure(ctx, &pulumirpc.ConfigureRequest{
		Args: configArgs,
	})
	require.NoError(t, err)

	assert.NotNil(t, raw.client, "client should be initialized after Configure")
}

// TestProvider_LayeredSchema verifies the unified provider serves three
// resource sources under one pulumiservice schema:
//
//  1. legacy raw gRPC server (manual-schema.json) — e.g. Stack
//  2. modern infer (WithResources)              — e.g. Team
//  3. metadata-driven api layer (rest.BuildSchema spliced via withCloudApiSchema)
//     — e.g. OrganizationWebhook at pulumiservice:api:*
//
// All three share the pulumiservice package name; v0 surface is implicit
// at the package root (pulumiservice:index:*), api is an explicit module.
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
		{"pulumiservice:index:Stack", "v0: legacy raw (manual-schema.json)"},
		{"pulumiservice:index:Team", "v0: modern infer (WithResources)"},
		{"pulumiservice:api:OrganizationWebhook", "api: metadata-driven (withCloudApiSchema)"},
	}
	for _, c := range mustHave {
		_, ok := spec.Resources[c.token]
		assert.Truef(t, ok, "missing %s — expected from %s", c.token, c.source)
	}

}

// getPolicyPacks' result must be a named type, not an inline object.
//
// Pulumi's schema format only supports $ref in array items — an inline object is
// flattened to a bare `{"type": "object"}` during schema generation, which is why
// every SDK used to render policyPacks as an untyped string map. If this reverts,
// source and publisher silently stop reaching typed SDKs even though the provider
// still puts them on the wire.
func TestProvider_PolicyPackSummaryIsANamedType(t *testing.T) {
	provider, err := MakeProvider(nil, "pulumiservice", "1.0.0")
	require.NoError(t, err)

	resp, err := provider.GetSchema(context.Background(), &pulumirpc.GetSchemaRequest{})
	require.NoError(t, err)

	var spec schema.PackageSpec
	require.NoError(t, json.Unmarshal([]byte(resp.Schema), &spec))

	// FunctionSpec.UnmarshalJSON parses `outputs` into ReturnType, leaving the
	// deprecated Outputs field nil.
	fn, ok := spec.Functions["pulumiservice:index:getPolicyPacks"]
	require.True(t, ok, "getPolicyPacks should be present")
	require.NotNil(t, fn.ReturnType)
	require.NotNil(t, fn.ReturnType.ObjectTypeSpec)

	packs, ok := fn.ReturnType.ObjectTypeSpec.Properties[policyPacksKey]
	require.True(t, ok, "policyPacks output should be present")
	require.NotNil(t, packs.Items, "policyPacks items should survive schema generation")
	assert.Equal(t, "#/types/pulumiservice:index:PolicyPackSummary", packs.Items.Ref)

	summary, ok := spec.Types["pulumiservice:index:PolicyPackSummary"]
	require.True(t, ok, "PolicyPackSummary type should be present")

	for _, field := range []string{nameKey, displayNameKey, versionsKey, versionTagsKey, sourceKey, publisherKey} {
		_, ok := summary.Properties[field]
		assert.Truef(t, ok, "PolicyPackSummary should declare %q", field)
	}

	// Optional so a degraded lookup is distinguishable from real data in typed SDKs.
	assert.NotContains(t, summary.Required, sourceKey)
	assert.NotContains(t, summary.Required, publisherKey)

	single, ok := spec.Functions["pulumiservice:index:getPolicyPack"]
	require.True(t, ok, "getPolicyPack should be present")
	require.NotNil(t, single.ReturnType)
	require.NotNil(t, single.ReturnType.ObjectTypeSpec)
	for _, field := range []string{sourceKey, publisherKey} {
		_, ok := single.ReturnType.ObjectTypeSpec.Properties[field]
		assert.Truef(t, ok, "getPolicyPack should declare %q", field)
		assert.NotContains(t, single.ReturnType.ObjectTypeSpec.Required, field)
	}
}
