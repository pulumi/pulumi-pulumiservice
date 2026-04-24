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

// snapshotFixtureOrgMember reads the user's current role from the org and
// returns a closure that restores it. Pass the closure to t.Cleanup so the
// shared test org returns to its pre-test state regardless of how the test
// exited.
//
// This replaces the older "always reset to member" cleanup, which was
// destructive for fixture users whose pre-test role was anything other than
// the default — e.g. snapping pulumi-bot back to "member" would have broken
// every subsequent test that relied on its admin-level access.
//
// Concurrency: this is a per-process snapshot. Tests in the same go-test
// invocation that mutate the same fixture user are still serialized by the
// absence of t.Parallel(). It does not protect against parallel invocations
// across CI shards or local terminals. The robust fix for cross-shard races
// is a per-test fixture user; tracked as a follow-up.
func snapshotFixtureOrgMember(t *testing.T, orgName, userName string) func() {
	t.Helper()
	token := os.Getenv("PULUMI_ACCESS_TOKEN")
	apiURL := os.Getenv("PULUMI_BACKEND_URL")
	if token == "" || apiURL == "" {
		return func() {} // no creds → nothing to snapshot (local dev without env)
	}
	client, err := pulumiapi.NewClient(&http.Client{Timeout: 60 * time.Second}, token, apiURL)
	if err != nil {
		t.Logf("snapshot fixture user: client init failed: %v", err)
		return func() {}
	}
	member, err := client.GetOrgMember(context.Background(), orgName, userName)
	if err != nil {
		t.Logf("snapshot fixture user: lookup failed: %v", err)
		return func() {}
	}
	if member == nil {
		t.Logf("snapshot fixture user: %q not found in %q", userName, orgName)
		return func() {}
	}

	// Capture both representations: fgaRoleId is authoritative for both
	// built-in and custom roles, but legacy responses without fgaRole need
	// the role-string fallback.
	var origFGARoleID *string
	if member.FGARole != nil {
		id := member.FGARole.ID
		origFGARoleID = &id
	}
	origRole := member.Role

	return func() {
		var restoreErr error
		if origFGARoleID != nil {
			restoreErr = client.UpdateOrgMemberRole(
				context.Background(), orgName, userName, "", origFGARoleID,
			)
		} else if origRole != "" {
			restoreErr = client.UpdateOrgMemberRole(
				context.Background(), orgName, userName, origRole, nil,
			)
		}
		if restoreErr != nil {
			t.Logf("restore fixture user role: %v", restoreErr)
		}
	}
}
