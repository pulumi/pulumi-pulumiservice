//go:build go || all
// +build go all

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestGoTeamsExample(t *testing.T) {
	testOpts := getGoBaseOptions(t).With(integration.ProgramTestOptions{
		Verbose:     true,
		Dir:         filepath.Join(getCwd(t), "go-teams"),
		SkipRefresh: true,
		Dependencies: []string{
			"github.com/pulumi/pulumi-pulumiservice/sdk",
			"github.com/pulumi/pulumi-random/sdk/v4",
		},
	})
	integration.ProgramTest(t, &testOpts)
}

func getGoBaseOptions(t *testing.T) integration.ProgramTestOptions {
	return integration.ProgramTestOptions{
		Dependencies: []string{
			"github.com/pulumi/pulumi-pulumiservice/sdk",
		},
	}
}
