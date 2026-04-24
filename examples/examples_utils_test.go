package examples

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
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

// resetFixtureOrgMember snaps the fixture user back to the default
// built-in role. Used as a test cleanup to ensure the shared test org is
// restored to a known state even if the Pulumi-level teardown leaves
// any custom role assignment dangling.
func resetFixtureOrgMember(orgName, userName string) error {
	token := os.Getenv("PULUMI_ACCESS_TOKEN")
	apiURL := os.Getenv("PULUMI_BACKEND_URL")
	if token == "" || apiURL == "" {
		return nil // no creds → nothing to clean up (local dev without env)
	}
	client, err := pulumiapi.NewClient(&http.Client{Timeout: 60 * time.Second}, token, apiURL)
	if err != nil {
		return err
	}
	return client.UpdateOrgMemberRole(context.Background(), orgName, userName, "member", nil)
}
