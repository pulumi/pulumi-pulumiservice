// Copyright 2016-2026, Pulumi Corporation.
//
// examples_canonical_test.go — integration tests for the 12 end-to-end
// canonical YAML examples under examples/canonical/. Each example is a
// realistic user story: organization bootstrap, GitHub OIDC CI, ESC
// rotation hooks, etc.
//
// These tests do an up/preview/refresh/destroy cycle against a real
// Pulumi Cloud org (via PULUMI_ACCESS_TOKEN + PULUMI_TEST_OWNER) and are
// the load-bearing proof that the v2 provider works end-to-end.
//
// Build tag `canonical` lets these run independently of the legacy
// per-resource `yaml` suite during v2 bring-up. CI's `all` tag runs both.

//go:build canonical || all
// +build canonical all

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// canonicalTest builds a pulumitest harness for one canonical scenario
// directory. All canonical examples default their organizationName to
// the test org configured via PULUMI_TEST_OWNER (or the hardcoded
// fallback), plus a per-run random digit suffix for unique resource names.
func canonicalTest(t *testing.T, dir string) {
	t.Helper()
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "canonical", dir),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
	)
	test.SetConfig(t, "organizationName", getOrgName())
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	runPulumiTest(t, test)
}

func TestCanonical01OrganizationBootstrap(t *testing.T) {
	canonicalTest(t, "01-organization-bootstrap")
}

func TestCanonical02GithubOidcCi(t *testing.T) {
	canonicalTest(t, "02-github-oidc-ci")
}

func TestCanonical03EscRotationHooks(t *testing.T) {
	canonicalTest(t, "03-esc-rotation-hooks")
}

func TestCanonical04DeploymentPipeline(t *testing.T) {
	canonicalTest(t, "04-deployment-pipeline")
}

func TestCanonical05AuditLogExport(t *testing.T) {
	canonicalTest(t, "05-audit-log-export")
}

func TestCanonical06SelfHostedAgents(t *testing.T) {
	canonicalTest(t, "06-self-hosted-agents")
}

func TestCanonical07TieredTeamAccess(t *testing.T) {
	canonicalTest(t, "07-tiered-team-access")
}

func TestCanonical08ApprovalChangeGates(t *testing.T) {
	canonicalTest(t, "08-approval-change-gates")
}

func TestCanonical09TagBasedRouting(t *testing.T) {
	canonicalTest(t, "09-tag-based-routing")
}

func TestCanonical10TemplateCatalog(t *testing.T) {
	canonicalTest(t, "10-template-catalog")
}

func TestCanonical11CrossCloudOidc(t *testing.T) {
	canonicalTest(t, "11-cross-cloud-oidc")
}

func TestCanonical12SamlTeamSync(t *testing.T) {
	canonicalTest(t, "12-saml-team-sync")
}

// Typed-language variants — one per SDK for each showcase scenario.
// Scenarios 01 (basic CRUD), 04 (PATCH upsert with nested objects), and
// 07 (multiple resources of one kind) are hand-written to exercise the
// distinct SDK idioms. Remaining scenarios stay YAML-primary; consumers
// use `pulumi convert --from yaml --language <lang>` to obtain typed
// programs on demand (each scenario's README documents this).

func TestCanonical01TypeScript(t *testing.T) { canonicalTest(t, "01-organization-bootstrap-ts") }
func TestCanonical01Python(t *testing.T)     { canonicalTest(t, "01-organization-bootstrap-py") }
func TestCanonical01Go(t *testing.T)         { canonicalTest(t, "01-organization-bootstrap-go") }
func TestCanonical01DotNet(t *testing.T)     { canonicalTest(t, "01-organization-bootstrap-cs") }
func TestCanonical01Java(t *testing.T)       { canonicalTest(t, "01-organization-bootstrap-java") }

func TestCanonical04TypeScript(t *testing.T) { canonicalTest(t, "04-deployment-pipeline-ts") }
func TestCanonical04Python(t *testing.T)     { canonicalTest(t, "04-deployment-pipeline-py") }
func TestCanonical04Go(t *testing.T)         { canonicalTest(t, "04-deployment-pipeline-go") }
func TestCanonical04DotNet(t *testing.T)     { canonicalTest(t, "04-deployment-pipeline-cs") }
func TestCanonical04Java(t *testing.T)       { canonicalTest(t, "04-deployment-pipeline-java") }

func TestCanonical07TypeScript(t *testing.T) { canonicalTest(t, "07-tiered-team-access-ts") }
func TestCanonical07Python(t *testing.T)     { canonicalTest(t, "07-tiered-team-access-py") }
func TestCanonical07Go(t *testing.T)         { canonicalTest(t, "07-tiered-team-access-go") }
func TestCanonical07DotNet(t *testing.T)     { canonicalTest(t, "07-tiered-team-access-cs") }
func TestCanonical07Java(t *testing.T)       { canonicalTest(t, "07-tiered-team-access-java") }
