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

	// Safety net: OrganizationMember flips this user's built-in role to
	// "admin" during the test. If teardown leaves them on any role other
	// than the default, snap them back here so we don't leak state.
	t.Cleanup(func() {
		if err := resetFixtureOrgMember(orgName, fixtureUser); err != nil {
			t.Logf("cleanup: could not reset fixture user role: %v", err)
		}
	})

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "py-rbac"),
		Config: map[string]string{
			"organizationName": orgName,
			"digits":           generateRandomFiveDigits(),
		},
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
		// Enabling custom roles on the team causes the service to add the
		// caller as a team member, which shows up on refresh as drift on
		// the pre-existing Team resource. Same follow-up the yaml-rbac
		// test is waiting on — not in scope here.
		ExpectRefreshChanges: true,
	})
}
