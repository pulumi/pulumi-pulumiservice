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
	ownerName := os.Getenv("PULUMI_TEST_SERVICE_OWNER_USER_NAME")
	if ownerName == "" {
		t.Skip("PULUMI_TEST_SERVICE_OWNER_USER_NAME not set, skipping Service example test")
	}
	return ownerName
}
