//go:build yaml

package examples

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/require"
)

// TestYamlOidcIssuerPreviewUnknownURL verifies that OidcIssuer previews cleanly
// when its url input is an unknown output. The pre-infer resource panicked
// reading url.StringValue() on an unknown value during preview.
func TestYamlOidcIssuerPreviewUnknownURL(t *testing.T) {
	program := `name: recreate-pulumiservice-issue
runtime: yaml
resources:
  dummy:
    type: random:RandomString
    properties:
      length: 10
  issuer:
    type: pulumiservice:OidcIssuer
    properties:
      name: foo
      organization: bar
      url: ${dummy.result}
`
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Pulumi.yaml"), []byte(program), 0o600))

	t.Setenv("PULUMI_BACKEND_URL", "file://"+t.TempDir())
	t.Setenv("PULUMI_CONFIG_PASSPHRASE", "")
	t.Setenv("PULUMI_ACCESS_TOKEN", "pul-dummy-token")

	test := pulumitest.NewPulumiTest(t, dir,
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.SkipInstall(),
	)
	test.Preview(t)
}
