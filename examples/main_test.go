package examples

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("PULUMI_TEST_OWNER", "service-provider-test-org")
	os.Setenv("PULUMI_TEST_USE_SERVICE", "true")
	m.Run()
}
