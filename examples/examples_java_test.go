//go:build java || all
// +build java all

package examples

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestJavaTeamsExamples(t *testing.T) {
	test := getJavaBase(t, "java-teams", integration.ProgramTestOptions{})

	integration.ProgramTest(t, &test)
}

func getJavaBase(t *testing.T, dir string, testSpecificOptions integration.ProgramTestOptions) integration.ProgramTestOptions {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		panic(err)
	}
	opts := integration.ProgramTestOptions{
		Dir: filepath.Join(getCwd(t), dir),
		Env: []string{fmt.Sprintf("PULUMI_REPO_ROOT=%s", repoRoot)},
		PrepareProject: func(*engine.Projinfo) error {
			return nil // needed because defaultPrepareProject does not know about java
		},
	}
	opts = opts.With(getBaseOptions()).With(testSpecificOptions)
	return opts
}

func getBaseOptions() integration.ProgramTestOptions {
	return integration.ProgramTestOptions{
		Dependencies: []string{"@pulumi/pulumi"},
	}
}
