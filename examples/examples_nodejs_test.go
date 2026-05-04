//go:build nodejs || all
// +build nodejs all

package examples

import (
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
	runPulumiTest(t, test)
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
	runPulumiTest(t, test)
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
