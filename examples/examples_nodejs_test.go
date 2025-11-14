//go:build nodejs || all
// +build nodejs all

package examples

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestAccessTokenExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "ts-access-tokens"),
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestStackTagsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "ts-stack-tags"),
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestDeploymentSettingsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Config: map[string]string{
			"my_secret": "my_secret_value",
			"password":  "my_password",
		},
		Quick: true,
		Dir:   path.Join(cwd, ".", "ts-deployment-settings"),
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestTeamStackPermissionsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "ts-team-stack-permissions"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestTeamsExample(t *testing.T) {
	cwd := getCwd(t)

	// Create a pulumitest instance with in-memory provider using the existing ts-teams directory
	testPulumiProgram(t, pulumitest.NewPulumiTest(t,
		filepath.Join(cwd, "ts-teams"),
		inMemoryProvider(),
		opttest.YarnLink("@pulumi/pulumiservice"),
	))
}

func TestNodejsWebhookExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "ts-webhooks"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestNodejsSchedulesExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:       path.Join(cwd, ".", "ts-schedules"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestNodejsEnvironmentsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "ts-environments"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
		ExpectRefreshChanges: true,
	})
}

func TestNodejsTemplateSourcesExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "ts-template-source"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestNodejsEnvironmentsFileAssetExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "ts-environments-file-asset"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestNodejsOidcIssuerExample(t *testing.T) {
	cwd := getCwd(t)
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "ts-oidc-issuer"),
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}

func TestNodejsApprovalRulesExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "ts-approval-rules"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}
