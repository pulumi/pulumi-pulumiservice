//go:build python || all

package examples

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestPythonTeamsExample(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:         path.Join(getCwd(t), "py-teams"),
		SkipRefresh: true,
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}

func TestPythonDeploymentSettingsExample(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:         path.Join(getCwd(t), "py-deployment-settings"),
		SkipRefresh: true,
		Config: map[string]string{
			"my-secret": "my-secret-value",
		},
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}

func TestPythonEnvironmentsExample(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:         path.Join(getCwd(t), "py-environments"),
		SkipRefresh: true,
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})
}
