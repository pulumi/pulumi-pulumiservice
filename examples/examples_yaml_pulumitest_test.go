//go:build yaml_pulumitest || all
// +build yaml_pulumitest all

// Copyright 2016-2025, Pulumi Corporation.

package examples

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestYamlExamplesWithPulumiTest(t *testing.T) {
	// Simple examples that don't need additional configuration
	simpleTests := map[string]struct {
		directoryName string
	}{
		"YamlAccessTokens":         {directoryName: "yaml-access-tokens"},
		"YamlTeams":                {directoryName: "yaml-teams"},
		"YamlTeamAccessToken":      {directoryName: "yaml-team-token"},
		"YamlOrgAccessToken":       {directoryName: "yaml-org-token"},
		"YamlTeamStackPermissions": {directoryName: "yaml-team-stack-permissions"},
		"YamlWebhook":              {directoryName: "yaml-webhooks"},
		// "YamlOidcIssuer":           {directoryName: "yaml-oidc-issuer"}, // Skipping - OIDC issuers already exist in team-ce
	}

	for name, testCase := range simpleTests {
		testCase := testCase // capture for parallel execution
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			
			test := pulumitest.NewPulumiTest(t, testCase.directoryName,
				opttest.LocalProviderPath("pulumiservice", "../bin"),
				opttest.UseAmbientBackend())

			// Configure the pulumiservice provider
			// Use the staging API URL if PULUMI_BACKEND_URL is set (matches CI environment)
			apiUrl := "https://api.pulumi.com"
			if backendUrl := os.Getenv("PULUMI_BACKEND_URL"); backendUrl != "" {
				apiUrl = backendUrl
			}
			test.SetConfig(t, "pulumiservice:apiUrl", apiUrl)
			
			// Set access token from environment variable (required for API access)
			token := os.Getenv("PULUMI_ACCESS_TOKEN")
			if token == "" {
				t.Fatal("PULUMI_ACCESS_TOKEN environment variable is required for pulumitest")
			}
			test.SetConfig(t, "pulumiservice:accessToken", token)

			// Set default organization if PULUMI_TEST_OWNER is available
			if orgName := os.Getenv("PULUMI_TEST_OWNER"); orgName != "" {
				test.SetConfig(t, "organizationName", orgName)
				test.SetConfig(t, "pulumiservice:organizationName", orgName)
			}

			// Set test user name if PULUMI_TEST_USER is available
			if userName := os.Getenv("PULUMI_TEST_USER"); userName != "" {
				test.SetConfig(t, "testUserName", userName)
			}

			// Deploy the infrastructure
			test.Up(t)

			// Verify no changes on subsequent preview
			previewResult := test.Preview(t)
			assertpreview.HasNoChanges(t, previewResult)

			// Refresh and verify state is consistent
			test.Refresh(t)
		})
	}
}

func TestYamlExamplesWithConfigWithPulumiTest(t *testing.T) {
	// Tests that need additional configuration parameters
	configTests := map[string]struct {
		directoryName string
		config        map[string]string
		stackName     string // Custom stack name for tests that need existing stacks
	}{
		"YamlDeploymentSettings": func() struct {
			directoryName string
			config        map[string]string
			stackName     string
		} {
			digits := generateRandomFiveDigitsPulumiTest()
			return struct {
				directoryName string
				config        map[string]string
				stackName     string
			}{
				directoryName: "yaml-deployment-settings",
				config: map[string]string{
					"digits": digits,
				},
				stackName: "test-stack-" + digits,
			}
		}(),
		"YamlDeploymentSettingsNoSource": func() struct {
			directoryName string
			config        map[string]string
			stackName     string
		} {
			digits := generateRandomFiveDigitsPulumiTest()
			return struct {
				directoryName string
				config        map[string]string
				stackName     string
			}{
				directoryName: "yaml-deployment-settings-no-source",
				config: map[string]string{
					"digits": digits,
				},
				stackName: "test-stack-" + digits,
			}
		}(),
		"YamlDeploymentSettingsCommit": func() struct {
			directoryName string
			config        map[string]string
			stackName     string
		} {
			digits := generateRandomFiveDigitsPulumiTest()
			return struct {
				directoryName string
				config        map[string]string
				stackName     string
			}{
				directoryName: "yaml-deployment-settings-commit",
				config: map[string]string{
					"digits": digits,
				},
				stackName: "test-stack-" + digits,
			}
		}(),
		"YamlSchedules": func() struct {
			directoryName string
			config        map[string]string
			stackName     string
		} {
			digits := generateRandomFiveDigitsPulumiTest()
			return struct {
				directoryName string
				config        map[string]string
				stackName     string
			}{
				directoryName: "yaml-schedules",
				config: map[string]string{
					"digits": digits,
				},
				stackName: "test-stack-" + digits,
			}
		}(),
		"YamlEnvironments": {
			directoryName: "yaml-environments",
			config: map[string]string{
				"digits": generateRandomFiveDigitsPulumiTest(),
			},
		},
		"YamlAgentPools": {
			directoryName: "yaml-agent-pools",
			config: map[string]string{
				"digits": generateRandomFiveDigitsPulumiTest(),
			},
		},
		"YamlTemplateSources": {
			directoryName: "yaml-template-sources",
			config: map[string]string{
				"digits": generateRandomFiveDigitsPulumiTest(),
			},
		},
		"YamlPolicyGroups": {
			directoryName: "yaml-policy-groups",
			config: map[string]string{
				"digits": generateRandomFiveDigitsPulumiTest(),
			},
		},
		"YamlApprovalRules": {
			directoryName: "yaml-approval-rules",
			config: map[string]string{
				"digits": generateRandomFiveDigitsPulumiTest(),
			},
		},
	}

	for name, testCase := range configTests {
		testCase := testCase // capture for parallel execution
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			
			// Build options for pulumitest
			options := []opttest.Option{
				opttest.LocalProviderPath("pulumiservice", "../bin"),
				opttest.UseAmbientBackend(),
			}
			
			// Add custom stack name if specified (needed for DeploymentSettings/Schedules tests)
			if testCase.stackName != "" {
				options = append(options, opttest.StackName(testCase.stackName))
			}
			
			test := pulumitest.NewPulumiTest(t, testCase.directoryName, options...)

			// Configure the pulumiservice provider
			// Use the staging API URL if PULUMI_BACKEND_URL is set (matches CI environment)
			apiUrl := "https://api.pulumi.com"
			if backendUrl := os.Getenv("PULUMI_BACKEND_URL"); backendUrl != "" {
				apiUrl = backendUrl
			}
			test.SetConfig(t, "pulumiservice:apiUrl", apiUrl)
			
			// Set access token from environment variable (required for API access)
			token := os.Getenv("PULUMI_ACCESS_TOKEN")
			if token == "" {
				t.Fatal("PULUMI_ACCESS_TOKEN environment variable is required for pulumitest")
			}
			test.SetConfig(t, "pulumiservice:accessToken", token)

			// Set the required configuration
			for key, value := range testCase.config {
				test.SetConfig(t, key, value)
			}

			// Set organization name from environment if available
			if orgName := os.Getenv("PULUMI_TEST_OWNER"); orgName != "" {
				test.SetConfig(t, "organizationName", orgName)
				test.SetConfig(t, "pulumiservice:organizationName", orgName)
			}

			// Set test user name if PULUMI_TEST_USER is available
			if userName := os.Getenv("PULUMI_TEST_USER"); userName != "" {
				test.SetConfig(t, "testUserName", userName)
			}

			// Deploy the infrastructure
			test.Up(t)

			// Verify no changes on subsequent preview
			previewResult := test.Preview(t)
			assertpreview.HasNoChanges(t, previewResult)

			// Refresh and verify state is consistent
			test.Refresh(t)
		})
	}
}

func TestYamlStackTagsExampleWithPulumiTest(t *testing.T) {
	// Special test case for stack tags that requires a two-step process
	t.Run("YamlStackTags", func(t *testing.T) {
		test := pulumitest.NewPulumiTest(t, "yaml-stack-tags",
			opttest.LocalProviderPath("pulumiservice", "../bin"),
			opttest.UseAmbientBackend())

		// Configure the pulumiservice provider
		// Set API URL to the real Pulumi Service (not the file backend used for state)
		test.SetConfig(t, "pulumiservice:apiUrl", "https://api.pulumi.com")
		
		// Set access token from environment variable (required for API access)
		if token := os.Getenv("PULUMI_ACCESS_TOKEN"); token != "" {
			test.SetConfig(t, "pulumiservice:accessToken", token)
		}

		// Set organization if available
		if orgName := os.Getenv("PULUMI_TEST_OWNER"); orgName != "" {
			test.SetConfig(t, "organizationName", orgName)
			test.SetConfig(t, "pulumiservice:organizationName", orgName)
		}

		// Set test user name if PULUMI_TEST_USER is available
		if userName := os.Getenv("PULUMI_TEST_USER"); userName != "" {
			test.SetConfig(t, "testUserName", userName)
		}

		// Deploy the stack tags
		test.Up(t)

		// Verify no changes on subsequent preview
		previewResult := test.Preview(t)
		assertpreview.HasNoChanges(t, previewResult)

		// Refresh and verify state is consistent
		test.Refresh(t)
	})
}

// generateRandomFiveDigitsPulumiTest generates a 5-digit random number string for yaml_pulumitest
func generateRandomFiveDigitsPulumiTest() string {
	return fmt.Sprintf("%05d", rand.Intn(100000))
}
