//go:build dotnet || all
// +build dotnet all

package examples

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestDotnetSmokeExample(t *testing.T) {
	runSmokeTest(t, "cs",
		// `pulumi install` runs before the reference below is added, so it
		// can't resolve the local SDK. `pulumi up` builds the program instead.
		opttest.SkipInstall(),
		opttest.DotNetReference("Pulumi.PulumiService", getCwd(t), "..", "sdk", "dotnet"),
	)
}
