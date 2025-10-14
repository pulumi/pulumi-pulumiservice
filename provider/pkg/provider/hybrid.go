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
	// Create a wrapped manual provider that injects clients into context
	manualProvider := &hybridManualProvider{
		pulumiserviceProvider: &pulumiserviceProvider{
			name:    name,
			schema:  mustSetSchemaVersion(schema, version),
			version: version,
		},
	}

	// Build the infer provider options for migrated resources
	inferOpts := buildInferOptions()

	// Wrap the manual provider with RPC middleware to make it compatible with p.Provider
	wrappedManual := rpc.Provider(manualProvider)

	// Combine both providers using infer.Wrap
	// The infer provider will handle its resources, and delegate to manual provider for others
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
		// Note: Config is not specified here - we're using the manual provider's configuration
		// Once all resources are migrated, we can define a proper infer Config
	}
}
