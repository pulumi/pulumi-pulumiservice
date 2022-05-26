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
