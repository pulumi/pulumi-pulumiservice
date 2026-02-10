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
)

func TestOidcIssuerPolicySerialization(t *testing.T) {
	t.Run("serialize and deserialize all policy types", func(t *testing.T) {
		provider := &PulumiServiceOidcIssuerResource{}

		teamName := "dream-team"
		userLogin := "testuser"
		runnerID := "runner-123"
		roleID := "role-456"
		maxExpiration := int64(3600)

		// Create input with all policy types
		input := PulumiServiceOidcIssuerInput{
			Organization:         "test-org",
			Name:                 "test-issuer",
			URL:                  "https://example.com",
			MaxExpirationSeconds: &maxExpiration,
			Thumbprints:          []string{"thumbprint1", "thumbprint2"},
			Policies: []PulumiServiceAuthPolicyDefinition{
				{
					Decision:              AuthPolicyDecisionAllow,
					TokenType:             AuthPolicyTokenTypeOrganization,
					AuthorizedPermissions: []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelAdmin},
					Rules: map[string]string{
						"aud": "urn:pulumi:org:test-org",
						"sub": "repo:organization/repo:*",
					},
				},
				{
					Decision:  AuthPolicyDecisionDeny,
					TokenType: AuthPolicyTokenTypePersonal,
					UserLogin: &userLogin,
					Rules: map[string]string{
						"aud": "urn:pulumi:org:test-org",
						"sub": "pulumi:deploy:org:test-org:project:test-project:*",
					},
				},
				{
					Decision:              AuthPolicyDecisionAllow,
					TokenType:             AuthPolicyTokenTypeTeam,
					TeamName:              &teamName,
					AuthorizedPermissions: []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelStandard},
					Rules: map[string]string{
						"aud": "urn:pulumi:org:test-org",
						"sub": "repo:organization/repo:*",
					},
				},
				{
					Decision:  AuthPolicyDecisionAllow,
					TokenType: AuthPolicyTokenTypeDeploymentRunner,
					RunnerID:  &runnerID,
					Rules: map[string]string{
						"aud": "urn:pulumi:org:test-org",
						"sub": "repo:organization/repo:*",
					},
				},
				{
					Decision:  AuthPolicyDecisionDeny,
					TokenType: AuthPolicyTokenTypeOrganization,
					RoleID:    &roleID,
					Rules: map[string]string{
						"aud": "urn:pulumi:org:test-org",
						"sub": "repo:organization/repo:*",
						"env": "production",
					},
				},
			},
		}

		// Convert to property map
		propertyMap := input.toPropertyMap()

		// Convert back to input
		result := provider.ToPulumiServiceOidcIssuerInput(propertyMap)

		// Verify basic fields
		assert.Equal(t, input.Organization, result.Organization)
		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.URL, result.URL)
		assert.Equal(t, input.MaxExpirationSeconds, result.MaxExpirationSeconds)
		assert.Equal(t, input.Thumbprints, result.Thumbprints)

		// Verify all policies are preserved
		assert.Len(t, result.Policies, 5, "Should have 5 policies")

		// Verify organization policy
		orgPolicy := result.Policies[0]
		assert.Equal(t, AuthPolicyDecisionAllow, orgPolicy.Decision)
		assert.Equal(t, AuthPolicyTokenTypeOrganization, orgPolicy.TokenType)
		assert.Equal(t, []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelAdmin}, orgPolicy.AuthorizedPermissions)
		assert.Equal(t, "urn:pulumi:org:test-org", orgPolicy.Rules["aud"])
		assert.Equal(t, "repo:organization/repo:*", orgPolicy.Rules["sub"])

		// Verify personal policy
		personalPolicy := result.Policies[1]
		assert.Equal(t, AuthPolicyDecisionDeny, personalPolicy.Decision)
		assert.Equal(t, AuthPolicyTokenTypePersonal, personalPolicy.TokenType)
		assert.Equal(t, &userLogin, personalPolicy.UserLogin)
		assert.Equal(t, "urn:pulumi:org:test-org", personalPolicy.Rules["aud"])
		assert.Equal(t, "pulumi:deploy:org:test-org:project:test-project:*", personalPolicy.Rules["sub"])

		// Verify team policy
		teamPolicy := result.Policies[2]
		assert.Equal(t, AuthPolicyDecisionAllow, teamPolicy.Decision)
		assert.Equal(t, AuthPolicyTokenTypeTeam, teamPolicy.TokenType)
		assert.Equal(t, &teamName, teamPolicy.TeamName)
		assert.Equal(t, []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelStandard}, teamPolicy.AuthorizedPermissions)
		assert.Equal(t, "urn:pulumi:org:test-org", teamPolicy.Rules["aud"])
		assert.Equal(t, "repo:organization/repo:*", teamPolicy.Rules["sub"])

		// Verify runner policy
		runnerPolicy := result.Policies[3]
		assert.Equal(t, AuthPolicyDecisionAllow, runnerPolicy.Decision)
		assert.Equal(t, AuthPolicyTokenTypeDeploymentRunner, runnerPolicy.TokenType)
		assert.Equal(t, &runnerID, runnerPolicy.RunnerID)
		assert.Equal(t, "urn:pulumi:org:test-org", runnerPolicy.Rules["aud"])
		assert.Equal(t, "repo:organization/repo:*", runnerPolicy.Rules["sub"])

		// Verify policy with role ID
		rolePolicy := result.Policies[4]
		assert.Equal(t, AuthPolicyDecisionDeny, rolePolicy.Decision)
		assert.Equal(t, AuthPolicyTokenTypeOrganization, rolePolicy.TokenType)
		assert.Equal(t, &roleID, rolePolicy.RoleID)
		assert.Equal(t, "urn:pulumi:org:test-org", rolePolicy.Rules["aud"])
		assert.Equal(t, "repo:organization/repo:*", rolePolicy.Rules["sub"])
		assert.Equal(t, "production", rolePolicy.Rules["env"])
	})
}
