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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	testOidcIssuerID   = "issuer-123"
	testOidcIssuerName = "test-issuer"
	testOidcIssuerURL  = "https://example.com"
	testAuthPolicyID   = "policy-1"
	testPolicyRoleID   = "role-456"
)

type oidcIssuerClientMock struct {
	config.Client
	register func(
		ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest,
	) (*pulumiapi.OidcIssuerRegistrationResponse, error)
	update func(
		ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest,
	) (*pulumiapi.OidcIssuerRegistrationResponse, error)
	getPolicies    func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error)
	updatePolicies func(
		ctx context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest,
	) (*pulumiapi.AuthPolicy, error)
}

func (c *oidcIssuerClientMock) RegisterOidcIssuer(
	ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest,
) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
	return c.register(ctx, org, req)
}

func (c *oidcIssuerClientMock) UpdateOidcIssuer(
	ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest,
) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
	return c.update(ctx, org, issuerID, req)
}

func (c *oidcIssuerClientMock) GetAuthPolicies(
	ctx context.Context, org, issuerID string,
) (*pulumiapi.AuthPolicy, error) {
	return c.getPolicies(ctx, org, issuerID)
}

func (c *oidcIssuerClientMock) UpdateAuthPolicies(
	ctx context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest,
) (*pulumiapi.AuthPolicy, error) {
	return c.updatePolicies(ctx, org, policyID, req)
}

// newOidcIssuerClientMock returns a client whose auth policy calls behave like
// Pulumi Cloud: the update response leaves out roleID, the read returns it.
func newOidcIssuerClientMock(t *testing.T) *oidcIssuerClientMock {
	t.Helper()
	stored := &pulumiapi.AuthPolicy{ID: testAuthPolicyID}
	return &oidcIssuerClientMock{
		getPolicies: func(_ context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			assert.Equal(t, gcMyOrg, org)
			assert.Equal(t, testOidcIssuerID, issuerID)
			return stored, nil
		},
		updatePolicies: func(
			_ context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest,
		) (*pulumiapi.AuthPolicy, error) {
			assert.Equal(t, gcMyOrg, org)
			assert.Equal(t, testAuthPolicyID, policyID)
			stored = &pulumiapi.AuthPolicy{ID: testAuthPolicyID, Version: stored.Version + 1}
			response := &pulumiapi.AuthPolicy{ID: testAuthPolicyID, Version: stored.Version}
			for i := range req.Definition {
				stored.Definition = append(stored.Definition, &req.Definition[i])
				withoutRole := req.Definition[i]
				withoutRole.RoleID = nil
				response.Definition = append(response.Definition, &withoutRole)
			}
			return response, nil
		},
	}
}

func testRolePolicies() []AuthPolicyDefinition {
	roleID := testPolicyRoleID
	return []AuthPolicyDefinition{{
		Decision:  AuthPolicyDecisionAllow,
		TokenType: AuthPolicyTokenTypeOrganization,
		RoleID:    &roleID,
		Rules:     map[string]string{"repository": "my-org/repo"},
	}}
}

func TestOidcIssuerCreateKeepsPolicyRoleID(t *testing.T) {
	mock := newOidcIssuerClientMock(t)
	mock.register = func(
		_ context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest,
	) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
		assert.Equal(t, gcMyOrg, org)
		return &pulumiapi.OidcIssuerRegistrationResponse{ID: testOidcIssuerID, Name: req.Name, URL: req.URL}, nil
	}
	ctx := config.WithMockClient(context.Background(), mock)

	policies := testRolePolicies()
	resp, err := (&OidcIssuer{}).Create(ctx, infer.CreateRequest[OidcIssuerInput]{
		Inputs: OidcIssuerInput{
			Organization: gcMyOrg,
			Name:         testOidcIssuerName,
			URL:          testOidcIssuerURL,
			Policies:     policies,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, oidcIssuerID(gcMyOrg, testOidcIssuerID), resp.ID)
	assert.Equal(t, policies, resp.Output.Policies)
}

func TestOidcIssuerUpdateKeepsPolicyRoleID(t *testing.T) {
	mock := newOidcIssuerClientMock(t)
	mock.update = func(
		_ context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest,
	) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
		assert.Equal(t, gcMyOrg, org)
		assert.Equal(t, testOidcIssuerID, issuerID)
		return &pulumiapi.OidcIssuerRegistrationResponse{ID: issuerID, Name: *req.Name, URL: testOidcIssuerURL}, nil
	}
	ctx := config.WithMockClient(context.Background(), mock)

	policies := testRolePolicies()
	resp, err := (&OidcIssuer{}).Update(ctx, infer.UpdateRequest[OidcIssuerInput, OidcIssuerState]{
		ID: oidcIssuerID(gcMyOrg, testOidcIssuerID),
		Inputs: OidcIssuerInput{
			Organization: gcMyOrg,
			Name:         testOidcIssuerName,
			URL:          testOidcIssuerURL,
			Policies:     policies,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, policies, resp.Output.Policies)
}

func TestOidcIssuerID(t *testing.T) {
	assert.Equal(t, "my-org/issuer-123", oidcIssuerID(gcMyOrg, "issuer-123"))
}

func TestSplitOidcIssuerID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		org, id, err := splitOidcIssuerID("my-org/issuer-123")
		require.NoError(t, err)
		assert.Equal(t, gcMyOrg, org)
		assert.Equal(t, "issuer-123", id)
	})

	t.Run("too few parts", func(t *testing.T) {
		_, _, err := splitOidcIssuerID("just-one")
		require.Error(t, err)
	})

	t.Run("too many parts", func(t *testing.T) {
		_, _, err := splitOidcIssuerID("a/b/c")
		require.Error(t, err)
	})
}

func TestOidcIssuerPolicySerialization(t *testing.T) {
	t.Parallel()

	teamName := "dream-team"
	userLogin := "testuser"
	runnerID := "runner-123"
	roleID := "role-456"
	audience := "urn:pulumi:org:test-org"
	sub := "repo:organization/repo:*"
	defaultRule := map[string]string{
		"aud": audience,
		"sub": sub,
	}
	maxExpiration := int64(3600)

	input := OidcIssuerInput{
		Organization:         "test-org",
		Name:                 "test-issuer",
		URL:                  "https://example.com",
		MaxExpirationSeconds: &maxExpiration,
		Thumbprints:          []string{"thumbprint1", "thumbprint2"},
		Policies: []AuthPolicyDefinition{
			{
				Decision:              AuthPolicyDecisionAllow,
				TokenType:             AuthPolicyTokenTypeOrganization,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelAdmin},
				Rules:                 defaultRule,
			},
			{
				Decision:  AuthPolicyDecisionDeny,
				TokenType: AuthPolicyTokenTypePersonal,
				UserLogin: &userLogin,
				Rules: map[string]string{
					"aud": audience,
					"sub": "pulumi:deploy:org:test-org:project:test-project:*",
					"env": "production",
				},
			},
			{
				Decision:              AuthPolicyDecisionAllow,
				TokenType:             AuthPolicyTokenTypeTeam,
				TeamName:              &teamName,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelStandard},
				Rules:                 defaultRule,
			},
			{
				Decision:  AuthPolicyDecisionAllow,
				TokenType: AuthPolicyTokenTypeDeploymentRunner,
				RunnerID:  &runnerID,
				Rules:     defaultRule,
			},
			{
				Decision:  AuthPolicyDecisionDeny,
				TokenType: AuthPolicyTokenTypeOrganization,
				RoleID:    &roleID,
				Rules:     defaultRule,
			},
		},
	}

	apiRequest := policiesToAPIRequest(input.Policies)
	apiPoliciesPtr := make([]*pulumiapi.AuthPolicyDefinition, len(apiRequest.Definition))
	for i := range apiRequest.Definition {
		apiPoliciesPtr[i] = &apiRequest.Definition[i]
	}

	resultPolicies := apiPoliciesToInputs(apiPoliciesPtr)
	assert.Equal(t, input.Policies, resultPolicies)
}
