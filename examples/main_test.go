package examples

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Set default test owner if not already set
	if os.Getenv("PULUMI_TEST_OWNER") == "" {
		if err := os.Setenv("PULUMI_TEST_OWNER", "service-provider-test-org"); err != nil {
			panic("failed to set PULUMI_TEST_OWNER: " + err.Error())
		}
	}
	if err := os.Setenv("PULUMI_TEST_USE_SERVICE", "true"); err != nil {
		panic("failed to set PULUMI_TEST_USE_SERVICE: " + err.Error())
	}
	m.Run()
}
