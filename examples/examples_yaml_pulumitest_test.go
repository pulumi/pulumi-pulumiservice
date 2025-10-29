//go:build yaml_pulumitest
// +build yaml_pulumitest

// Copyright 2016-2025, Pulumi Corporation.

package examples

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// createEmptyPulumiProject creates a temporary directory with an empty Pulumi.yaml project
func createEmptyPulumiProject(t *testing.T, projectName string) string {
	tempDir, err := ioutil.TempDir("", "pulumitest-empty-"+projectName)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a minimal Pulumi.yaml with the specified project name but no resources or config
	pulumiYaml := fmt.Sprintf(`name: %s
runtime: yaml
description: Empty project for multistep testing
resources: {}
`, projectName)

	pulumiYamlPath := filepath.Join(tempDir, "Pulumi.yaml")
	if err := ioutil.WriteFile(pulumiYamlPath, []byte(pulumiYaml), 0644); err != nil {
		t.Fatalf("Failed to write Pulumi.yaml: %v", err)
	}

	return tempDir
}

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
			
			// Get organization from PULUMI_TEST_OWNER (like integration tests do)
			orgName := os.Getenv("PULUMI_TEST_OWNER")
			if orgName == "" {
				orgName = "service-provider-test-org" // fallback to default
			}
			
			// Create fully qualified stack name: {org}/{project}/{stack}
			projectName := testCase.directoryName
			stackName := fmt.Sprintf("%s/%s/test", orgName, projectName)
			
			test := pulumitest.NewPulumiTest(t, testCase.directoryName,
				opttest.LocalProviderPath("pulumiservice", "../bin"),
				opttest.UseAmbientBackend(),
				opttest.StackName(stackName))

			// UseAmbientBackend() automatically handles API URL and access token from environment

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
			
			// Get organization from PULUMI_TEST_OWNER (like integration tests do)
			orgName := os.Getenv("PULUMI_TEST_OWNER")
			if orgName == "" {
				orgName = "service-provider-test-org" // fallback to default
			}
			
			// Add custom stack name if specified (needed for DeploymentSettings/Schedules tests)
			if testCase.stackName != "" {
				// Create fully qualified stack name: {org}/{project}/{stack}
				projectName := testCase.directoryName
				qualifiedStackName := fmt.Sprintf("%s/%s/%s", orgName, projectName, testCase.stackName)
				options = append(options, opttest.StackName(qualifiedStackName))
			} else {
				// Create default fully qualified stack name
				projectName := testCase.directoryName
				qualifiedStackName := fmt.Sprintf("%s/%s/test", orgName, projectName)
				options = append(options, opttest.StackName(qualifiedStackName))
			}
			
			test := pulumitest.NewPulumiTest(t, testCase.directoryName, options...)

			// UseAmbientBackend() automatically handles API URL and access token from environment

			// Set the required configuration
			for key, value := range testCase.config {
				test.SetConfig(t, key, value)
			}

			// Don't override organization - let it use the ambient backend organization

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
		// Get organization from PULUMI_TEST_OWNER (like integration tests do)
		orgName := os.Getenv("PULUMI_TEST_OWNER")
		if orgName == "" {
			orgName = "service-provider-test-org" // fallback to default
		}
		
		// Generate unique stack name for this test run
		digits := generateRandomFiveDigitsPulumiTest()
		stackName := "test-stack-" + digits
		
		// Create fully qualified stack name: {org}/{project}/{stack}
		projectName := "yaml-stack-tags-example"
		qualifiedStackName := fmt.Sprintf("%s/%s/%s", orgName, projectName, stackName)
		
		// Step 1: Create empty project directory
		emptyDir := createEmptyPulumiProject(t, projectName)
		defer os.RemoveAll(emptyDir) // cleanup temp directory
		
		// Step 2: Start with empty project directory
		test := pulumitest.NewPulumiTest(t, emptyDir,
			opttest.LocalProviderPath("pulumiservice", "../bin"),
			opttest.UseAmbientBackend(),
			opttest.StackName(qualifiedStackName))

		// UseAmbientBackend() automatically handles API URL and access token from environment

		// Set organization if available
		// Don't override organization - let it use the ambient backend organization

		// Set test user name if PULUMI_TEST_USER is available
		if userName := os.Getenv("PULUMI_TEST_USER"); userName != "" {
			test.SetConfig(t, "testUserName", userName)
		}

		// Step 3: Create empty stack first
		test.Up(t) // This creates an empty stack with no resources
		
		// Step 4: Update source to stack tags and deploy
		test.UpdateSource(t, "yaml-stack-tags")
		test.Up(t)

		// Step 5: Verify no changes on subsequent preview
		previewResult := test.Preview(t)
		assertpreview.HasNoChanges(t, previewResult)

		// Step 6: Refresh and verify state is consistent
		test.Refresh(t)
	})
}

// Multistep tests for resources that need empty stacks created first
func TestYamlDeploymentSettingsWithPulumiTest(t *testing.T) {
	t.Run("YamlDeploymentSettings", func(t *testing.T) {
		// Get organization from PULUMI_TEST_OWNER (like integration tests do)
		orgName := os.Getenv("PULUMI_TEST_OWNER")
		if orgName == "" {
			orgName = "service-provider-test-org" // fallback to default
		}
		
		digits := generateRandomFiveDigitsPulumiTest()
		stackName := "test-stack-" + digits
		
		// Create fully qualified stack name: {org}/{project}/{stack}
		projectName := "yaml-deployment-settings-example"
		qualifiedStackName := fmt.Sprintf("%s/%s/%s", orgName, projectName, stackName)
		
		// Step 1: Create empty project directory
		emptyDir := createEmptyPulumiProject(t, projectName)
		defer os.RemoveAll(emptyDir) // cleanup temp directory
		
		// Step 2: Start with empty project directory
		test := pulumitest.NewPulumiTest(t, emptyDir,
			opttest.LocalProviderPath("pulumiservice", "../bin"),
			opttest.UseAmbientBackend(),
			opttest.StackName(qualifiedStackName))

		// UseAmbientBackend() automatically handles API URL and access token from environment
		test.SetConfig(t, "digits", digits)
		
		// Don't override organization - let it use the ambient backend organization

		// Step 3: Create empty stack first
		test.Up(t) // This creates an empty stack with no resources
		
		// Step 4: Update source to deployment settings and deploy
		test.UpdateSource(t, "yaml-deployment-settings")
		test.Up(t)

		// Step 5: Verify no changes on subsequent preview
		previewResult := test.Preview(t)
		assertpreview.HasNoChanges(t, previewResult)

		// Step 6: Refresh and verify state is consistent
		test.Refresh(t)
	})
}

func TestYamlDeploymentSettingsNoSourceWithPulumiTest(t *testing.T) {
	t.Run("YamlDeploymentSettingsNoSource", func(t *testing.T) {
		// Get organization from PULUMI_TEST_OWNER (like integration tests do)
		orgName := os.Getenv("PULUMI_TEST_OWNER")
		if orgName == "" {
			orgName = "service-provider-test-org" // fallback to default
		}
		
		digits := generateRandomFiveDigitsPulumiTest()
		stackName := "test-stack-" + digits
		
		// Create fully qualified stack name: {org}/{project}/{stack}
		projectName := "yaml-deployment-settings-example" // match the project name in yaml-deployment-settings-no-source
		qualifiedStackName := fmt.Sprintf("%s/%s/%s", orgName, projectName, stackName)
		
		// Step 1: Create empty project directory
		emptyDir := createEmptyPulumiProject(t, projectName)
		defer os.RemoveAll(emptyDir) // cleanup temp directory
		
		// Step 2: Start with empty project directory
		test := pulumitest.NewPulumiTest(t, emptyDir,
			opttest.LocalProviderPath("pulumiservice", "../bin"),
			opttest.UseAmbientBackend(),
			opttest.StackName(qualifiedStackName))

		// UseAmbientBackend() automatically handles API URL and access token from environment
		test.SetConfig(t, "digits", digits)
		
		// Don't override organization - let it use the ambient backend organization

		// Step 3: Create empty stack first
		test.Up(t) // This creates an empty stack with no resources
		
		// Step 4: Update source to deployment settings and deploy
		test.UpdateSource(t, "yaml-deployment-settings-no-source")
		test.Up(t)

		// Step 5: Verify no changes on subsequent preview
		previewResult := test.Preview(t)
		assertpreview.HasNoChanges(t, previewResult)

		// Step 6: Refresh and verify state is consistent
		test.Refresh(t)
	})
}

func TestYamlDeploymentSettingsCommitWithPulumiTest(t *testing.T) {
	t.Run("YamlDeploymentSettingsCommit", func(t *testing.T) {
		// Get organization from PULUMI_TEST_OWNER (like integration tests do)
		orgName := os.Getenv("PULUMI_TEST_OWNER")
		if orgName == "" {
			orgName = "service-provider-test-org" // fallback to default
		}
		
		digits := generateRandomFiveDigitsPulumiTest()
		stackName := "test-stack-" + digits
		
		// Create fully qualified stack name: {org}/{project}/{stack}
		projectName := "yaml-deployment-settings-commit-example"
		qualifiedStackName := fmt.Sprintf("%s/%s/%s", orgName, projectName, stackName)
		
		// Step 1: Create empty project directory
		emptyDir := createEmptyPulumiProject(t, projectName)
		defer os.RemoveAll(emptyDir) // cleanup temp directory
		
		// Step 2: Start with empty project directory
		test := pulumitest.NewPulumiTest(t, emptyDir,
			opttest.LocalProviderPath("pulumiservice", "../bin"),
			opttest.UseAmbientBackend(),
			opttest.StackName(qualifiedStackName))

		// UseAmbientBackend() automatically handles API URL and access token from environment
		test.SetConfig(t, "digits", digits)
		
		// Don't override organization - let it use the ambient backend organization

		// Step 3: Create empty stack first
		test.Up(t) // This creates an empty stack with no resources
		
		// Step 4: Update source to deployment settings and deploy
		test.UpdateSource(t, "yaml-deployment-settings-commit")
		test.Up(t)

		// Step 5: Verify no changes on subsequent preview
		previewResult := test.Preview(t)
		assertpreview.HasNoChanges(t, previewResult)

		// Step 6: Refresh and verify state is consistent
		test.Refresh(t)
	})
}

func TestYamlSchedulesWithPulumiTest(t *testing.T) {
	t.Run("YamlSchedules", func(t *testing.T) {
		// Get organization from PULUMI_TEST_OWNER (like integration tests do)
		orgName := os.Getenv("PULUMI_TEST_OWNER")
		if orgName == "" {
			orgName = "service-provider-test-org" // fallback to default
		}
		
		digits := generateRandomFiveDigitsPulumiTest()
		stackName := "test-stack-" + digits
		
		// Create fully qualified stack name: {org}/{project}/{stack}
		projectName := "pulumi-service-schedules-example-yaml"
		qualifiedStackName := fmt.Sprintf("%s/%s/%s", orgName, projectName, stackName)
		
		// Step 1: Create empty project directory
		emptyDir := createEmptyPulumiProject(t, projectName)
		defer os.RemoveAll(emptyDir) // cleanup temp directory
		
		// Step 2: Start with empty project directory
		test := pulumitest.NewPulumiTest(t, emptyDir,
			opttest.LocalProviderPath("pulumiservice", "../bin"),
			opttest.UseAmbientBackend(),
			opttest.StackName(qualifiedStackName))

		// UseAmbientBackend() automatically handles API URL and access token from environment
		test.SetConfig(t, "digits", digits)
		
		// Don't override organization - let it use the ambient backend organization

		// Step 3: Create empty stack first
		test.Up(t) // This creates an empty stack with no resources
		
		// Step 4: Update source to schedules and deploy
		test.UpdateSource(t, "yaml-schedules")
		test.Up(t)

		// Step 5: Verify no changes on subsequent preview
		previewResult := test.Preview(t)
		assertpreview.HasNoChanges(t, previewResult)

		// Step 6: Refresh and verify state is consistent
		test.Refresh(t)
	})
}

// generateRandomFiveDigitsPulumiTest generates a 5-digit random number string for yaml_pulumitest
func generateRandomFiveDigitsPulumiTest() string {
	return fmt.Sprintf("%05d", rand.Intn(100000))
}
