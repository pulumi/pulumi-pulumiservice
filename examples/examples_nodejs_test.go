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

func TestStackTagsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:       true,
		SkipRefresh: true,
		Dir:         path.Join(cwd, ".", "ts-stack-tags"),
		Config: map[string]string{
			"orgName": "service-provider-test-org",
		},
		Dependencies: []string{
			"@pulumi/pulumiservice",
		},
	})
}
