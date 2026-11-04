// Copyright 2016-2020, Pulumi Corporation.
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
	"fmt"

	_ "embed" // For the README.

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	mw "github.com/pulumi/pulumi-go-provider/middleware"
	ctxmw "github.com/pulumi/pulumi-go-provider/middleware/context"
	"github.com/pulumi/pulumi-go-provider/middleware/dispatch"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/cloud"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/functions"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/resources"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

const (
	respectSchemaVersionKey = "respectSchemaVersion"
	readmeKey               = "readme"
)

//go:embed README.md
var readme string

// MakeProvider builds the Pulumi Cloud provider in two layers:
//
//  1. dispatch.Wrap serves the api resources (pulumiservice:api:*), generated
//     from the Pulumi Cloud OpenAPI spec and provider/pkg/cloud/metadata.json.
//  2. infer.NewProviderBuilder serves the pulumiservice:index:* resources and
//     functions, handles configuration, and sets the package metadata.
func MakeProvider(host *provider.HostClient, name, version string) (pulumirpc.ResourceProviderServer, error) {
	customs := map[tokens.Type]mw.CustomResource{}
	for tok, h := range rest.Resources(cloud.Spec(), cloud.Metadata()) {
		customs[tokens.Type(tok)] = h
	}
	api := dispatch.Wrap(p.Provider{
		GetSchema: cloudAPISchema(cloud.Spec(), cloud.Metadata(), name),
	}, dispatch.Options{Customs: customs})
	// The api resources send requests through the transport that
	// config.Config builds during Configure.
	api = ctxmw.Wrap(api, func(ctx context.Context) context.Context {
		if t := config.GetTransport(ctx); t != nil {
			return rest.WithTransport(ctx, t)
		}
		return ctx
	})

	provider, err := infer.NewProviderBuilder().
		WithDisplayName("Pulumi Cloud").
		WithDescription("A native Pulumi package for creating and managing Pulumi Cloud constructs.").
		WithHomepage("https://pulumi.com").
		WithRepository("https://github.com/pulumi/pulumi-pulumiservice").
		WithKeywords("pulumi", "kind/native", "category/infrastructure").
		WithLicense("Apache-2.0").
		WithPublisher("Pulumi").
		WithNamespace("pulumi").
		WithWrapped(api).
		WithResources(inferResources()...).
		WithFunctions(inferFunctions()...).
		WithModuleMap(inferModuleMap).
		WithConfig(infer.Config(&config.Config{})).
		WithLanguageMap(map[string]any{
			"csharp": map[string]any{
				"namespaces": map[string]any{
					"pulumiservice": "PulumiService",
					"api":           "Api",
				},
				"packageReferences": map[string]any{
					"Pulumi": "3.*",
				},
				respectSchemaVersionKey: true,
			},
			"go": map[string]any{
				"generateResourceContainerTypes": true,
				"importBasePath":                 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice",
				respectSchemaVersionKey:          true,
			},
			"java": map[string]any{
				"basePackage": "",
				"buildFiles":  "gradle",
				"dependencies": map[string]any{
					"com.google.code.findbugs:jsr305": "3.0.2",
					"com.google.code.gson:gson":       "2.8.9",
					"com.pulumi:pulumi":               "1.16.2",
				},
				"gradleNexusPublishPluginVersion": "2.0.0",
				"gradleTest":                      "",
				readmeKey:                         readme,
			},
			"nodejs": map[string]any{
				"packageName": "@pulumi/pulumiservice",
				"dependencies": map[string]any{
					"@pulumi/pulumi": "^3.0.0",
				},
				respectSchemaVersionKey: true,
				readmeKey:               readme,
			},
			"python": map[string]any{
				"packageName": "pulumi_pulumiservice",
				"requires": map[string]any{
					// 3.235.0 is the first runtime that preserves
					// `__`-prefixed input keys (pulumi/pulumi#22834),
					// which OrganizationRole.permissions relies on for
					// the `__type` discriminator at every level of the
					// descriptor tree.
					"pulumi": ">=3.235.0,<4.0.0",
				},
				"pyproject": map[string]any{
					"enabled": true,
				},
				respectSchemaVersionKey: true,
				readmeKey:               readme,
			},
		}).Build()
	if err != nil {
		return nil, err
	}
	provider = withFunctionDeprecations(provider, functions.DeprecatedFunctions)
	return p.RawServer(name, version, resources.WithEnvironmentCompat(provider))(host)
}

// cloudAPISchema returns the schema of the api resources. infer merges it
// with the schema of the resources it serves.
func cloudAPISchema(
	spec *rest.Spec, metadata *rest.Metadata, pkg string,
) func(context.Context, p.GetSchemaRequest) (p.GetSchemaResponse, error) {
	return func(context.Context, p.GetSchemaRequest) (p.GetSchemaResponse, error) {
		fragment, err := rest.BuildSchema(spec, metadata, pkg)
		if err != nil {
			return p.GetSchemaResponse{}, fmt.Errorf("cloudAPISchema: %w", err)
		}
		out, err := json.Marshal(fragment)
		if err != nil {
			return p.GetSchemaResponse{}, fmt.Errorf("cloudAPISchema: encode schema: %w", err)
		}
		return p.GetSchemaResponse{Schema: string(out)}, nil
	}
}

// withFunctionDeprecations sets each function's deprecation message, which
// pulumi-go-provider doesn't emit for functions yet
// (pulumi/pulumi-go-provider#636). It wraps the built provider because infer
// adds its functions to the schema outside every inner layer.
func withFunctionDeprecations(prov p.Provider, messages map[string]string) p.Provider {
	inner := prov.GetSchema
	prov.GetSchema = func(ctx context.Context, req p.GetSchemaRequest) (p.GetSchemaResponse, error) {
		resp, err := inner(ctx, req)
		if err != nil {
			return resp, err
		}
		var spec schema.PackageSpec
		if err := json.Unmarshal([]byte(resp.Schema), &spec); err != nil {
			return resp, fmt.Errorf("withFunctionDeprecations: parse schema: %w", err)
		}
		for token, message := range messages {
			fn, ok := spec.Functions[token]
			if !ok {
				return resp, fmt.Errorf("withFunctionDeprecations: no function %q in the schema", token)
			}
			fn.DeprecationMessage = message
			spec.Functions[token] = fn
		}
		out, err := json.Marshal(spec)
		if err != nil {
			return resp, fmt.Errorf("withFunctionDeprecations: re-encode schema: %w", err)
		}
		resp.Schema = string(out)
		return resp, nil
	}
	return prov
}

// inferModuleMap puts the Go packages that hold the infer resources and
// functions in the index module.
var inferModuleMap = map[tokens.ModuleName]tokens.ModuleName{
	"resources": "index",
	"functions": "index",
}

func inferResources() []infer.InferredResource {
	return []infer.InferredResource{
		infer.Resource(&resources.AccessToken{}),
		infer.Resource(&resources.DeploymentSettings{}),
		infer.Resource(&resources.AgentPool{}),
		infer.Resource(&resources.ApprovalRule{}),
		infer.Resource(&resources.DeploymentSchedule{}),
		infer.Resource(&resources.DriftSchedule{}),
		infer.Resource(&resources.Environment{}),
		infer.Resource(&resources.EnvironmentRotationSchedule{}),
		infer.Resource(&resources.EnvironmentVersionTag{}),
		infer.Resource(&resources.InsightsAccount{}),
		infer.Resource(&resources.OidcIssuer{}),
		infer.Resource(&resources.OrgAccessToken{}),
		infer.Resource(&resources.OrganizationDefaultRole{}),
		infer.Resource(&resources.OrganizationMember{}),
		infer.Resource(&resources.OrganizationMemberEntityAccess{}),
		infer.Resource(&resources.OrganizationRole{}),
		infer.Resource(&resources.PolicyGroup{}),
		infer.Resource(&resources.PolicyPack{}),
		infer.Resource(&resources.Stack{}),
		infer.Resource(&resources.StackTag{}),
		infer.Resource(&resources.StackTags{}),
		infer.Resource(&resources.TTLSchedule{}),
		infer.Resource(&resources.Team{}),
		infer.Resource(&resources.TeamAccessToken{}),
		infer.Resource(&resources.TeamEnvironmentPermission{}),
		infer.Resource(&resources.TeamRoleAssignment{}),
		infer.Resource(&resources.TeamStackPermission{}),
		infer.Resource(&resources.TemplateSource{}),
		infer.Resource(&resources.Webhook{}),
	}
}

func inferFunctions() []infer.InferredFunction {
	return []infer.InferredFunction{
		infer.Function(&functions.BuildAccountMatchRuleFunction{}),
		infer.Function(&functions.BuildAllowPermissionsFunction{}),
		infer.Function(&functions.BuildEnvironmentScopedPermissionsFunction{}),
		infer.Function(&functions.BuildInsightsAccountScopedPermissionsFunction{}),
		infer.Function(&functions.BuildRolePermissionsFunction{}),
		infer.Function(&functions.BuildStackMatchRuleFunction{}),
		infer.Function(&functions.BuildStackScopedPermissionsFunction{}),
		infer.Function(&functions.GetCurrentUserFunction{}),
		infer.Function(&functions.GetEnvironmentFunction{}),
		infer.Function(&functions.GetInsightsAccountFunction{}),
		infer.Function(&functions.GetInsightsAccountsFunction{}),
		infer.Function(&functions.GetOrganizationMemberFunction{}),
		infer.Function(&functions.GetOrganizationMembersFunction{}),
		infer.Function(&functions.GetOrganizationRoleFunction{}),
		infer.Function(&functions.GetOrganizationRoleScopesFunction{}),
		infer.Function(&functions.GetOrganizationRolesFunction{}),
		infer.Function(&functions.GetPolicyPackFunction{}),
		infer.Function(&functions.GetPolicyPacksFunction{}),
		infer.Function(&functions.GetStackFunction{}),
	}
}
