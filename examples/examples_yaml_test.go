//go:build yaml || all
// +build yaml all

package examples

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/apitype"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type Resource struct {
	Type       string                 `yaml:"type"`
	Properties map[string]interface{} `yaml:"properties"`
	Options    map[string]interface{} `yaml:"options"`
}
type YamlProgram struct {
	Name        string              `yaml:"name"`
	Runtime     string              `yaml:"runtime"`
	Description string              `yaml:"description"`
	Resources   map[string]Resource `yaml:"resources"`
}

// mustParseDescriptor builds a typed apitype.PermissionDescriptor from a
// wire-shape JSON literal — used by RBAC integration fixtures that author
// permission descriptors via direct REST.
func mustParseDescriptor(t *testing.T, wireJSON string) apitype.PermissionDescriptor {
	t.Helper()
	var d apitype.PermissionDescriptor
	require.NoError(t, apitype.UnmarshalJSONPermissionDescriptor([]byte(wireJSON), &d))
	require.NotNil(t, d, "wire JSON must parse to a known descriptor variant")
	return d
}

func TestYamlTeamsExample(t *testing.T) {

	// This test builds a repro of https://github.com/pulumi/pulumi-pulumiservice/issues/73.
	// The scenario is where a team is partially created due to a invalid member being specified.
	// This leaves a team resource around even though the creation failed.
	// A subsequent update to the stack with the invalid team member removed
	// should run Update against the partially created team, instead of attempting
	// to Create the team
	t.Run("Properly Updates After Partial Creation Failure", func(t *testing.T) {
		teamName := uuid.NewString()
		projectName := "yaml-teams-fail-" + uuid.NewString()[0:10]

		// this function writes out a Pulumi.yaml file to a temp directory
		writeTeamProgram := func(members []string) string {
			prog := YamlProgram{
				Name:    projectName,
				Runtime: "yaml",
				Resources: map[string]Resource{
					"team": {
						Type: "pulumiservice:index:Team",
						Properties: map[string]interface{}{
							"name":             teamName,
							"teamType":         "pulumi",
							"displayName":      teamName,
							"organizationName": ServiceProviderTestOrg,
							"members":          members,
						},
					},
				},
			}

			return writePulumiYaml(t, prog)
		}

		pulumiBot := "pulumi-bot"
		providerUser := "service-provider-example-user"

		notFoundDir := writeTeamProgram([]string{pulumiBot, "not-found-user"})
		correctUpdateDir := writeTeamProgram([]string{pulumiBot, providerUser})

		first := &strings.Builder{}
		firstOut := io.MultiWriter(os.Stdout, first)

		second := &strings.Builder{}
		secondOut := io.MultiWriter(os.Stdout, second)

		integration.ProgramTest(t, &integration.ProgramTestOptions{
			Quick:         true,
			ExpectFailure: true,
			SkipRefresh:   true,
			Dir:           notFoundDir,
			Stdout:        firstOut,
			Stderr:        firstOut,
			EditDirs: []integration.EditDir{
				{
					Dir: correctUpdateDir,
					// Additive specifies that we're copying our directory on top of the previous one.
					// This overwrites the previous Pulumi.yaml.
					Additive: true,
					Verbose:  true,
					Stdout:   secondOut,
					Stderr:   secondOut,
					// explicitly do not expect a failure
					ExpectFailure: false,
				},
			},
		})

		// ensure first run's output printed an error that user was not found
		assert.Contains(t, first.String(), "User 'not-found-user' not found")

		// ensure second run's output showed that members field was updated
		assert.Contains(t, second.String(), "[diff: ~members]")
	})

	t.Run("Yaml Teams Example", func(t *testing.T) {
		cwd, _ := os.Getwd()
		integration.ProgramTest(t, &integration.ProgramTestOptions{
			Quick: true,
			Dir:   path.Join(cwd, ".", "yaml-teams"),
		})
	})

}

func TestYamlStackTagsExample(t *testing.T) {

	// yaml-stack-tags-example applies tags to it's own stack. To do this, we need to
	// first create an empty stack, then add the stack tag.

	tmpdir := writePulumiYaml(t, YamlProgram{
		Name:        "yaml-stack-tags-example",
		Runtime:     "yaml",
		Description: "A minimal Pulumi YAML program",
	})

	cwd, err := os.Getwd()
	require.NoError(t, err)

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   tmpdir,
		EditDirs: []integration.EditDir{
			{
				Dir: path.Join(cwd, ".", "yaml-stack-tags"),
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             path.Join(cwd, ".", "yaml-stack-tags"),
				ExpectNoChanges: true,
			},
		},
	})
}

func TestYamlStackTagsPluralExample(t *testing.T) {

	// yaml-stack-tags-plural-example applies multiple tags to it's own stack using the StackTags resource.
	// To do this, we need to first create an empty stack, then add the tags.

	tmpdir := writePulumiYaml(t, YamlProgram{
		Name:        "yaml-stack-tags-plural-example",
		Runtime:     "yaml",
		Description: "Example using StackTags resource to manage multiple tags at once",
	})

	cwd, err := os.Getwd()
	require.NoError(t, err)

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   tmpdir,
		Config: map[string]string{
			"organization": getOrgName(),
		},
		EditDirs: []integration.EditDir{
			{
				Dir: path.Join(cwd, ".", "yaml-stack-tags-plural"),
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             path.Join(cwd, ".", "yaml-stack-tags-plural"),
				ExpectNoChanges: true,
			},
		},
	})
}

func TestYamlDeploymentSettingsExample(t *testing.T) {

	// Set up tmpdir with a Pulumi.yml with no resources
	// mimicking the deletion of resource
	newProgram := YamlProgram{
		Name:        "yaml-deployment-settings-example",
		Runtime:     "yaml",
		Description: "Deployment settings test",
	}

	tmpdir := writePulumiYaml(t, newProgram)

	cwd, _ := os.Getwd()
	digits := generateRandomFiveDigits()

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:     true,
		Dir:       path.Join(cwd, ".", "yaml-deployment-settings"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		EditDirs: []integration.EditDir{
			{
				Dir: tmpdir,
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             tmpdir,
				ExpectNoChanges: true,
			},
		},
	})
}

func TestYamlDeploymentSettingsNoSourceExample(t *testing.T) {

	// Set up tmpdir with a Pulumi.yml with no resources
	// mimicking the deletion of resource
	newProgram := YamlProgram{
		Name:    "yaml-deployment-settings-example-no-source",
		Runtime: "yaml",
	}

	tmpdir := writePulumiYaml(t, newProgram)

	cwd, _ := os.Getwd()
	digits := generateRandomFiveDigits()

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:     true,
		Dir:       path.Join(cwd, ".", "yaml-deployment-settings-no-source"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		EditDirs: []integration.EditDir{
			{
				Dir: tmpdir,
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             tmpdir,
				ExpectNoChanges: true,
			},
		},
	})
}

func TestYamlDeploymentSettingsCommitExample(t *testing.T) {

	// Set up tmpdir with a Pulumi.yml with no resources
	// mimicking the deletion of resource
	newProgram := YamlProgram{
		Name:    "yaml-deployment-settings-commit-example",
		Runtime: "yaml",
	}

	tmpdir := writePulumiYaml(t, newProgram)

	cwd, _ := os.Getwd()
	digits := generateRandomFiveDigits()

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:     true,
		Dir:       path.Join(cwd, ".", "yaml-deployment-settings-commit"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		EditDirs: []integration.EditDir{
			{
				Dir: tmpdir,
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             tmpdir,
				ExpectNoChanges: true,
			},
		},
	})
}

func TestYamlTeamAccessTokenExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-team-token"),
	})
}

func TestYamlOrgAccessTokenExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-org-token"),
	})
}

func TestYamlTeamStackPermissionsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-team-stack-permissions"),
	})
}

func TestYamlWebhookExample(t *testing.T) {
	cwd := getCwd(t)
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-webhooks"),
	})
}

func TestYamlSchedulesExample(t *testing.T) {

	t.Run("Yaml Schedules Example", func(t *testing.T) {
		cwd := getCwd(t)
		digits := generateRandomFiveDigits()
		integration.ProgramTest(t, &integration.ProgramTestOptions{
			Dir:       path.Join(cwd, ".", "yaml-schedules"),
			StackName: "test-stack-" + digits,
			Config: map[string]string{
				"digits": digits,
			},
		})
	})

	t.Run("Schedules are replaced on timestamp update", func(t *testing.T) {
		writeScheduleProgram := func(timestamp time.Time) string {
			return writePulumiYaml(t, YamlProgram{
				Name:    "yaml-schedule-reschedule",
				Runtime: "yaml",
				Resources: map[string]Resource{
					"settings": {
						// deployment settings are required to be setup before schedules
						Type: "pulumiservice:DeploymentSettings",
						Properties: map[string]any{
							"organization": ServiceProviderTestOrg,
							"project":      "${pulumi.project}",
							"stack":        "${pulumi.stack}",
						},
					},
					"deployment-schedule": {
						Type: "pulumiservice:DeploymentSchedule",
						Properties: map[string]any{
							"organization":    ServiceProviderTestOrg,
							"project":         "${pulumi.project}",
							"stack":           "${pulumi.stack}",
							"timestamp":       timestamp.Format(time.RFC3339),
							"pulumiOperation": "refresh",
						},
						Options: map[string]any{
							"dependsOn": []string{"${settings}"},
						},
					},
					"ttl-schedule": {
						Type: "pulumiservice:TtlSchedule",
						Properties: map[string]any{
							"organization":       ServiceProviderTestOrg,
							"project":            "${pulumi.project}",
							"stack":              "${pulumi.stack}",
							"timestamp":          timestamp.Format(time.RFC3339),
							"deleteAfterDestroy": false,
						},
						Options: map[string]any{
							"dependsOn": []string{"${settings}"},
						},
					},
				},
			})
		}

		// create some initial one-time schedules
		initialDir := writeScheduleProgram(time.Now().Add(1 * time.Hour))
		// and then reschedule them
		rescheduleDir := writeScheduleProgram(time.Now().Add(2 * time.Hour))

		update := &strings.Builder{}
		updateOut := io.MultiWriter(os.Stdout, update)

		integration.ProgramTest(t, &integration.ProgramTestOptions{
			StackName:   "test-stack-" + generateRandomFiveDigits(),
			Dir:         initialDir,
			Quick:       true,
			SkipRefresh: true,
			EditDirs: []integration.EditDir{
				{
					Dir:             rescheduleDir,
					Additive:        true,
					Stdout:          updateOut,
					Stderr:          updateOut,
					Verbose:         true,
					ExpectNoChanges: false,
				},
			},
		})

		// expect the update to cause schedule replacements
		assert.Contains(t, update.String(), "deployment-schedule replaced")
		assert.Contains(t, update.String(), "ttl-schedule replaced")
	})
}

func TestYamlEnvironmentsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-environments"),
		Config: map[string]string{
			"digits": digits,
		},
	})
}

func TestYamlAgentPoolsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-agent-pools"),
		Config: map[string]string{
			"digits": digits,
		},
	})
}

func TestYamlTemplateSourcesExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-template-sources"),
		Config: map[string]string{
			"digits": digits,
		},
	})
}

func TestYamlOidcIssuerExample(t *testing.T) {
	cwd := getCwd(t)
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-oidc-issuer"),
	})
}

func TestYamlPolicyGroupsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-policy-groups"),
		Config: map[string]string{
			"digits":           digits,
			"organizationName": os.Getenv("PULUMI_TEST_OWNER"),
		},
	})
}

func TestYamlPolicyGroupsAccountsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-policy-groups-accounts"),
		Config: map[string]string{
			"digits":           digits,
			"organizationName": getOrgName(),
			"roleArn":          getInsightsRoleArn(t),
		},
	})
}

func TestYamlApprovalRuleExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-approval-rules"),
		Config: map[string]string{
			"organizationName": getOrgName(),
			"digits":           digits,
		},
	})
}

func TestYamlStackExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-stack"),
		Config: map[string]string{
			"organizationName": getOrgName(),
			"digits":           digits,
		},
	})
}

func TestYamlInsightsAccountExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "yaml-insights-account"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	test.SetConfig(t, "organizationName", getOrgName())
	test.SetConfig(t, "roleArn", getInsightsRoleArn(t))
	runPulumiTest(t, test)
}

func TestYamlDeploymentSettingsVcsExample(t *testing.T) {
	t.Skip("requires an existing Azure DevOps integration; run manually against a configured environment")
}

func TestYamlRbacExample(t *testing.T) {
	// Requires the Custom Roles feature to be enabled on the test
	// organization. If it isn't, CreateRole will return a feature-flag
	// error and the test fails loudly — which is what we want to learn.
	orgName := getOrgName()
	const fixtureUser = "service-provider-example-user"

	// Snapshot the fixture user's role before the test mutates it, and
	// restore it on cleanup. OrganizationRole.Delete with force=true
	// *should* revoke their custom role assignment on destroy, but if a
	// mis-wired teardown leaves the user on the test's role, this hook
	// still gets the org back to a known state.
	t.Cleanup(snapshotFixtureOrgMember(t, orgName, fixtureUser))

	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "yaml-rbac"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.StackName(randomStackName()),
	)
	test.SetConfig(t, "digits", generateRandomFiveDigits())
	test.SetConfig(t, "organizationName", orgName)

	up := test.Up(t)
	// Sanity-check adoption: OrganizationMember should have hit the 409
	// branch and adopted the pre-existing membership.
	if adopted, ok := up.Outputs["memberAdopted"]; ok {
		assert.Equal(t, true, adopted.Value, "expected OrganizationMember to adopt the existing membership")
	}
	// Sanity-check env-scoped role wiring: the Environment resource must
	// surface a non-empty UUID, and the role pinned to it must reach the
	// service. An empty UUID would mean the metadata fetch silently
	// skipped and the role's `identity` was never resolved.
	if envID, ok := up.Outputs["scopedEnvironmentId"]; ok {
		assert.NotEmpty(t, envID.Value, "expected Environment.environmentId to be populated")
	}
	if scopedRoleID, ok := up.Outputs["scopedRoleId"]; ok {
		assert.NotEmpty(t, scopedRoleID.Value, "expected env-scoped OrganizationRole to be created")
	}
	preview := test.Preview(t)
	assertpreview.HasNoChanges(t, preview)
	refresh := test.Refresh(t)
	assertrefresh.HasNoChanges(t, refresh)
	test.Destroy(t)
}

// TestYamlRbacComposeImport pins the provider's Read path against a
// `PermissionDescriptorCompose` role authored out-of-band: `pulumi
// import` must run end-to-end without crashing the provider, and the
// imported state must reference the upstream policy id. The provider-
// level descriptor round-trip is unit-pinned by TestImportRepro_Compose
// in resources/.
//
// Cloud requires a `uxPurpose:"role"` entry to reference
// `uxPurpose:"policy"` entries via Compose's `permissionDescriptors`;
// role-references-role is a 400. PSP doesn't expose policies as a
// managed resource type today, so the test provisions both the policy
// fixture and the Compose role directly via `pulumiapi.Client`
// (bypassing PSP's resource layer).
//
// Known gap (not asserted here): pulumi/pulumi's import-codegen path
// (`pkg/importer/hcl2.go`'s `generateValue`) drops `__`-prefixed map
// keys as "internal properties" before emitting source code. The
// generated YAML therefore carries `permissions: {permissionDescriptors:
// [<id>]}` without the `__type: PermissionDescriptorCompose` key the
// provider's Check requires on a subsequent `pulumi up`. Users
// importing a Compose role today must hand-add the `__type` line to
// their generated program. The fix is upstream and orthogonal to the
// provider; until it lands, the drift-check (Preview HasNoChanges) is
// disabled here.
//
// Flow:
//
//  1. Build a `pulumiapi.Client` from PULUMI_ACCESS_TOKEN /
//     PULUMI_BACKEND_URL. Skip the test if either is unset (local dev
//     without creds).
//  2. Create a policy fixture (uxPurpose=policy) carrying a trivial
//     PermissionDescriptorAllow for `stack:read`.
//  3. Create the Compose role (uxPurpose=role) whose `details` is
//     `{"__type":"PermissionDescriptorCompose","permissionDescriptors":[<policyId>]}`.
//  4. Run `pulumi import pulumiservice:index:OrganizationRole importedComposedRole
//     <org>/<roleId> --out <tmpdir>/imported.yaml` against an empty
//     target stack.
//  5. Assert: import succeeded (return code 0; previously failed with
//     "unknown __type PermissionDescriptorCompose" before the provider
//     learned to pass the descriptor through), and the generated YAML
//     references the policy id (so the Compose body survived even
//     without the `__type` discriminator).
//
// Cleanup, in registration order (LIFO):
//
//   - `pulumitest.NewPulumiTest(importTarget)` registers a destroy hook
//     that runs first. Its destroy attempts to delete the role (now in
//     state from `import`); the role is gone after this.
//   - `t.Cleanup(deleteRole)` runs next; the role is already gone, so
//     `client.DeleteRole`'s 404-swallow path returns nil.
//   - `t.Cleanup(deletePolicy)` runs last; nothing references the policy
//     by then, so the delete succeeds.
func TestYamlRbacComposeImport(t *testing.T) {
	token := os.Getenv("PULUMI_ACCESS_TOKEN")
	apiURL := os.Getenv("PULUMI_BACKEND_URL")
	if token == "" || apiURL == "" {
		t.Skip("requires PULUMI_ACCESS_TOKEN and PULUMI_BACKEND_URL to provision the policy + role fixtures via direct REST")
	}

	orgName := getOrgName()
	digits := generateRandomFiveDigits()
	ctx := context.Background()

	httpClient := &http.Client{Timeout: 60 * time.Second}
	client, err := pulumiapi.NewClient(httpClient, token, apiURL)
	require.NoError(t, err, "must be able to construct a pulumiapi client")

	// Step 1: policy fixture. Carries a trivial Allow descriptor so
	// CreateRole's `details must not be empty` validation passes.
	policyName := fmt.Sprintf("yaml-rbac-import-policy-%s", digits)
	policyDetails := mustParseDescriptor(t,
		`{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`)
	policy, err := client.CreateRole(ctx, orgName, apitype.PermissionDescriptorBase{
		Name:         policyName,
		Description:  "Compose-import test policy fixture",
		ResourceType: "global",
		UxPurpose:    apitype.PermissionDescriptorUXPurposePolicy,
		Details:      policyDetails,
	})
	require.NoError(t, err, "must be able to create a uxPurpose=policy fixture via direct REST")
	require.NotEmpty(t, policy.ID)
	t.Cleanup(func() {
		if err := client.DeleteRole(ctx, orgName, policy.ID, true); err != nil {
			t.Logf("cleanup: failed to delete policy %q: %v", policy.ID, err)
		}
	})

	// Step 2: Compose role referencing the policy. Authoring this via
	// direct REST (rather than a PSP resource) simulates a UI-authored or
	// out-of-band role — what `pulumi import` actually faces in the wild.
	roleName := fmt.Sprintf("yaml-rbac-import-role-%s", digits)
	composeDetails := mustParseDescriptor(t, fmt.Sprintf(
		`{"__type":"PermissionDescriptorCompose","permissionDescriptors":[%q]}`,
		policy.ID,
	))
	role, err := client.CreateRole(ctx, orgName, apitype.PermissionDescriptorBase{
		Name:         roleName,
		Description:  "Compose-import test role; references policy via Compose",
		ResourceType: "global",
		UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
		Details:      composeDetails,
	})
	require.NoError(t, err, "must be able to create a Compose role referencing the policy")
	require.NotEmpty(t, role.ID)
	t.Cleanup(func() {
		if err := client.DeleteRole(ctx, orgName, role.ID, true); err != nil {
			t.Logf("cleanup: failed to delete role %q: %v", role.ID, err)
		}
	})

	// Step 3: empty target stack. Registered AFTER the role so its
	// destroy runs first (LIFO), removing the role from Cloud before our
	// own DeleteRole cleanup hook fires.
	importTarget := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "yaml-rbac-import-target"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.StackName(randomStackName()),
	)

	// Step 4: import the Compose role into the empty target stack with
	// `--out` so we can inspect the generated YAML directly.
	outFile := filepath.Join(importTarget.CurrentStack().Workspace().WorkDir(), "imported.yaml")
	importResult := importTarget.Import(t,
		"pulumiservice:index:OrganizationRole",
		"importedComposedRole",
		fmt.Sprintf("%s/%s", orgName, role.ID),
		"",
		"--out", outFile,
	)
	require.Zero(t, importResult.ReturnCode,
		"pulumi import must succeed.\nstdout:\n%s\nstderr:\n%s",
		importResult.Stdout, importResult.Stderr)

	// Step 5: assert the generated YAML references the policy id —
	// proves the Compose body's `permissionDescriptors` array survived
	// import. The `__type: PermissionDescriptorCompose` line itself is
	// dropped by upstream's import-codegen __-prefix filter (see header
	// comment); deliberately not asserted here.
	contents, err := os.ReadFile(outFile)
	require.NoError(t, err, "must be able to read the import --out file")
	imported := string(contents)

	assert.Contains(t, imported, policy.ID,
		"imported program must reference the policy id inside permissionDescriptors")
}

func writePulumiYaml(t *testing.T, yamlContents interface{}) string {
	tmpdir := t.TempDir()
	b, err := yaml.Marshal(yamlContents)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(tmpdir, "Pulumi.yaml"), b, 0600)
	if err != nil {
		t.Fatal(err)
	}
	return tmpdir
}
