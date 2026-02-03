package examples

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}
	return cwd
}

func generateRandomFiveDigits() string {
	return fmt.Sprintf("%05d", rand.Intn(100000)) //nolint:gosec // For testing.
}

// getOrgName returns the organization name from PULUMI_TEST_OWNER env var,
// or defaults to "service-provider-test-org" if not set
func getOrgName() string {
	if org := os.Getenv("PULUMI_TEST_OWNER"); org != "" {
		return org
	}
	return "service-provider-test-org"
}
