//go:build python || all
// +build python all

package examples

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestPythonTeamsExample(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "py-teams"),
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}

func TestPythonDeploymentSettingsExample(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "py-deployment-settings"),
		Config: map[string]string{
			"my-secret": "my-secret-value",
		},
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}

func TestPythonEnvironmentsExample(t *testing.T) {
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "py-environments"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}

func TestPythonApprovalRulesExample(t *testing.T) {
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "py-approval-rules"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}

func TestPythonRbacExample(t *testing.T) {
	// Requires the Custom Roles feature to be enabled on the test
	// organization. If it isn't, CreateRole returns a feature-flag error
	// and the test fails loudly, which is what we want to learn.
	orgName := getOrgName()
	const fixtureUser = "service-provider-example-user"

	// Snapshot the fixture user's role before the test mutates it, and
	// restore it on cleanup. OrganizationMember flips this user's role
	// to "admin" during the test; the restore puts them back on whatever
	// role they had pre-test (rather than blindly resetting to "member").
	t.Cleanup(snapshotFixtureOrgMember(t, orgName, fixtureUser))

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "py-rbac"),
		Config: map[string]string{
			"organizationName": orgName,
			"digits":           generateRandomFiveDigits(),
		},
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}
