//go:build nodejs || all
// +build nodejs all

package examples

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestAccessTokenExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-access-tokens"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	runPulumiTest(t, test)
}

func TestStackTagsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-stack-tags"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	runPulumiTest(t, test)
}

func TestDeploymentSettingsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-deployment-settings"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "my_secret", "my_secret_value")
	test.SetConfig(t, "password", "my_password")
	test.SetConfig(t, "imagePassword", "my_image_password")

	// The example passes sshPrivateKey as a plaintext literal. The provider must
	// still store it as a secret: for a property nested inside a type the
	// schema's `"secret": true` only reaches the .NET SDK, so for this Node.js
	// program only Check keeps it out of the state file.
	// Guards https://github.com/pulumi/pulumi-pulumiservice/issues/1037.
	runPulumiTestAfterUp(t, test, func(t *testing.T, test *pulumitest.PulumiTest) {
		inputs := stateInputs(t, test, "pulumiservice:index:DeploymentSettings")
		assertStateSecret(t, inputs, "sourceContext", "git", "gitAuth", "sshAuth", "sshPrivateKey")
		assertStateSecret(t, inputs, "sourceContext", "git", "gitAuth", "sshAuth", "password")
	})
}

func TestTeamStackPermissionsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-team-stack-permissions"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestTeamsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-teams"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestNodejsWebhookExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-webhooks"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestNodejsSchedulesExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-schedules"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestNodejsEnvironmentsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-environments"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())

	// Run the Pulumi program
	test.Up(t)

	// Run preview to ensure no changes after initial deployment
	previewResult := test.Preview(t)
	assertpreview.HasNoChanges(t, previewResult)

	// Skip refresh assertion since this example expects refresh changes
	test.Refresh(t)

	// Clean up - destroy the stack
	test.Destroy(t)
}

func TestNodejsTemplateSourcesExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-template-source"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	test.SetConfig(t, "sourceURL", randomTemplateSourceURL())
	runPulumiTest(t, test)
}

// templateSourceDirs are directories of github.com/pulumi/templates: the base
// language templates and the architecture templates. Pulumi Cloud allows only
// one template source per URL in an organization, so each run picks one at
// random to avoid colliding with a concurrent run or with a source that an
// earlier run left behind.
var templateSourceDirs = []string{
	"bun", "csharp", "fsharp", "go", "hcl", "java-gradle", "java-jbang", "java", "javascript",
	"python", "typescript", "visualbasic", "yaml",
	"container-aws-csharp", "container-aws-go", "container-aws-hcl", "container-aws-python",
	"container-aws-typescript", "container-aws-yaml",
	"container-azure-csharp", "container-azure-go", "container-azure-hcl", "container-azure-python",
	"container-azure-typescript",
	"container-gcp-csharp", "container-gcp-go", "container-gcp-hcl", "container-gcp-python",
	"container-gcp-typescript",
	"helm-kubernetes-csharp", "helm-kubernetes-go", "helm-kubernetes-hcl", "helm-kubernetes-python",
	"helm-kubernetes-typescript", "helm-kubernetes-yaml",
	"kubernetes-aws-csharp", "kubernetes-aws-go", "kubernetes-aws-hcl", "kubernetes-aws-python",
	"kubernetes-aws-typescript", "kubernetes-aws-yaml",
	"kubernetes-azure-csharp", "kubernetes-azure-go", "kubernetes-azure-hcl",
	"kubernetes-azure-python", "kubernetes-azure-typescript", "kubernetes-azure-yaml",
	"kubernetes-gcp-csharp", "kubernetes-gcp-go", "kubernetes-gcp-hcl", "kubernetes-gcp-python",
	"kubernetes-gcp-typescript", "kubernetes-gcp-yaml",
	"serverless-aws-csharp", "serverless-aws-go", "serverless-aws-hcl", "serverless-aws-python",
	"serverless-aws-typescript", "serverless-aws-yaml",
	"serverless-azure-csharp", "serverless-azure-go", "serverless-azure-hcl",
	"serverless-azure-python", "serverless-azure-typescript", "serverless-azure-yaml",
	"serverless-cloudflare-csharp", "serverless-cloudflare-go", "serverless-cloudflare-hcl",
	"serverless-cloudflare-python", "serverless-cloudflare-typescript", "serverless-cloudflare-yaml",
	"serverless-gcp-csharp", "serverless-gcp-go", "serverless-gcp-hcl", "serverless-gcp-python",
	"serverless-gcp-typescript", "serverless-gcp-yaml",
	"static-website-aws-bun", "static-website-aws-csharp", "static-website-aws-go",
	"static-website-aws-hcl", "static-website-aws-python", "static-website-aws-typescript",
	"static-website-aws-yaml",
	"static-website-azure-csharp", "static-website-azure-go", "static-website-azure-hcl",
	"static-website-azure-python", "static-website-azure-typescript", "static-website-azure-yaml",
	"static-website-cloudflare-csharp", "static-website-cloudflare-go",
	"static-website-cloudflare-hcl", "static-website-cloudflare-python",
	"static-website-cloudflare-typescript", "static-website-cloudflare-yaml",
	"static-website-gcp-csharp", "static-website-gcp-go", "static-website-gcp-hcl",
	"static-website-gcp-python", "static-website-gcp-typescript", "static-website-gcp-yaml",
	"vm-aws-csharp", "vm-aws-go", "vm-aws-hcl", "vm-aws-python", "vm-aws-typescript", "vm-aws-yaml",
	"vm-azure-csharp", "vm-azure-go", "vm-azure-hcl", "vm-azure-python", "vm-azure-typescript",
	"vm-azure-yaml",
	"vm-gcp-csharp", "vm-gcp-go", "vm-gcp-hcl", "vm-gcp-python", "vm-gcp-typescript", "vm-gcp-yaml",
	"webapp-kubernetes-csharp", "webapp-kubernetes-go", "webapp-kubernetes-hcl",
	"webapp-kubernetes-python", "webapp-kubernetes-typescript", "webapp-kubernetes-yaml",
}

func randomTemplateSourceURL() string {
	dir := templateSourceDirs[rand.Intn(len(templateSourceDirs))] //nolint:gosec // For testing.
	return "https://github.com/pulumi/templates/" + dir
}

func TestNodejsEnvironmentsFileAssetExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-environments-file-asset"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestNodejsOidcIssuerExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-oidc-issuer"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	for k, v := range oidcIssuerTestConfig() {
		test.SetConfig(t, k, v)
	}
	runPulumiTest(t, test)
}

func TestNodejsApprovalRulesExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-approval-rules"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestNodejsPolicyPackExample(t *testing.T) {
	exampleDir := filepath.Join(getCwd(t), "ts-policy-pack")
	yarnInstall(t, filepath.Join(exampleDir, "policy-pack"))
	test := pulumitest.NewPulumiTest(t,
		exampleDir,
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	test.SetConfig(t, "organizationName", getOrgName())
	runPulumiTest(t, test)
}

func TestNodejsInsightsAccountInvokesExample(t *testing.T) {
	digits := generateRandomFiveDigits()
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-insights-account-invokes"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", digits)
	test.SetConfig(t, "organizationName", getOrgName())
	test.SetConfig(t, "roleArn", getInsightsRoleArn(t))
	upResult := runPulumiTest(t, test)

	// Verify the resource outputs
	resourceAccountName := upResult.Outputs["resourceAccountName"].Value.(string)
	expectedAccountName := "test-invoke-account-" + digits
	assert.Equal(t, expectedAccountName, resourceAccountName)

	// Verify the getInsightsAccount invoke outputs match the resource
	fetchedAccountName := upResult.Outputs["fetchedAccountName"].Value.(string)
	assert.Equal(t, resourceAccountName, fetchedAccountName)

	fetchedInsightsAccountID := upResult.Outputs["fetchedInsightsAccountId"].Value.(string)
	resourceInsightsAccountID := upResult.Outputs["resourceInsightsAccountId"].Value.(string)
	assert.Equal(t, resourceInsightsAccountID, fetchedInsightsAccountID)

	fetchedProvider := upResult.Outputs["fetchedProvider"].Value.(string)
	assert.Equal(t, "aws", fetchedProvider)

	fetchedScanSchedule := upResult.Outputs["fetchedScanSchedule"].Value.(string)
	assert.Equal(t, "none", fetchedScanSchedule)

	fetchedScheduledScanEnabled := upResult.Outputs["fetchedScheduledScanEnabled"].Value.(bool)
	assert.False(t, fetchedScheduledScanEnabled)

	// Verify the getInsightsAccounts invoke outputs
	accountsCount := upResult.Outputs["accountsCount"].Value.(float64)
	assert.GreaterOrEqual(t, accountsCount, float64(1))

	createdAccountInList := upResult.Outputs["createdAccountInList"].Value.(bool)
	assert.True(t, createdAccountInList)
}

func TestNodejsRbacExample(t *testing.T) {
	// Requires the Custom Roles feature to be enabled on the test
	// organization. CreateRole returns a feature-flag error otherwise,
	// which fails the test loudly — that's the signal we want.
	orgName := getOrgName()
	const fixtureUser = "service-provider-example-user"

	// Snapshot the fixture user's role before the test mutates it, and
	// restore it on cleanup. OrganizationRole.Delete with force=true
	// *should* revoke the assignment on destroy, but if a teardown leaves
	// the user on the test's role this hook still gets the org back to a
	// known state.
	t.Cleanup(snapshotFixtureOrgMember(t, orgName, fixtureUser))

	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-rbac"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "organizationName", orgName)
	test.SetConfig(t, "targetUsername", fixtureUser)
	test.SetConfig(t, "nameSuffix", generateRandomFiveDigits())

	up := test.Up(t)
	if adopted, ok := up.Outputs["memberAdopted"]; ok {
		assert.Equal(t, true, adopted.Value, "expected OrganizationMember to adopt the existing membership")
	}
	// Env-scoped role wiring: the env's UUID must flow into the role's
	// permission tree. Empty UUID → metadata fetch silently skipped.
	if envID, ok := up.Outputs["scopedEnvironmentId"]; ok {
		assert.NotEmpty(t, envID.Value, "expected Environment.environmentId to be populated")
	}
	if scopedRoleID, ok := up.Outputs["scopedRoleId"]; ok {
		assert.NotEmpty(t, scopedRoleID.Value, "expected env-scoped OrganizationRole to be created")
	}
	preview := test.Preview(t)
	assertpreview.HasNoChanges(t, preview)
	refresh := test.Refresh(t)
	assertrefresh.HasNoChanges(t, refresh)
	test.Destroy(t)
}
