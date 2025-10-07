package examples

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testOwner := os.Getenv("PULUMI_TEST_OWNER")
	if testOwner == "" {
		_ = os.Setenv("PULUMI_TEST_OWNER", "service-provider-test-org")
	}
	_ = os.Setenv("PULUMI_TEST_USE_SERVICE", "true")
	m.Run()
}
