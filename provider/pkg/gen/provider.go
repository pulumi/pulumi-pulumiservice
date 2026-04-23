// Copyright 2016-2026, Pulumi Corporation.
//
// provider.go — fixed, hand-authored provider-identity blocks in the
// emitted schema: config variables, provider resource, and language
// metadata. These don't vary with the OpenAPI spec or the resource map,
// so they live here as literals rather than being derived.

package gen

import "encoding/json"

const (
	accessTokenDesc = "Access Token to authenticate with Pulumi Cloud."
	apiURLDesc      = "Optional override of Pulumi Cloud API endpoint."
	apiURLDefault   = "https://api.pulumi.com"
)

// providerVars returns a fresh copy of the provider-level property shapes
// (accessToken, apiUrl) each time it's called. Used for both the top-level
// `config` block and the `provider` resource's properties — they must not
// share a map, or future divergence (e.g., an output-only provider
// property) would mutate both sides at once.
func providerVars() map[string]pulumiProperty {
	return map[string]pulumiProperty{
		"accessToken": {
			Description: accessTokenDesc,
			Type:        "string",
			Secret:      true,
			DefaultInfo: &pulumiDefaultInfo{Environment: []string{"PULUMI_ACCESS_TOKEN"}},
		},
		"apiUrl": {
			Description: apiURLDesc,
			Type:        "string",
			Default:     apiURLDefault,
			DefaultInfo: &pulumiDefaultInfo{Environment: []string{"PULUMI_BACKEND_URL"}},
		},
	}
}

// providerConfig returns the schema's top-level `config` block. These are
// the variables that can appear in Pulumi stack config under
// `pulumiservice:<key>`; they map 1:1 to the provider resource's inputs.
func providerConfig() *pulumiConfig {
	return &pulumiConfig{Variables: providerVars()}
}

// providerResource returns the schema's `provider` block — the shape of
// the explicit Provider resource a user can construct (e.g.,
// `new pulumiservice.Provider("p", { accessToken })`). Mirrors config.
func providerResource() *pulumiResource {
	return &pulumiResource{
		Properties:      providerVars(),
		InputProperties: providerVars(),
	}
}

// languageBlocks returns the per-language codegen metadata embedded under
// `language.<lang>`. `pulumi package gen-sdk` reads these to decide package
// names, module paths, namespaces, and dependency pins. Values match what
// the v1 manual schema shipped, to preserve package identity across the
// v1→v2 boundary.
func languageBlocks() map[string]json.RawMessage {
	return map[string]json.RawMessage{
		"csharp": json.RawMessage(`{
  "namespaces": { "pulumiservice": "PulumiService" },
  "packageReferences": { "Pulumi": "3.*" },
  "respectSchemaVersion": true
}`),
		"go": json.RawMessage(`{
  "generateResourceContainerTypes": true,
  "importBasePath": "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice",
  "respectSchemaVersion": true
}`),
		"java": json.RawMessage(`{
  "basePackage": "",
  "buildFiles": "gradle",
  "dependencies": {
    "com.google.code.findbugs:jsr305": "3.0.2",
    "com.google.code.gson:gson": "2.8.9",
    "com.pulumi:pulumi": "1.16.2"
  },
  "gradleNexusPublishPluginVersion": "2.0.0",
  "gradleTest": ""
}`),
		"nodejs": json.RawMessage(`{
  "packageName": "@pulumi/pulumiservice",
  "dependencies": { "@pulumi/pulumi": "^3.0.0" },
  "respectSchemaVersion": true
}`),
		"python": json.RawMessage(`{
  "packageName": "pulumi_pulumiservice",
  "requires": { "pulumi": ">=3.0.0,<4.0.0" },
  "pyproject": { "enabled": true },
  "respectSchemaVersion": true
}`),
	}
}
