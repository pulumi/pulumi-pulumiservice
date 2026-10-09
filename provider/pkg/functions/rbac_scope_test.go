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

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
)

func allScopeValues() map[string][]string {
	values := map[string][]string{}
	for _, v := range RbacOrganizationScope("").Values() {
		values["RbacOrganizationScope"] = append(values["RbacOrganizationScope"], string(v.Value))
	}
	for _, v := range RbacStackScope("").Values() {
		values["RbacStackScope"] = append(values["RbacStackScope"], string(v.Value))
	}
	for _, v := range RbacEnvironmentScope("").Values() {
		values["RbacEnvironmentScope"] = append(values["RbacEnvironmentScope"], string(v.Value))
	}
	for _, v := range RbacInsightsAccountScope("").Values() {
		values["RbacInsightsAccountScope"] = append(values["RbacInsightsAccountScope"], string(v.Value))
	}
	return values
}

// TestRbacScopesAreValidInCloudSDK guards the provider-owned scope enums: the
// helpers validate scopes against the Pulumi Cloud SDK, so an enum value the
// SDK doesn't know would be offered to users and then rejected.
// On failure, bump github.com/pulumi/pulumi-cloud-sdk/go.
func TestRbacScopesAreValidInCloudSDK(t *testing.T) {
	t.Parallel()
	for enum, values := range allScopeValues() {
		assert.NotEmpty(t, values, enum)
		for _, v := range values {
			assert.Truef(t, apitype.RbacPermission(v).IsValid(),
				"%s %q is not a valid scope in the Pulumi Cloud SDK", enum, v)
		}
	}
}

// TestRbacScopesMissingFromEnums reports Cloud SDK scopes that no scope enum
// offers yet. It never fails: adding scopes to the public enums is a
// deliberate change made with scripts/gen-rbac-scopes.sh.
func TestRbacScopesMissingFromEnums(t *testing.T) {
	t.Parallel()
	known := map[string]bool{}
	for _, values := range allScopeValues() {
		for _, v := range values {
			known[v] = true
		}
	}
	for _, p := range apitype.RbacPermission("").AllValues() {
		if !known[string(p)] {
			t.Logf("Cloud SDK scope %q is not in any scope enum", p)
		}
	}
}
