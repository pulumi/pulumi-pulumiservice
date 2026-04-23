package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	psp "github.com/pulumi/pulumi-pulumiservice/provider/pkg/provider"
)

// The default test org to use.
var ServiceProviderTestOrg = "service-provider-test-org"

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

// inMemoryProvider attaches an in-process v2 provider server to the
// pulumitest harness. Schema + metadata are read from the canonical binary
// embed paths, and the full custom-resource set is registered — this
// mirrors what the shipped provider binary does at startup, minus the
// gRPC plugin handshake.
func inMemoryProvider() opttest.Option {
	provider := func(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
		schemaBytes, err := os.ReadFile(schemaPath())
		if err != nil {
			return nil, fmt.Errorf("reading embedded schema: %w", err)
		}
		metadataBytes, err := os.ReadFile(metadataPath())
		if err != nil {
			return nil, fmt.Errorf("reading embedded metadata: %w", err)
		}

		return psp.New("pulumiservice", "2.0.0-alpha.1+dev", schemaBytes, metadataBytes)
	}
	return opttest.AttachProviderServer("pulumiservice", provider)
}

// schemaPath / metadataPath locate the generator-emitted artifacts that
// the shipped binary embeds. Tests read them fresh so schema edits take
// effect without rebuilding the binary.
func schemaPath() string {
	return filepath.Join(repoRoot(), "provider", "cmd", "pulumi-resource-pulumiservice", "schema.json")
}
func metadataPath() string {
	return filepath.Join(repoRoot(), "provider", "cmd", "pulumi-resource-pulumiservice", "metadata.json")
}

// repoRoot walks up from the examples/ directory to find the repo root.
// Tests always run with cwd=examples/, so this is a single parent hop.
func repoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(cwd)
}

// runPulumiTest performs the same basic steps as
// [github.com/pulumi/pulumi/pkg/v3/testing/integration.ProgramTest].
func runPulumiTest(t *testing.T, test *pulumitest.PulumiTest) auto.UpResult {
	// Run the Pulumi program
	upResult := test.Up(t)

	// Run preview to ensure no changes after initial deployment
	previewResult := test.Preview(t)
	assertpreview.HasNoChanges(t, previewResult)

	// Run refresh to ensure no changes
	refreshResult := test.Refresh(t)
	assertrefresh.HasNoChanges(t, refreshResult)

	// Clean up - destroy the stack
	test.Destroy(t)

	return upResult
}
