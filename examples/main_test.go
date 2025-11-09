package examples

import (
	"os"
	"testing"
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
