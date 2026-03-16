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

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOidcIssuerPolicySerialization(t *testing.T) {
	t.Parallel()
	provider := &PulumiServiceOidcIssuerResource{}

	teamName := "dream-team"
	userLogin := "testuser"
	runnerID := "runner-123"
	roleID := "role-456"
	maxExpiration := int64(3600)

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
				Decision:              AuthPolicyDecisionDeny,
				TokenType:             AuthPolicyTokenTypePersonal,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{},
				UserLogin:             &userLogin,
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
				Decision:              AuthPolicyDecisionAllow,
				TokenType:             AuthPolicyTokenTypeDeploymentRunner,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{},
				RunnerID:              &runnerID,
				Rules: map[string]string{
					"aud": "urn:pulumi:org:test-org",
					"sub": "repo:organization/repo:*",
				},
			},
			{
				Decision:              AuthPolicyDecisionDeny,
				TokenType:             AuthPolicyTokenTypeOrganization,
				RoleID:                &roleID,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{},
				Rules: map[string]string{
					"aud": "urn:pulumi:org:test-org",
					"sub": "repo:organization/repo:*",
					"env": "production",
				},
			},
		},
	}

	propertyMap := input.toPropertyMap()
	result := provider.ToPulumiServiceOidcIssuerInput(propertyMap)
	assert.Equal(t, input, result)
}
