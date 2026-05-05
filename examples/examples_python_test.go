//go:build python || all
// +build python all

package examples

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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

	rbacDir := path.Join(getCwd(t), "py-rbac")

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: rbacDir,
		Config: map[string]string{
			"organizationName": orgName,
			"digits":           generateRandomFiveDigits(),
		},
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
		ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			// Env-scoped role wiring must populate. An empty UUID would
			// mean the metadata fetch silently skipped, leaving the
			// role's `identity` unresolved.
			if envID, ok := stack.Outputs["scopedEnvironmentId"]; ok {
				assert.NotEmpty(t, envID, "expected Environment.environment_id to be populated")
			}
			if scopedRoleID, ok := stack.Outputs["scopedRoleId"]; ok {
				assert.NotEmpty(t, scopedRoleID, "expected env-scoped OrganizationRole to be created")
			}
		},
		// Re-apply the same program to pin the descriptor round-trip:
		// helper output (`__type` at every level) → API → role record
		// → Read response → state → next preview must converge to no
		// changes. Drift in the Group(Condition) wrap/unwrap heuristic
		// would surface here as a `~` on `permissions` instead.
		EditDirs: []integration.EditDir{
			{
				Dir:             rbacDir,
				Additive:        true, // preserve venv from initial deploy
				ExpectNoChanges: true,
			},
		},
	})
}
