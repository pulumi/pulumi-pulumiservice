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

package functions

import "github.com/pulumi/pulumi-go-provider/infer"

// Pulumi Cloud RBAC scopes, the most granular access rights, written
// `object:action` (for example `stack:read`). Each scope applies to one entity
// type, so there is one enum per type; a rule for one type can't name another
// type's scopes.
//
// The values are provider-owned: they are generated from the Pulumi Cloud
// scope catalog by scripts/gen-rbac-scopes.sh and checked in, so the public
// enums change only when someone regenerates and reviews them. Tests check
// every value against the Pulumi Cloud SDK.

// RbacOrganizationScope is a scope that applies organization-wide.
type RbacOrganizationScope string

func (RbacOrganizationScope) Values() []infer.EnumValue[RbacOrganizationScope] {
	return rbacOrganizationScopeValues
}

// RbacStackScope is a scope that applies to stacks.
type RbacStackScope string

func (RbacStackScope) Values() []infer.EnumValue[RbacStackScope] {
	return rbacStackScopeValues
}

// RbacEnvironmentScope is a scope that applies to ESC environments.
type RbacEnvironmentScope string

func (RbacEnvironmentScope) Values() []infer.EnumValue[RbacEnvironmentScope] {
	return rbacEnvironmentScopeValues
}

// RbacInsightsAccountScope is a scope that applies to Insights accounts.
type RbacInsightsAccountScope string

func (RbacInsightsAccountScope) Values() []infer.EnumValue[RbacInsightsAccountScope] {
	return rbacInsightsAccountScopeValues
}

// checkScopes returns the scopes as strings, or an error naming the first one
// that isn't a member of the enum T. The schema already restricts the values
// in typed SDKs; this also covers YAML and any caller that bypasses the enum.
func checkScopes[T ~string](scopes []T, values []infer.EnumValue[T], kind string) ([]string, error) {
	known := make(map[T]bool, len(values))
	for _, v := range values {
		known[v.Value] = true
	}
	out := make([]string, len(scopes))
	for i, s := range scopes {
		if !known[s] {
			return nil, &scopeKindError{scope: string(s), kind: kind}
		}
		out[i] = string(s)
	}
	return out, nil
}

type scopeKindError struct{ scope, kind string }

func (e *scopeKindError) Error() string {
	return "`" + e.scope + "` is not " + e.kind + " scope; each rule list accepts only scopes for its " +
		"own entity type (see getOrganizationRoleScopes)"
}
