//go:build yaml

package examples

import (
	"os"
	"path"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestYamlTeamsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:       true,
		SkipRefresh: true,
		Dir:         path.Join(cwd, ".", "yaml-teams"),
		// don't prepare project at all, not required for yaml
		PrepareProject: func(_ *engine.Projinfo) error {
			return nil
		},
	})
}
