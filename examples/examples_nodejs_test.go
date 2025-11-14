//go:build nodejs || all
// +build nodejs all

package examples

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestAccessTokenExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-access-tokens"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
	)
	runPulumiTest(t, test)
}

func TestStackTagsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-stack-tags"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
	)
	runPulumiTest(t, test)
}

func TestDeploymentSettingsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-deployment-settings"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
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
	)
	runPulumiTest(t, test)
}

func TestNodejsWebhookExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-webhooks"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestNodejsSchedulesExample(t *testing.T) {
	digits := generateRandomFiveDigits()
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-schedules"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
		opttest.StackName("test-stack-"+digits),
	)
	test.SetConfig(t, "digits", digits)
	runPulumiTest(t, test)
}

func TestNodejsEnvironmentsExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-environments"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
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
	)
	runPulumiTest(t, test)
}

func TestNodejsApprovalRulesExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "ts-approval-rules"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.YarnLink("@pulumi/pulumiservice"),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}
