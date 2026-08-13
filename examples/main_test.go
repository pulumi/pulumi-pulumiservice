package examples

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
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

func inMemoryProvider() opttest.Option {
	provider := func(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
		return psp.MakeProvider(nil, "pulumiservice", "1.0.0")
	}
	return opttest.AttachProviderServer("pulumiservice", provider)
}

// yarnInstall runs `yarn install` in dir. The PolicyPack provider boots the
// pack's language analyzer to introspect policies, which needs the pack's
// runtime deps installed (same prerequisite as `pulumi policy publish`).
func yarnInstall(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("yarn", "install", "--non-interactive")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("yarn install in %s failed: %v\n%s", dir, err, out)
	}
}

// runPulumiTest performs the same basic steps as
// [github.com/pulumi/pulumi/pkg/v3/testing/integration.ProgramTest].
func runPulumiTest(t *testing.T, test *pulumitest.PulumiTest) auto.UpResult {
	return runPulumiTestAfterUp(t, test, nil)
}

// runPulumiTestAfterUp is runPulumiTest with a hook that runs against the live
// stack right after `up`, for assertions that need the stack state before it is
// destroyed.
func runPulumiTestAfterUp(
	t *testing.T,
	test *pulumitest.PulumiTest,
	afterUp func(t *testing.T, test *pulumitest.PulumiTest),
) auto.UpResult {
	// Run the Pulumi program
	upResult := test.Up(t)

	if afterUp != nil {
		afterUp(t, test)
	}

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

// stateInputs returns the state inputs of the single resource of the given type.
func stateInputs(t *testing.T, test *pulumitest.PulumiTest, resourceType string) map[string]any {
	t.Helper()

	var deployment apitype.DeploymentV3
	exported := test.ExportStack(t)
	require.NoError(t, json.Unmarshal(exported.Deployment, &deployment))

	for _, res := range deployment.Resources {
		if string(res.Type) == resourceType {
			return res.Inputs
		}
	}
	t.Fatalf("no %s resource found in the exported stack", resourceType)
	return nil
}

// assertStateSecret asserts that the value at path within inputs is stored as an
// encrypted secret rather than plaintext.
func assertStateSecret(t *testing.T, inputs map[string]any, path ...string) {
	t.Helper()

	value := any(inputs)
	for _, key := range path {
		object, ok := value.(map[string]any)
		require.Truef(t, ok, "%v: %q is not nested under an object", path, key)
		value, ok = object[key]
		require.Truef(t, ok, "%v: %q is missing from the state inputs", path, key)
	}

	object, ok := value.(map[string]any)
	require.Truef(t, ok, "%v is %#v, not a secret envelope", path, value)
	assert.Equalf(t, resource.SecretSig, object[resource.SigKey],
		"%v is not stored as a secret in state", path)
}
