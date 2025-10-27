//go:build nodejs || all
// +build nodejs all

package examples

import (
	"os"
	"path"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/assert"
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
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "ts-teams"),
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
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

func TestNodejsInsightsAccountMethodsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true, // Skip extra preview/update/refresh to avoid 409 conflicts from calling triggerScan() multiple times
		Dir:   path.Join(cwd, ".", "ts-insights-account-methods"),
		Config: map[string]string{
			"digits":           digits,
			"organizationName": getOrgName(),
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
		ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
			// Verify resource method outputs are present
			outputs := stackInfo.Outputs

			// Verify triggerScan() method outputs
			// scanId is optional (HTTP 204 responses don't include it)
			if scanId, ok := outputs["scanId"].(string); ok {
				assert.NotEmpty(t, scanId, "scanId should not be empty if present")
			}

			// scanStatus is required
			scanStatus, ok := outputs["scanStatus"].(string)
			assert.True(t, ok, "scanStatus should be a string")
			assert.NotEmpty(t, scanStatus, "scanStatus should not be empty")
			assert.Contains(t, []string{"queued", "running", "succeeded", "failed"}, scanStatus, "scanStatus should be a valid workflow status")

			// scanTimestamp is optional
			if timestamp, ok := outputs["scanTimestamp"].(string); ok {
				assert.NotEmpty(t, timestamp, "scanTimestamp should not be empty if present")
			}
		},
	})
}
