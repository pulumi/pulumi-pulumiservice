//go:build nodejs || all
// +build nodejs all

package examples

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestNodejsSmokeExample(t *testing.T) {
	runSmokeTest(t, "ts", opttest.YarnLink("@pulumi/pulumiservice"))
}
