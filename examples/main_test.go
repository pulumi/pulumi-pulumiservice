package examples

import (
	"os"
	"testing"

	pgo "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/embedded"
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
// pulumitest harness. The provider is built directly from the
// embedded OpenAPI spec + resource map — the same inputs the
// production binary uses — so the integration suite exercises the
// real schema and metadata paths without any pre-built artifacts.
func inMemoryProvider() opttest.Option {
	return opttest.AttachProviderServer("pulumiservice",
		func(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
			prov, err := psp.New(embedded.Spec(), embedded.ResourceMap())
			if err != nil {
				return nil, err
			}
			return pgo.RawServer("pulumiservice", "2.0.0-alpha.1+dev", prov)(nil)
		})
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
