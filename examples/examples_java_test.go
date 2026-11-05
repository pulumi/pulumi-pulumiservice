//go:build java || all
// +build java all

package examples

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

// The Java program resolves the pulumiservice SDK from the local Maven
// repository, where `make install_java_sdk` publishes it.
func TestJavaSmokeExample(t *testing.T) {
	runSmokeTest(t, "java", nil,
		// `pulumi install` tries to download the provider plugin at the SDK's
		// dev version, which isn't published. `pulumi up` builds the program
		// instead, and uses the in-memory provider.
		opttest.SkipInstall(),
	)
}
