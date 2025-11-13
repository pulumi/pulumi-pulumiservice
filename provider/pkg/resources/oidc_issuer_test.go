// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"slices"
	"sort"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// Regression test for issue #542 (https://github.com/pulumi/pulumi-pulumiservice/issues/542)
// and issue #538 (https://github.com/pulumi/pulumi-pulumiservice/issues/538)
//
// This test verifies that the OIDC Issuer resource preserves the order of auth policies
// returned by the API, and does not introduce spurious diffs through client-side reordering.
//
// The bug was caused by the sortPolicies() function which used gob.NewEncoder for comparison.
// Gob encoding doesn't guarantee deterministic ordering of map keys, making the comparison
// non-deterministic and causing false diffs when running `pulumi refresh` repeatedly.
//
// The fix (PR #542) removed the sortPolicies() function entirely, trusting the API to provide
// consistent ordering. These tests ensure that policy ordering is preserved throughout the
// Read() and Check() operations without any client-side reordering.

// Mock implementation of pulumiapi.OidcClient for testing
// Unused methods will panic when called via embedded interface
type OidcClientMock struct {
	pulumiapi.OidcClient
	getOidcIssuerFunc   func(ctx context.Context, org string, id string) (*pulumiapi.OidcIssuerRegistrationResponse, error)
	getAuthPoliciesFunc func(ctx context.Context, org string, id string) (*pulumiapi.AuthPolicy, error)
}

func (m *OidcClientMock) GetOidcIssuer(ctx context.Context, org string, id string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
	return m.getOidcIssuerFunc(ctx, org, id)
}

func (m *OidcClientMock) GetAuthPolicies(ctx context.Context, org string, id string) (*pulumiapi.AuthPolicy, error) {
	return m.getAuthPoliciesFunc(ctx, org, id)
}

// createTestPolicies creates three distinct test policies in a specific order
// Each policy has unique characteristics to make the ordering observable in tests
func createTestPolicies() []*pulumiapi.AuthPolicyDefinition {
	teamA := "team-alpha"
	userC := "user-charlie"

	return []*pulumiapi.AuthPolicyDefinition{
		// PolicyA: Organization token with repository rules
		{
			Decision:              "allow",
			TokenType:             "organization",
			AuthorizedPermissions: []string{"admin"},
			Rules: map[string]string{
				"repository": "repo-a",
			},
		},
		// PolicyB: Team token with team constraints
		{
			Decision:              "allow",
			TokenType:             "team",
			TeamName:              &teamA,
			AuthorizedPermissions: []string{"standard"},
			Rules: map[string]string{
				"environment": "staging",
			},
		},
		// PolicyC: Personal token with user constraints
		{
			Decision:              "deny",
			TokenType:             "personal",
			UserLogin:             &userC,
			AuthorizedPermissions: []string{},
			Rules: map[string]string{
				"project": "critical-infra",
			},
		},
	}
}

// createReversedTestPolicies returns test policies in reversed order
func createReversedTestPolicies() []*pulumiapi.AuthPolicyDefinition {
	policies := createTestPolicies()
	slices.Reverse(policies)
	return policies
}

// extractPolicyOrderFromProperties extracts policy identifiable characteristics from resource.PropertyMap.
// Input structure: PropertyMap with "policies" key containing array of policy objects.
// Each policy object has "tokenType" and "rules" properties.
// Returns a slice of strings representing the order of policies (e.g., ["organization:repository", "team:environment"]).
func extractPolicyOrderFromProperties(props *resource.PropertyMap) ([]string, error) {
	if props == nil {
		return nil, nil
	}

	policiesValue, ok := (*props)["policies"]
	if !ok || !policiesValue.HasValue() {
		return nil, nil
	}

	policiesArray := policiesValue.ArrayValue()
	order := make([]string, 0, len(policiesArray))

	for _, policyValue := range policiesArray {
		policyMap := policyValue.ObjectValue()

		// Use tokenType + first rule key as identifier
		tokenType := policyMap["tokenType"].StringValue()

		// Get the first rule key to make each policy identifiable
		// Sort keys to ensure deterministic ordering (Go map iteration is non-deterministic)
		rulesMap := policyMap["rules"].ObjectValue()
		ruleKeys := make([]string, 0, len(rulesMap))
		for k := range rulesMap {
			ruleKeys = append(ruleKeys, string(k))
		}
		sort.Strings(ruleKeys)

		var ruleKey string
		if len(ruleKeys) > 0 {
			ruleKey = ruleKeys[0]
		}

		identifier := tokenType + ":" + ruleKey
		order = append(order, identifier)
	}

	return order, nil
}

// TestOidcIssuer_PolicyOrderingDoesNotCauseDrift verifies that the OIDC Issuer resource preserves policy order
// and does not introduce spurious diffs.
func TestOidcIssuer_PolicyOrderingDoesNotCauseDrift(t *testing.T) {
	t.Run("Read preserves policy order from API", func(t *testing.T) {
		// Setup: Mock returns policies in specific order [A, B, C]
		testPolicies := createTestPolicies()
		mockedClient := &OidcClientMock{
			getOidcIssuerFunc: func(ctx context.Context, org string, id string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
				return &pulumiapi.OidcIssuerRegistrationResponse{
					ID:          "issuer-123",
					Name:        "test-issuer",
					URL:         "https://token.actions.githubusercontent.com",
					Issuer:      "https://token.actions.githubusercontent.com",
					Thumbprints: []string{"a1b2c3d4"},
				}, nil
			},
			getAuthPoliciesFunc: func(ctx context.Context, org string, id string) (*pulumiapi.AuthPolicy, error) {
				return &pulumiapi.AuthPolicy{
					ID:         "policy-abc",
					Version:    1,
					Definition: testPolicies,
				}, nil
			},
		}

		provider := PulumiServiceOidcIssuerResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/issuer-123",
			Urn: "urn:pulumi:test::test::pulumiservice:index:OidcIssuer::test-issuer",
		}

		// Execute: Call Read() method
		resp, err := provider.Read(&req)

		// Assert: No error and response is valid
		require.NoError(t, err, "Read should not return an error")
		require.NotNil(t, resp, "Read response should not be nil")
		require.Equal(t, "test-org/issuer-123", resp.Id, "Resource ID should be preserved")

		// Assert: Policies are returned in the same order as the API
		outputMap, err := plugin.UnmarshalProperties(resp.Properties, util.StandardUnmarshal)
		require.NoError(t, err, "Should unmarshal output properties")

		outputOrder, err := extractPolicyOrderFromProperties(&outputMap)
		require.NoError(t, err, "Should extract policy order")
		require.Len(t, outputOrder, 3, "Should have 3 policies")

		// Verify order matches API response: organization:repository, team:environment, personal:project
		expectedOrder := []string{
			"organization:repository",
			"team:environment",
			"personal:project",
		}
		assert.Equal(t, expectedOrder, outputOrder, "Policy order should match API response order")
	})

	t.Run("Check preserves input policy order", func(t *testing.T) {
		// Setup: Create input properties with policies in specific order
		teamName := "team-alpha"
		userLogin := "user-charlie"

		inputPolicies := []PulumiServiceAuthPolicyDefinition{
			{
				Decision:              AuthPolicyDecisionAllow,
				TokenType:             AuthPolicyTokenTypeOrganization,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelAdmin},
				Rules:                 map[string]string{"repository": "repo-a"},
			},
			{
				Decision:              AuthPolicyDecisionAllow,
				TokenType:             AuthPolicyTokenTypeTeam,
				TeamName:              &teamName,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{AuthPolicyPermissionLevelStandard},
				Rules:                 map[string]string{"environment": "staging"},
			},
			{
				Decision:              AuthPolicyDecisionDeny,
				TokenType:             AuthPolicyTokenTypePersonal,
				UserLogin:             &userLogin,
				AuthorizedPermissions: []AuthPolicyPermissionLevel{},
				Rules:                 map[string]string{"project": "critical-infra"},
			},
		}

		input := PulumiServiceOidcIssuerInput{
			Organization: "test-org",
			Name:         "test-issuer",
			URL:          "https://token.actions.githubusercontent.com",
			Policies:     inputPolicies,
		}

		inputMap := input.toPropertyMap()
		inputProps, err := plugin.MarshalProperties(inputMap, util.StandardMarshal)
		require.NoError(t, err, "Should marshal input properties")

		provider := PulumiServiceOidcIssuerResource{
			Client: &OidcClientMock{}, // Check doesn't use client
		}

		req := pulumirpc.CheckRequest{
			Urn:  "urn:pulumi:test::test::pulumiservice:index:OidcIssuer::test-issuer",
			News: inputProps,
		}

		// Execute: Call Check() method
		resp, err := provider.Check(&req)

		// Assert: No error and response is valid
		require.NoError(t, err, "Check should not return an error")
		require.NotNil(t, resp, "Check response should not be nil")
		require.Nil(t, resp.Failures, "Check should not return failures")

		// Assert: Extract policy order from checked inputs
		checkedMap, err := plugin.UnmarshalProperties(resp.Inputs, util.StandardUnmarshal)
		require.NoError(t, err, "Should unmarshal checked properties")

		checkedOrder, err := extractPolicyOrderFromProperties(&checkedMap)
		require.NoError(t, err, "Should extract policy order from checked inputs")

		// Assert: Check preserves input order without reordering
		expectedOrder := []string{
			"organization:repository",
			"team:environment",
			"personal:project",
		}
		assert.Equal(t, expectedOrder, checkedOrder, "Check should preserve input policy order without reordering")
	})

	t.Run("Different API policy orders are respected", func(t *testing.T) {
		// Setup: Mock returns policies in REVERSED order [C, B, A]
		reversedPolicies := createReversedTestPolicies()
		mockedClient := &OidcClientMock{
			getOidcIssuerFunc: func(ctx context.Context, org string, id string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
				return &pulumiapi.OidcIssuerRegistrationResponse{
					ID:          "issuer-789",
					Name:        "reversed-issuer",
					URL:         "https://token.actions.githubusercontent.com",
					Issuer:      "https://token.actions.githubusercontent.com",
					Thumbprints: []string{"i9j0k1l2"},
				}, nil
			},
			getAuthPoliciesFunc: func(ctx context.Context, org string, id string) (*pulumiapi.AuthPolicy, error) {
				return &pulumiapi.AuthPolicy{
					ID:         "policy-ghi",
					Version:    1,
					Definition: reversedPolicies,
				}, nil
			},
		}

		provider := PulumiServiceOidcIssuerResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/issuer-789",
			Urn: "urn:pulumi:test::test::pulumiservice:index:OidcIssuer::reversed-issuer",
		}

		// Execute: Call Read() method
		resp, err := provider.Read(&req)

		// Assert: No error and response is valid
		require.NoError(t, err, "Read should not return an error")
		require.NotNil(t, resp, "Read response should not be nil")

		// Assert: Extract policy order
		outputMap, err := plugin.UnmarshalProperties(resp.Properties, util.StandardUnmarshal)
		require.NoError(t, err, "Should unmarshal output properties")

		outputOrder, err := extractPolicyOrderFromProperties(&outputMap)
		require.NoError(t, err, "Should extract policy order")
		require.Len(t, outputOrder, 3, "Should have 3 policies")

		// Assert: Provider respects API ordering (reversed), doesn't impose its own sort
		expectedReversedOrder := []string{
			"personal:project",
			"team:environment",
			"organization:repository",
		}
		assert.Equal(t, expectedReversedOrder, outputOrder, "Provider should respect API's reversed policy order without imposing alphabetical or other sorting")
	})
}
