//go:build dotnet || all

package examples

import (
	"path"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestDotnetTeamsExamples(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "cs-teams"),
		Dependencies: []string{
			"Pulumi.Random",t
			"Pulumi.PulumiService",
		},
	})
}
