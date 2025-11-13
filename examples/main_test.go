package examples

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

// The default test org to use.
var ServiceProviderTestOrg string = "service-provider-test-org"

func TestMain(m *testing.M) {
	// Set default test owner if not already set
	if testOwner := os.Getenv("PULUMI_TEST_OWNER"); testOwner == "" {
		if err := os.Setenv("PULUMI_TEST_OWNER", ServiceProviderTestOrg); err != nil {
			panic("failed to set PULUMI_TEST_OWNER: " + err.Error())
		}
	} else {
		ServiceProviderTestOrg = testOwner
	}
	if err := os.Setenv("PULUMI_TEST_USE_SERVICE", "true"); err != nil {
		panic("failed to set PULUMI_TEST_USE_SERVICE: " + err.Error())
	}

	m.Run()
}

// getBaseOptions returns common ProgramTestOptions with GOCOVERDIR set for coverage collection.
// This ensures that when the provider binary (built with -cover) runs during tests,
// it writes coverage data to the specified directory.
func getBaseOptions() integration.ProgramTestOptions {
	opts := integration.ProgramTestOptions{}

	// If PROVIDER_GOCOVERDIR is set, pass it as GOCOVERDIR to the provider binary
	// We use PROVIDER_GOCOVERDIR instead of GOCOVERDIR because go test -cover sets its own
	// GOCOVERDIR for test coverage, and we want to collect provider coverage separately
	if coverDir := os.Getenv("PROVIDER_GOCOVERDIR"); coverDir != "" {
		opts.Env = []string{"GOCOVERDIR=" + coverDir}
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		coverageDir := filepath.Join(cwd, "..", "bin", "coverage")
		coverageBinaryPath := filepath.Join(coverageDir, "pulumi-resource-pulumiservice")
		if _, err := os.Stat(coverageBinaryPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				panic("Missing coverage binary at " + coverageBinaryPath)
			} else {
				panic(err)
			}
		}
		opts.LocalProviders = []integration.LocalDependency{
			{
				Package: "pulumiservice",
				Path:    coverageDir,
			},
		}
	}

	return opts
}
