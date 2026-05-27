// Copyright 2016-2026, Pulumi Corporation.
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestPolicyPackID_RoundTrip(t *testing.T) {
	id := policyPackID("acme", "guard", "1.2.3")
	assert.Equal(t, "acme/guard/1.2.3", id)

	org, name, tag, err := splitPolicyPackID(id)
	require.NoError(t, err)
	assert.Equal(t, "acme", org)
	assert.Equal(t, "guard", name)
	assert.Equal(t, "1.2.3", tag)
}

func TestSplitPolicyPackID_InvalidShape(t *testing.T) {
	for _, id := range []string{"", "only-one", "two/parts", "four/parts/here/extra"} {
		_, _, _, err := splitPolicyPackID(id)
		assert.Errorf(t, err, "expected error for id %q", id)
	}
}

func TestNormalizeConfigSchema(t *testing.T) {
	// nil and empty pass through untouched
	assert.Nil(t, normalizeConfigSchema(nil))
	empty := map[string]any{}
	assert.Equal(t, empty, normalizeConfigSchema(empty))

	// existing "type" is preserved
	withType := map[string]any{"type": "string"}
	got := normalizeConfigSchema(withType)
	assert.Equal(t, "string", got["type"])

	// missing "type" gets defaulted to object without mutating the input
	in := map[string]any{"properties": map[string]any{"x": map[string]any{"type": "number"}}}
	out := normalizeConfigSchema(in)
	assert.Equal(t, "object", out["type"])
	_, hadType := in["type"]
	assert.False(t, hadType, "input must not be mutated")
}

func TestToAPIPolicies(t *testing.T) {
	in := []PolicyPackPolicyInput{
		{
			Name:             "no-secrets",
			DisplayName:      "No Secrets",
			Description:      "block secret literals",
			EnforcementLevel: "mandatory",
			Message:          "remove the secret",
			ConfigSchema:     map[string]any{"type": "object"},
		},
		{Name: "minimal"},
	}
	got := toAPIPolicies(in)
	require.Len(t, got, 2)
	assert.Equal(t, pulumiapi.Policy{
		Name:             "no-secrets",
		DisplayName:      "No Secrets",
		Description:      "block secret literals",
		EnforcementLevel: "mandatory",
		Message:          "remove the secret",
		ConfigSchema:     map[string]any{"type": "object"},
	}, got[0])
	assert.Equal(t, "minimal", got[1].Name)
}

func TestPoliciesEqual(t *testing.T) {
	a := []PolicyPackPolicyInput{{Name: "a", EnforcementLevel: "advisory"}}
	b := []PolicyPackPolicyInput{{Name: "a", EnforcementLevel: "advisory"}}
	assert.True(t, policiesEqual(a, b))

	b[0].EnforcementLevel = "mandatory"
	assert.False(t, policiesEqual(a, b))

	assert.True(t, policiesEqual(nil, nil))
	assert.False(t, policiesEqual(a, nil))
}

func TestConvertAnalyzerConfigSchema(t *testing.T) {
	assert.Nil(t, convertAnalyzerConfigSchema(nil))

	got := convertAnalyzerConfigSchema(&plugin.AnalyzerPolicyConfigSchema{
		Properties: map[string]plugin.JSONSchema{
			"threshold": {"type": "number"},
		},
		Required: []string{"threshold"},
	})
	assert.Equal(t, "object", got["type"])
	assert.Equal(t, []string{"threshold"}, got["required"])

	props, ok := got["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, props, "threshold")

	// Empty schema still gets a type but no properties/required keys.
	empty := convertAnalyzerConfigSchema(&plugin.AnalyzerPolicyConfigSchema{})
	assert.Equal(t, map[string]any{"type": "object"}, empty)
}
