// Copyright 2026, Pulumi Corporation.
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

package resources

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testOrgSetID   = "org-set"
	testStackSetID = "stack-set"
	testEnvSetID   = "env-set"
	testStackUUID  = "20dede8f-f399-4deb-bd19-174410f209c6"
	testEnvUUID    = "58dd45c3-23fd-40b0-9b7e-3db51e09877b"
)

// testSetTypes stands in for the organization's permission sets.
func testSetTypes(id string) (RbacResourceType, error) {
	switch id {
	case testStackSetID:
		return RbacResourceTypeStack, nil
	case testEnvSetID:
		return RbacResourceTypeEnvironment, nil
	default:
		return RbacResourceTypeGlobal, nil
	}
}

// consoleRole mirrors the shapes the Pulumi Cloud console writes, captured
// from roles created in the console: org-level access and "all stacks" share
// one unconditional Compose, then one Condition per entity or tag rule.
const consoleRole = `{
  "__type": "PermissionDescriptorGroup",
  "entries": [
    {"__type": "PermissionDescriptorCompose", "permissionDescriptors": ["org-set", "stack-set"]},
    {"__type": "PermissionDescriptorCondition",
     "condition": {"__type": "PermissionExpressionEqual",
       "left": {"__type": "PermissionExpressionStack"},
       "right": {"__type": "PermissionLiteralExpressionStack", "identity": "20dede8f-f399-4deb-bd19-174410f209c6"}},
     "subNode": {"__type": "PermissionDescriptorCompose", "permissionDescriptors": ["stack-set"]}},
    {"__type": "PermissionDescriptorCondition",
     "condition": {"__type": "PermissionExpressionEqual",
       "left": {"__type": "PermissionExpressionEnvironment"},
       "right": {"__type": "PermissionLiteralExpressionEnvironment",
         "identity": "58dd45c3-23fd-40b0-9b7e-3db51e09877b"}},
     "subNode": {"__type": "PermissionDescriptorCompose", "permissionDescriptors": ["env-set"]}},
    {"__type": "PermissionDescriptorCondition",
     "condition": {"__type": "PermissionExpressionAnd",
       "left": {"__type": "PermissionExpressionAnd",
         "left": {"__type": "PermissionExpressionEqual",
           "left": {"__type": "PermissionExpressionTag",
             "context": {"__type": "PermissionExpressionStack"}, "key": "Owner"},
           "right": {"__type": "PermissionLiteralExpressionString", "value": "DevTeam"}},
         "right": {"__type": "PermissionExpressionNot",
           "node": {"__type": "PermissionExpressionEqual",
             "left": {"__type": "PermissionExpressionTag",
               "context": {"__type": "PermissionExpressionStack"}, "key": "env"},
             "right": {"__type": "PermissionLiteralExpressionString", "value": "prod"}}}},
       "right": {"__type": "PermissionExpressionHasTag",
         "context": {"__type": "PermissionExpressionStack"}, "key": "team"}},
     "subNode": {"__type": "PermissionDescriptorCompose", "permissionDescriptors": ["stack-set"]}}
  ]
}`

func consoleRoleModel() RbacRoleCore {
	return RbacRoleCore{
		OrganizationPermissionSetIds: []string{testOrgSetID},
		EntityRules: []RbacEntityRule{
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{All: pointerTo(true)}},
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{Id: pointerTo(testStackUUID)}},
			{PermissionSetIds: []string{testEnvSetID}, Environment: &RbacEntitySelector{Id: pointerTo(testEnvUUID)}},
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{Tags: []RbacTagCondition{
				{Key: "Owner", Value: pointerTo("DevTeam")},
				{Key: "env", Value: pointerTo("prod"), Operator: pointerTo(RbacTagOperatorNotEquals)},
				{Key: "team"},
			}}},
		},
	}
}

func TestBuildPolicyDetailsMatchesConsole(t *testing.T) {
	t.Parallel()
	details, err := buildPolicyDetails(consoleRoleModel())
	require.NoError(t, err)
	got, err := json.Marshal(details)
	require.NoError(t, err)
	assert.JSONEq(t, consoleRole, string(got))
}

func TestParsePolicyDetailsRoundTrip(t *testing.T) {
	t.Parallel()
	details := mustParseDescriptor(t, consoleRole)
	org, rules, err := parsePolicyDetails(details, testSetTypes)
	require.NoError(t, err)

	want := consoleRoleModel()
	assert.Equal(t, want.OrganizationPermissionSetIds, org)
	// Tag values come back as explicit pointers and default operators as nil;
	// compare through the canonical form, which normalizes both.
	assert.Equal(t, canonicalRoleKey(want.OrganizationPermissionSetIds, want.EntityRules),
		canonicalRoleKey(org, rules))
	require.Len(t, rules, 4)
	assert.Equal(t, pointerTo(true), rules[0].Stack.All, "all-stacks rule comes from the unconditional Compose")
}

func TestParsePolicyDetailsEnvironmentTagRule(t *testing.T) {
	t.Parallel()
	details := mustParseDescriptor(t, `{"__type":"PermissionDescriptorGroup","entries":[
	  {"__type":"PermissionDescriptorCondition",
	   "condition":{"__type":"PermissionExpressionNot","node":{"__type":"PermissionExpressionHasTag",
	     "context":{"__type":"PermissionExpressionEnvironment"},"key":"sensitive"}},
	   "subNode":{"__type":"PermissionDescriptorCompose","permissionDescriptors":["env-set"]}}]}`)
	_, rules, err := parsePolicyDetails(details, testSetTypes)
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.NotNil(t, rules[0].Environment)
	assert.Equal(t, []RbacTagCondition{{Key: "sensitive", Operator: pointerTo(RbacTagOperatorNotEquals)}},
		rules[0].Environment.Tags)
}

func TestParsePolicyDetailsUnrepresentable(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"raw allow": `{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`,
		"allow inside group": `{"__type":"PermissionDescriptorGroup","entries":[
		  {"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}]}`,
		"or condition": `{"__type":"PermissionDescriptorGroup","entries":[
		  {"__type":"PermissionDescriptorCondition",
		   "condition":{"__type":"PermissionExpressionOr",
		     "left":{"__type":"PermissionExpressionHasTag","context":{"__type":"PermissionExpressionStack"},"key":"a"},
		     "right":{"__type":"PermissionExpressionHasTag","context":{"__type":"PermissionExpressionStack"},"key":"b"}},
		   "subNode":{"__type":"PermissionDescriptorCompose","permissionDescriptors":["stack-set"]}}]}`,
		"mixed tag contexts": `{"__type":"PermissionDescriptorGroup","entries":[
		  {"__type":"PermissionDescriptorCondition",
		   "condition":{"__type":"PermissionExpressionAnd",
		     "left":{"__type":"PermissionExpressionHasTag","context":{"__type":"PermissionExpressionStack"},"key":"a"},
		     "right":{"__type":"PermissionExpressionHasTag",
		       "context":{"__type":"PermissionExpressionEnvironment"},"key":"b"}},
		   "subNode":{"__type":"PermissionDescriptorCompose","permissionDescriptors":["stack-set"]}}]}`,
		"condition granting allow": `{"__type":"PermissionDescriptorGroup","entries":[
		  {"__type":"PermissionDescriptorCondition",
		   "condition":{"__type":"PermissionExpressionHasTag","context":{"__type":"PermissionExpressionStack"},"key":"a"},
		   "subNode":{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}}]}`,
	}
	for name, wire := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := parsePolicyDetails(mustParseDescriptor(t, wire), testSetTypes)
			assert.True(t, errors.Is(err, errUnrepresentable), "got %v", err)
		})
	}
}

func TestBuildPolicyDetailsDedupesUnconditionalSets(t *testing.T) {
	t.Parallel()
	details, err := buildPolicyDetails(RbacRoleCore{
		OrganizationPermissionSetIds: []string{testOrgSetID},
		EntityRules: []RbacEntityRule{
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{All: pointerTo(true)}},
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{All: pointerTo(true)}},
		},
	})
	require.NoError(t, err)
	got, err := json.Marshal(details)
	require.NoError(t, err)
	assert.JSONEq(t, `{"__type":"PermissionDescriptorGroup","entries":[
	  {"__type":"PermissionDescriptorCompose","permissionDescriptors":["org-set","stack-set"]}]}`, string(got))
}

func TestCanonicalRoleKeyIgnoresLayout(t *testing.T) {
	t.Parallel()
	a := consoleRoleModel()
	b := consoleRoleModel()
	slicesReverse(b.EntityRules)
	b.EntityRules[0].Stack.Tags[0].Operator = pointerTo(RbacTagOperatorEquals) // explicit default
	assert.Equal(t, canonicalRoleKey(a.OrganizationPermissionSetIds, a.EntityRules),
		canonicalRoleKey(b.OrganizationPermissionSetIds, b.EntityRules))

	b.EntityRules[0].Stack.Tags[0].Value = pointerTo("OtherTeam")
	assert.NotEqual(t, canonicalRoleKey(a.OrganizationPermissionSetIds, a.EntityRules),
		canonicalRoleKey(b.OrganizationPermissionSetIds, b.EntityRules))
}

func slicesReverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
