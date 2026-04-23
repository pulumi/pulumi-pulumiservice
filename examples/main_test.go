package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/gen"
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
// pulumitest harness. Schema + metadata are generated from the pinned
// OpenAPI spec and resource-map.yaml at test-setup time — the tests
// don't depend on any pre-built artifacts under bin/ or
// provider/cmd/..., so `go test ./...` works against a clean checkout
// without first running `make v2_gen`.
func inMemoryProvider() opttest.Option {
	return opttest.AttachProviderServer("pulumiservice",
		func(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
			schemaBytes, metadataBytes, err := generateSchemaAndMetadata()
			if err != nil {
				return nil, err
			}
			return psp.New("pulumiservice", "2.0.0-alpha.1+dev", schemaBytes, metadataBytes)
		})
}

// generateSchemaAndMetadata runs the generator against the committed
// spec + resource map. Memoized so the (non-trivial) emission only
// happens once per test binary run.
var (
	genOnce     sync.Once
	genSchema   []byte
	genMetadata []byte
	genErr      error
)

func generateSchemaAndMetadata() ([]byte, []byte, error) {
	genOnce.Do(func() {
		spec := filepath.Join(repoRoot(), "provider", "spec", "openapi_public.json")
		resMap := filepath.Join(repoRoot(), "provider", "resource-map.yaml")
		if genSchema, genErr = gen.EmitSchema(spec, resMap); genErr != nil {
			genErr = fmt.Errorf("emitting schema: %w", genErr)
			return
		}
		if genMetadata, genErr = gen.EmitMetadata(spec, resMap); genErr != nil {
			genErr = fmt.Errorf("emitting metadata: %w", genErr)
			return
		}
	})
	return genSchema, genMetadata, genErr
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
