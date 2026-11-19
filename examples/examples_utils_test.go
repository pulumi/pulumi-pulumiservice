package examples

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}
	return cwd
}

const serializedTestsEnv = "PULUMI_TEST_SERIALIZED"

func skipUnlessSerializedTests(t *testing.T) {
	t.Helper()
	if os.Getenv(serializedTestsEnv) == "" {
		t.Skipf("runs in the serialized shared-org job; set %s to run it", serializedTestsEnv)
	}
}

func generateRandomFiveDigits() string {
	return fmt.Sprintf("%05d", rand.Intn(100000)) //nolint:gosec // For testing.
}

// randomStackName returns a per-test-call stack name with enough entropy
// to avoid collisions across CI shards, parallel PR runs, and previous
// runs that left orphaned stacks behind. Use it via opttest.StackName so
// every pulumitest invocation gets a distinct stack name; otherwise
// pulumitest defaults to "test" and concurrent runs hit "stack already
// exists" on init.
func randomStackName() string {
	return "test-" + uuid.NewString()[:8]
}

// uniquePolicyPackName returns a policy pack name no earlier run has used.
// Pulumi Cloud refuses to publish a name and version tag that was ever
// published, even after the version is deleted, so a name from the five-digit
// space eventually collides with a past run.
func uniquePolicyPackName() string {
	return "example-pack-" + uuid.NewString()
}

// tlsThumbprint returns the SHA256 fingerprint (hex) of the TLS certificate
// currently served by host — the value Pulumi Cloud calculates when an OIDC
// issuer is registered. Issuers rotate certificates (GitHub serves 90-day
// Let's Encrypt certs as of July 2026), so tests compute thumbprints at
// runtime instead of pinning literals that go stale on every rotation.
// Panics on failure so it is usable from apiCase Config funcs, which don't
// receive a *testing.T.
func tlsThumbprint(host string) string {
	conn, err := tls.Dial("tcp", host+":443", nil)
	if err != nil {
		panic(fmt.Sprintf("computing TLS thumbprint for %s: %v", host, err))
	}
	defer conn.Close()
	return fmt.Sprintf("%x", sha256.Sum256(conn.ConnectionState().PeerCertificates[0].Raw))
}

// oidcIssuerTestConfig returns issuer URL and thumbprint config for the OIDC
// issuer examples. Pulumi Cloud lets each issuer URL be registered only once
// per org, so runs that overlapped, within a PR or across PRs, failed with
// "issuer ... already registered" when they all used the examples' default
// URLs. GitHub's issuer serves OIDC discovery under any path, so each call
// returns its own https://token.actions.githubusercontent.com/<random> URLs,
// for the Pulumi issuer as well as the GitHub one. The tests cover issuer
// CRUD only, never token exchange, so which issuer the URL points at doesn't
// matter.
func oidcIssuerTestConfig() map[string]string {
	base := "https://token.actions.githubusercontent.com/pulumiservice-test-" + uuid.NewString()[:8]
	thumbprint := tlsThumbprint("token.actions.githubusercontent.com")
	return map[string]string{
		"pulumiIssuerUrl":  base + "-pulumi",
		"pulumiThumbprint": thumbprint,
		"githubIssuerUrl":  base + "-github",
		"githubThumbprint": thumbprint,
	}
}

// getOrgName returns the organization name from PULUMI_TEST_OWNER env var,
// or defaults to "service-provider-test-org" if not set
func getOrgName() string {
	if org := os.Getenv("PULUMI_TEST_OWNER"); org != "" {
		return org
	}
	return "service-provider-test-org"
}

// getInsightsRoleArn returns the AWS IAM role ARN that the Insights example
// tests embed in their generated ESC environments. The role's trust policy
// must allow `sts:AssumeRoleWithWebIdentity` from Pulumi Cloud's OIDC issuer
// for the test org; see infra/test-aws-oidc/ for the program that
// provisions it. Locally, the value to set is the `roleArn` output of that
// stack:
//
//	pulumi -C infra/test-aws-oidc stack -s ci output roleArn
//
// Fails the test if PULUMI_TEST_AWS_INSIGHTS_ROLE_ARN is unset rather than
// falling back to a placeholder, since after pulumi-service#41970 the
// Insights API rejects environments without working provider credentials at
// account-create time.
func getInsightsRoleArn(t *testing.T) string {
	t.Helper()
	arn := os.Getenv("PULUMI_TEST_AWS_INSIGHTS_ROLE_ARN")
	if arn == "" {
		t.Fatal("PULUMI_TEST_AWS_INSIGHTS_ROLE_ARN is required for Insights example tests; " +
			"set it to the `roleArn` output of the infra/test-aws-oidc stack")
	}
	return arn
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
	// the role-string fallback. Treat FGARole with an empty ID the same as
	// FGARole=nil; that's what we see when the snapshot races into another
	// shard's mid-flight role teardown and observes an inconsistent
	// intermediate state.
	var origFGARoleID *string
	if member.FGARole != nil && member.FGARole.ID != "" {
		id := member.FGARole.ID
		origFGARoleID = &id
	}
	origRole := member.Role

	return func() {
		var restoreErr error
		switch {
		case origFGARoleID != nil:
			restoreErr = client.UpdateOrgMemberRole(
				context.Background(), orgName, userName, "", origFGARoleID,
			)
		case origRole != "":
			restoreErr = client.UpdateOrgMemberRole(
				context.Background(), orgName, userName, origRole, nil,
			)
		default:
			t.Logf("restore fixture user role: snapshot captured no role; skipping restore")
			return
		}
		if restoreErr != nil {
			t.Logf("restore fixture user role: %v", restoreErr)
		}
	}
}

// setSecretConfig sets an encrypted config value on the test's stack.
// pulumitest's SetConfig only sets plaintext values, which a program that
// declares the key as secret rejects.
func setSecretConfig(t *testing.T, test *pulumitest.PulumiTest, key, value string) {
	t.Helper()
	err := test.CurrentStack().SetConfig(test.Context(), key, auto.ConfigValue{Value: value, Secret: true})
	require.NoError(t, err, "setting secret config %q", key)
}

// templateSourceDirs are directories of github.com/pulumi/templates: the base
// language templates and the architecture templates. Pulumi Cloud allows only
// one template source per URL in an organization, so each run picks one at
// random to avoid colliding with a concurrent run or with a source that an
// earlier run left behind.
var templateSourceDirs = []string{
	"bun", "csharp", "fsharp", "go", "hcl", "java-gradle", "java-jbang", "java", "javascript",
	"python", "typescript", "visualbasic", "yaml",
	"container-aws-csharp", "container-aws-go", "container-aws-hcl", "container-aws-python",
	"container-aws-typescript", "container-aws-yaml",
	"container-azure-csharp", "container-azure-go", "container-azure-hcl", "container-azure-python",
	"container-azure-typescript",
	"container-gcp-csharp", "container-gcp-go", "container-gcp-hcl", "container-gcp-python",
	"container-gcp-typescript",
	"helm-kubernetes-csharp", "helm-kubernetes-go", "helm-kubernetes-hcl", "helm-kubernetes-python",
	"helm-kubernetes-typescript", "helm-kubernetes-yaml",
	"kubernetes-aws-csharp", "kubernetes-aws-go", "kubernetes-aws-hcl", "kubernetes-aws-python",
	"kubernetes-aws-typescript", "kubernetes-aws-yaml",
	"kubernetes-azure-csharp", "kubernetes-azure-go", "kubernetes-azure-hcl",
	"kubernetes-azure-python", "kubernetes-azure-typescript", "kubernetes-azure-yaml",
	"kubernetes-gcp-csharp", "kubernetes-gcp-go", "kubernetes-gcp-hcl", "kubernetes-gcp-python",
	"kubernetes-gcp-typescript", "kubernetes-gcp-yaml",
	"serverless-aws-csharp", "serverless-aws-go", "serverless-aws-hcl", "serverless-aws-python",
	"serverless-aws-typescript", "serverless-aws-yaml",
	"serverless-azure-csharp", "serverless-azure-go", "serverless-azure-hcl",
	"serverless-azure-python", "serverless-azure-typescript", "serverless-azure-yaml",
	"serverless-cloudflare-csharp", "serverless-cloudflare-go", "serverless-cloudflare-hcl",
	"serverless-cloudflare-python", "serverless-cloudflare-typescript", "serverless-cloudflare-yaml",
	"serverless-gcp-csharp", "serverless-gcp-go", "serverless-gcp-hcl", "serverless-gcp-python",
	"serverless-gcp-typescript", "serverless-gcp-yaml",
	"static-website-aws-bun", "static-website-aws-csharp", "static-website-aws-go",
	"static-website-aws-hcl", "static-website-aws-python", "static-website-aws-typescript",
	"static-website-aws-yaml",
	"static-website-azure-csharp", "static-website-azure-go", "static-website-azure-hcl",
	"static-website-azure-python", "static-website-azure-typescript", "static-website-azure-yaml",
	"static-website-cloudflare-csharp", "static-website-cloudflare-go",
	"static-website-cloudflare-hcl", "static-website-cloudflare-python",
	"static-website-cloudflare-typescript", "static-website-cloudflare-yaml",
	"static-website-gcp-csharp", "static-website-gcp-go", "static-website-gcp-hcl",
	"static-website-gcp-python", "static-website-gcp-typescript", "static-website-gcp-yaml",
	"vm-aws-csharp", "vm-aws-go", "vm-aws-hcl", "vm-aws-python", "vm-aws-typescript", "vm-aws-yaml",
	"vm-azure-csharp", "vm-azure-go", "vm-azure-hcl", "vm-azure-python", "vm-azure-typescript",
	"vm-azure-yaml",
	"vm-gcp-csharp", "vm-gcp-go", "vm-gcp-hcl", "vm-gcp-python", "vm-gcp-typescript", "vm-gcp-yaml",
	"webapp-kubernetes-csharp", "webapp-kubernetes-go", "webapp-kubernetes-hcl",
	"webapp-kubernetes-python", "webapp-kubernetes-typescript", "webapp-kubernetes-yaml",
}

func randomTemplateSourceURL() string {
	dir := templateSourceDirs[rand.Intn(len(templateSourceDirs))] //nolint:gosec // For testing.
	return "https://github.com/pulumi/templates/" + dir
}
