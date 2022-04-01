package examples

import (
	"os"
	"path"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestAccessTokenExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:       true,
		SkipRefresh: true,
		Dir:         path.Join(cwd, ".", "access-tokens"),
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}
