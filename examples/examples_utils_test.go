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
	return fmt.Sprintf("%05d", rand.Intn(100000))
}

func getOrgName(t *testing.T) string {
	orgName := os.Getenv("PULUMI_TEST_OWNER")
	if orgName == "" {
		orgName = "service-provider-test-org"
	}
	return orgName
}

func getOwnerName(t *testing.T) string {
	// For the Service resource, the owner is a user, not an org
	// This should be a valid username in the org
	ownerName := os.Getenv("PULUMI_TEST_SERVICE_OWNER")
	if ownerName == "" {
		t.Skip("PULUMI_TEST_SERVICE_OWNER not set, skipping Service example test")
	}
	return ownerName
}
