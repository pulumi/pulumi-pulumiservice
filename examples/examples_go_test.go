//go:build go || all
// +build go all

package examples

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestGoSmokeExample(t *testing.T) {
	runSmokeTest(t, "go", nil,
		// `pulumi install` runs before the reference below is added, so it
		// can't resolve the local SDK. `pulumi up` builds the program instead.
		opttest.SkipInstall(),
		opttest.GoModReplacement("github.com/pulumi/pulumi-pulumiservice/sdk", getCwd(t), "..", "sdk"),
	)
}
