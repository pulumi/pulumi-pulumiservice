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
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/middleware/rpc"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	inferResources "github.com/pulumi/pulumi-pulumiservice/provider/pkg/infer"
)

// MakeHybridProvider creates a provider that supports both manual (legacy) and infer-based resources.
//
// This hybrid approach allows us to gradually migrate resources from manual implementation to infer
// without breaking existing functionality. As resources are migrated, they will be removed from the
// manual provider and added to the infer provider.
//
// The migration plan:
// - Phase 0: Foundation setup (Go 1.24, pulumi-go-provider v1.1.2+)
// - Phase 1: POC with 3 simple resources (StackTag, OrgAccessToken, AgentPool)
// - Phase 2-4: Migrate remaining resources in complexity order
// - Phase 5: Migrate data sources (invoke functions)
// - Phase 6: Remove manual implementation entirely
//
// See Convert-to-infer.md for the complete migration plan.
func MakeHybridProvider(name, version, schema string) p.Provider {
	manualProvider := &pulumiserviceProvider{
		name:    name,
		schema:  mustSetSchemaVersion(schema, version),
		version: version,
	}

	inferOpts := buildInferOptions()

	// Wrap the manual provider with rpc.Provider to convert it to p.Provider interface
	wrappedManual := rpc.Provider(manualProvider)

	// Wrap with infer to add infer-based resources (like StackTag)
	// infer.Wrap will handle infer resources and delegate unknown resources to wrappedManual
	return infer.Wrap(wrappedManual, inferOpts)
}

// buildInferOptions constructs the infer.Options for resources that have been migrated to use infer.
//
// As resources are migrated, they should be registered here using infer.Resource().
// Once all resources are migrated, we can remove the manual provider entirely and use
// infer.NewProviderBuilder() directly in main.go.
func buildInferOptions() infer.Options {
	return infer.Options{
		Resources: []infer.InferredResource{
			// Phase 1 POC resources
			infer.Resource(&inferResources.StackTag{}),

			// Remaining resources to be migrated:
			// infer.Resource(&inferResources.OrgAccessToken{}),
			// infer.Resource(&inferResources.AgentPool{}),
		},
		Components: []infer.InferredComponent{
			// TODO: Add components if needed
		},
		Functions: []infer.InferredFunction{
			// TODO: Add invoke functions here as they are migrated
			// Phase 5 data sources (will be added later):
			// infer.Function[GetPolicyPacks](),
			// infer.Function[GetPolicyPack](),
		},
		// Config is specified here to enable proper client injection via context
		// infer.Config[*config.Config](&config.Config{}) will automatically call Config.Configure() and
		// inject the config into context, making it available to resources via GetConfig(ctx)
		Config: infer.Config[*config.Config](&config.Config{}),
	}
}
