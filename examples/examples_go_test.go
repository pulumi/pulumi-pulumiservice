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
	})
	integration.ProgramTest(t, &testOpts)
}

func TestGoEnvironmentsExample(t *testing.T) {
	testOpts := getGoBaseOptions(t).With(integration.ProgramTestOptions{
		Verbose:     true,
		Dir:         filepath.Join(getCwd(t), "go-environments"),
		SkipRefresh: true,
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
