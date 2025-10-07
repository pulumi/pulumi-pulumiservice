package examples

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Only set PULUMI_TEST_OWNER if not already set by environment
	if os.Getenv("PULUMI_TEST_OWNER") == "" {
		_ = os.Setenv("PULUMI_TEST_OWNER", "service-provider-test-org")
	}
	_ = os.Setenv("PULUMI_TEST_USE_SERVICE", "true")
	m.Run()
}
