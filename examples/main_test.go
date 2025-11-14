package examples

import (
	"os"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	psp "github.com/pulumi/pulumi-pulumiservice/provider/pkg/provider"
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

func inMemoryProvider() opttest.Option {
	provider := func(pt providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
		return psp.MakeProvider(nil, "pulumiservice", "1.0.0")
	}
	return opttest.AttachProviderServer("pulumiservice", provider)
}

// runPulumiTest performs the same basic steps as
// [github.com/pulumi/pulumi/pkg/v3/testing/integration.ProgramTest].
func runPulumiTest(t *testing.T, test *pulumitest.PulumiTest) {
	// Run the Pulumi program
	test.Up(t)

	// Run preview to ensure no changes after initial deployment
	previewResult := test.Preview(t)
	assertpreview.HasNoChanges(t, previewResult)

	// Run refresh to ensure no changes
	refreshResult := test.Refresh(t)
	assertrefresh.HasNoChanges(t, refreshResult)

	// Clean up - destroy the stack
	test.Destroy(t)
}
