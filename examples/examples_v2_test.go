//go:build v2 || all
// +build v2 all

// V2 integration tests over hand-maintained per-language examples under
// examples/v2/<name>/{yaml,typescript,python,go,csharp,java}/.
//
// We previously authored each example as a single PCL (`main.pp`) and
// invoked `pulumi convert` at test time to derive the language outputs.
// `pulumi convert` mishandles multi-module schemas (our `v2` module under
// the `pulumiservice` package) — string literals are emitted as bare
// identifiers in Python, no pulumi.String wrappers are inserted in Go,
// nested object literals fall through to invalid C# syntax. The single
// fix would be in pulumi/pulumi's PCL→language codegen; until then,
// each language is hand-maintained in its own directory.
//
// TestV2 iterates every (example, language) pair. The yaml lane uses the
// in-process provider directly and runs full up/preview-no-changes/refresh
// /destroy when the v2 surface supports it. Other lanes always run
// preview-only — that's enough to verify the program compiles, resolves
// v2 type tokens against the in-process provider, and constructs valid
// resource registrations.
//
// Some examples are PreviewOnlyAll because their v2 create operations
// are blocked by missing idField support, mutation-style API shapes, or
// non-idempotent endpoints; see the v2Cases table.

package examples

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/assertrefresh"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// v2Case describes how to exercise one example.
type v2Case struct {
	// Name is the directory under examples/v2/<name>/.
	Name string
	// Config returns the config to set on the test stack via SetConfig.
	// Lazy because ServiceProviderTestOrg is mutated by TestMain after
	// package-level initialization.
	Config func() map[string]string
	// UpdateOverrides, when set, are applied between the initial Up and a
	// follow-up Up — exercising the provider's Update operation against
	// the live backend. Each key replaces the same key in Config(); other
	// config carries over. Only one extra Up is performed; if your
	// resource needs a multi-step update flow, write a dedicated test.
	// Only consulted on the yaml lane (where FullE2E runs).
	UpdateOverrides func() map[string]string
	// FullE2E runs up + preview-no-changes + refresh + destroy on the yaml
	// lane only. Other lanes always run preview-only.
	FullE2E bool
	// PreviewOnlyAll forces preview-only on every lane (including yaml).
	// Used when the v2 surface is partial or blocked at create time.
	PreviewOnlyAll bool
	// SkipReason, when non-empty, skips this example on every lane.
	SkipReason string
	// SkipLang skips specific language lanes for this example.
	SkipLang map[string]string
	// Dir overrides the example directory under examples/v2/. Defaults to
	// Name when empty. Useful for running multiple tests against the same
	// example (e.g. one for in-place update, another for replace).
	Dir string
}

var v2Cases = []v2Case{
	{
		Name:    "webhooks",
		FullE2E: true,
		Config:  webhookConfig,
		// Mutate the webhook secret (rotation) — verifies the provider
		// Update path actually patches the existing webhook rather than
		// replacing it.
		UpdateOverrides: func() map[string]string {
			return map[string]string{"secretValue": "rotated-" + generateRandomFiveDigits()}
		},
	},
	{
		Name:    "oidc-issuer",
		Config:  oidcIssuerConfig,
		FullE2E: true,
		// Bump the maxExpiration window — UpdateOidcIssuer should patch
		// the existing issuer registration in place.
		UpdateOverrides: func() map[string]string {
			return map[string]string{"maxExpiration": "7200"}
		},
	},
	{
		Name:    "policy-groups",
		Config:  policyGroupsConfig,
		FullE2E: true, // create=NewPolicyGroup, update=BatchUpdatePolicyGroup
	},
	{
		Name:           "policy-packs",
		Config:         orgOnlyConfig,
		PreviewOnlyAll: true, // Create returns an upload URL; finishing the resource requires uploading a policy zip we don't have here.
	},
	{
		Name:    "environments",
		Config:  environmentsConfig,
		FullE2E: true,
	},
	{
		Name:    "approval-rules",
		Config:  orgOnlyConfig,
		FullE2E: true,
	},
	{
		Name:    "rbac",
		Config:  rbacConfig,
		FullE2E: true, // Role + Team both have clean CRUD now.
		// Rotate the Role description — exercises UpdateRole.
		UpdateOverrides: func() map[string]string {
			return map[string]string{"roleDescription": "Updated description from rbac update test."}
		},
	},
	{
		Name:    "schedules",
		Config:  schedulesConfig,
		FullE2E: true,
		// Reschedule from 7am → 8am cron — UpdateScheduledDeployment.
		UpdateOverrides: func() map[string]string {
			return map[string]string{"scheduleCron": "0 8 * * *"}
		},
	},
	{
		Name:    "stack",
		Config:  stackConfig,
		FullE2E: true,
		// Stack v2 has no top-level Update endpoint — tag mutation goes
		// through separate UpdateStackTags / AddStackTag endpoints we
		// don't yet model. Skip the Update step.
	},
	{
		Name:    "deployment-settings",
		Config:  stackConfig,
		FullE2E: true,
		// Switch executor image — exercises PatchDeploymentSettings.
		UpdateOverrides: func() map[string]string {
			return map[string]string{"executorImage": "pulumi/pulumi:latest"}
		},
	},
	{
		Name:    "teams",
		Config:  teamsConfig,
		FullE2E: true,
		// Rotate the team description — exercises UpdateTeam.
		UpdateOverrides: func() map[string]string {
			return map[string]string{"teamDescription": "Rotated description from update test."}
		},
	},
	{
		// Reuses the teams example. Mutates teamSuffix on the second Up,
		// which changes the team's name (a path-param-derived field marked
		// replaceOnChanges) and forces a replace cycle. Without
		// deleteBeforeReplace=true on Team, the engine would create-new-
		// then-delete-old and the new POST would 409-conflict against the
		// still-existing team. With dbr, the engine destroys the old team
		// first, then creates the renamed one.
		Name:    "teams-replace",
		Dir:     "teams",
		Config:  teamsConfig,
		FullE2E: true,
		UpdateOverrides: func() map[string]string {
			return map[string]string{"teamSuffix": "rep-" + generateRandomFiveDigits()}
		},
	},
	{
		Name:           "access-tokens",
		Config:         accessTokensConfig,
		PreviewOnlyAll: true, // PersonalToken is tied to the caller user; safer not to mutate live state.
	},
	{
		Name:    "agent-pools",
		Config:  agentPoolsConfig,
		FullE2E: true,
		// Rotate the description — exercises PatchOrgAgentPool.
		UpdateOverrides: func() map[string]string {
			return map[string]string{"poolDescription": "Rotated description from update test."}
		},
	},
	{
		Name:           "insights-account",
		Config:         insightsAccountConfig,
		PreviewOnlyAll: true, // Account creation needs real cloud credentials in the referenced ESC env.
	},
	{
		Name:    "stack-tags",
		Config:  stackTagsConfig,
		FullE2E: true, // Stack + StackTag both have clean CRUD.
	},
	{
		Name:    "template-sources",
		Config:  templateSourcesConfig,
		FullE2E: true,
	},
	{
		Name:           "organization-members",
		Config:         organizationMembersConfig,
		PreviewOnlyAll: true, // Adding/removing real users to the test org isn't safe in CI.
	},
	{
		Name:           "audit-log-export",
		Config:         orgOnlyConfig,
		PreviewOnlyAll: true, // Real S3 export needs valid AWS role + bucket.
	},
	{
		Name:           "auth-policy",
		Config:         orgOnlyConfig,
		PreviewOnlyAll: true, // AuthPolicy mutates org-wide auth — preview-only.
	},
	{
		Name:           "default-organization",
		Config:         orgOnlyConfig,
		PreviewOnlyAll: true, // Mutates the caller user's default org.
	},
	{
		Name:           "environment-config",
		Config:         environmentConfigConfig,
		PreviewOnlyAll: true, // Targets an existing ESC environment we don't provision here.
	},
	{
		Name:           "policy-issue",
		Config:         policyIssueConfig,
		PreviewOnlyAll: true, // issueId path-param refers to an existing policy violation.
	},
	{
		Name:           "saml-organization",
		Config:         orgOnlyConfig,
		PreviewOnlyAll: true, // Real SAML config needs an IdP EntityDescriptor.
	},
	{
		Name:           "service",
		Config:         serviceCatalogConfig,
		PreviewOnlyAll: true, // Service catalog items reference stacks we don't own.
	},
	{
		Name:           "stack-config",
		Config:         stackConfigCatalogConfig,
		PreviewOnlyAll: true, // StackConfig links to an ESC env we don't provision here.
	},
	{
		Name:           "task",
		Config:         taskConfig,
		PreviewOnlyAll: true, // taskID path-param refers to an existing agent task definition.
	},
	{
		Name:           "vcs-integrations",
		Config:         vcsIntegrationsConfig,
		PreviewOnlyAll: true, // Integration IDs reference existing org-installed VCS apps.
	},
	{
		Name:           "custom-vcs",
		Config:         customVcsConfig,
		PreviewOnlyAll: true, // Custom VCS registration needs valid creds in the referenced ESC env.
	},
	{
		Name:           "platform-bootstrap",
		Config:         platformBootstrapConfig,
		PreviewOnlyAll: true, // Composite demo (~18 resources). FullE2E is too entangled to run automatically — OrgTemplateCollection has no read op so leftover state from a failed run conflicts on subsequent attempts.
	},
}

var v2Languages = []string{"yaml", "typescript", "python", "go", "csharp", "java"}

func TestV2(t *testing.T) {
	sweepV2TestState(t)
	for _, ex := range v2Cases {
		ex := ex
		for _, lang := range v2Languages {
			lang := lang
			t.Run(ex.Name+"/"+lang, func(t *testing.T) {
				t.Parallel()
				if ex.SkipReason != "" {
					t.Skip(ex.SkipReason)
				}
				if reason := skipLangReason(t, lang); reason != "" {
					t.Skip(reason)
				}
				if reason, ok := ex.SkipLang[lang]; ok {
					t.Skip(reason)
				}
				runV2Case(t, ex, lang)
			})
		}
	}
}

// runV2Case stages the per-language example into a temp dir, applies any
// language-specific local-SDK linkage, and runs preview (or full E2E for
// the yaml lane on FullE2E cases).
func runV2Case(t *testing.T, ex v2Case, lang string) {
	dir := ex.Dir
	if dir == "" {
		dir = ex.Name
	}
	src := filepath.Join(getCwd(t), "v2", dir, lang)
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("%s example missing for %s: %v", lang, ex.Name, err)
	}

	// Stage to a temp dir so per-test patches (go.mod require injection,
	// pulumitest's csproj/go.mod edits, our Pulumi.yaml uv-toolchain rewrite)
	// don't dirty the committed source tree.
	workdir := stageExampleDir(t, src)

	switch lang {
	case "go":
		injectGoSdkRequire(t, workdir, repoRoot(t))
	case "csharp":
		// providertest's DotNetReference runs *after* `pulumi install` (which
		// compiles dotnet). We patch the csproj here, before NewPulumiTest's
		// install step, so the local ProjectReference is in place.
		dropCsharpPackageRef(t, workdir)
		addCsharpProjectRef(t, workdir, repoRoot(t))
		retargetCsprojFramework(t, workdir)
	case "python":
		// Rewrite Pulumi.yaml's `runtime: python` into `runtime.options.toolchain: uv`
		// so pulumi creates a uv-managed venv (uv ignores PEP 668, which
		// otherwise blocks pip's bootstrap on Homebrew Python).
		setPythonToolchainUv(t, workdir)
	}

	opts := []opttest.Option{
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
		opttest.StackName(randomStackName()),
	}
	opts = append(opts, langLinkOptions(t, lang)...)

	test := pulumitest.NewPulumiTest(t, workdir, opts...)
	if lang == "python" {
		// pulumitest's auto-install uses the bundled pip toolchain even when
		// Pulumi.yaml says `toolchain: uv`; replicate via the CLI.
		runPulumiInstall(t, test.WorkingDir())
		// Editable-install the local sdk into the uv-managed venv.
		uvPipInstallEditable(t, test.WorkingDir(), filepath.Join(repoRoot(t), "sdk", "python"))
	}
	if ex.Config != nil {
		for k, v := range ex.Config() {
			test.SetConfig(t, k, v)
		}
	}

	previewOnly := ex.PreviewOnlyAll || lang != "yaml" || !ex.FullE2E
	if previewOnly {
		test.Preview(t)
		return
	}

	// Full E2E flow with optional Update step:
	//   up(initial) → [setConfig(updates) → up → preview-no-changes → refresh] → destroy
	// The Update step verifies the provider's Update operation actually
	// patches the backend; without it, FullE2E only proves Create works.
	test.Up(t)
	if ex.UpdateOverrides != nil {
		for k, v := range ex.UpdateOverrides() {
			test.SetConfig(t, k, v)
		}
		test.Up(t)
	}
	previewResult := test.Preview(t)
	assertpreview.HasNoChanges(t, previewResult)
	refreshResult := test.Refresh(t)
	assertrefresh.HasNoChanges(t, refreshResult)
	test.Destroy(t)
}

// langLinkOptions wires the test's program at the locally built SDK.
// yaml uses the in-process provider directly so needs no linkage.
func langLinkOptions(t *testing.T, lang string) []opttest.Option {
	t.Helper()
	switch lang {
	case "yaml":
		return nil
	case "typescript":
		return []opttest.Option{opttest.YarnLink("@pulumi/pulumiservice")}
	case "python":
		return []opttest.Option{opttest.SkipInstall()}
	case "go":
		return nil // injectGoSdkRequire writes the require+replace before NewPulumiTest
	case "csharp":
		return nil // ProjectReference is written by addCsharpProjectRef before NewPulumiTest
	case "java":
		return []opttest.Option{opttest.Env("PULUMI_REPO_ROOT", repoRoot(t))}
	}
	t.Fatalf("unknown language: %s", lang)
	return nil
}

// skipLangReason returns a non-empty string if the lane should be skipped
// because its prerequisite local SDK isn't built or its toolchain is
// missing. Empty means run.
func skipLangReason(t *testing.T, lang string) string {
	t.Helper()
	root := repoRoot(t)
	switch lang {
	case "yaml":
		return ""
	case "typescript":
		if _, err := os.Stat(filepath.Join(root, "sdk", "nodejs", "bin", "v2")); err != nil {
			return "local @pulumi/pulumiservice (sdk/nodejs/bin) not built; run `make build_nodejs install_nodejs_sdk`"
		}
	case "python":
		if _, err := os.Stat(filepath.Join(root, "sdk", "python", "pulumi_pulumiservice", "v2")); err != nil {
			return "local pulumi_pulumiservice (sdk/python) not built; run `make build_python`"
		}
		if _, err := exec.LookPath("uv"); err != nil {
			return "uv not installed; install it (e.g. `brew install uv`) so pulumi's python toolchain bypasses PEP 668"
		}
	case "go":
		if _, err := os.Stat(filepath.Join(root, "sdk", "go", "pulumiservice", "v2")); err != nil {
			return "local pulumiservice/v2 (sdk/go) not built; run `make build_go`"
		}
	case "csharp":
		if _, err := filepath.Glob(filepath.Join(root, "sdk", "dotnet", "*.csproj")); err != nil {
			return "local Pulumi.PulumiService (sdk/dotnet) not built; run `make build_dotnet`"
		}
	case "java":
		// Local sdk/java is published to ~/.m2 by `make build_java`.
		if _, err := os.Stat(filepath.Join(root, "sdk", "java", "build", "libs")); err != nil {
			return "local pulumiservice java sdk not built; run `make build_java`"
		}
		// We need a JDK matching the toolchain version in the regenerated
		// build.gradle (currently 17). gradle finds it via `org.gradle.java.home`
		// or JAVA_HOME, so just check that some JDK is on PATH.
		if _, err := exec.LookPath("java"); err != nil {
			return "java not on PATH"
		}
	}
	return ""
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join(getCwd(t), ".."))
	if err != nil {
		t.Fatalf("repoRoot: %v", err)
	}
	return root
}

// stageExampleDir copies an example into a per-test temp dir so the
// committed source tree is never modified by per-test SDK linking.
func stageExampleDir(t *testing.T, src string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), filepath.Base(src))
	if err := copyTree(src, dst); err != nil {
		t.Fatalf("stage example dir %s: %v", src, err)
	}
	return dst
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(out, info.Mode())
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		w, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, in)
		return err
	})
}

// injectGoSdkRequire writes both the require and the replace into go.mod
// before pulumitest's first `go mod tidy` runs. The hand-maintained
// go.mod ships without a require for the local pulumiservice sdk so
// users can `go get` the published version against a stable repo path;
// in tests we point it at the workspace.
func injectGoSdkRequire(t *testing.T, workdir, root string) {
	t.Helper()
	sdkAbs, err := filepath.Abs(filepath.Join(root, "sdk"))
	if err != nil {
		t.Fatalf("abs sdk: %v", err)
	}
	cmd := exec.Command("go", "mod", "edit",
		"-require=github.com/pulumi/pulumi-pulumiservice/sdk@v0.0.0",
		"-replace=github.com/pulumi/pulumi-pulumiservice/sdk="+sdkAbs,
	)
	cmd.Dir = workdir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("inject go sdk require/replace: %v\n%s", err, string(out))
	}
}

// dropCsharpPackageRef removes the rendered <PackageReference Include="Pulumi.PulumiService" .../>
// line if present so the local <ProjectReference> doesn't conflict.
func dropCsharpPackageRef(t *testing.T, workdir string) {
	t.Helper()
	csprojPath := singleCsproj(t, workdir)
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		t.Fatalf("read csproj: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.Contains(line, `Include="Pulumi.PulumiService"`) {
			continue
		}
		kept = append(kept, line)
	}
	if err := os.WriteFile(csprojPath, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatalf("write csproj: %v", err)
	}
}

// addCsharpProjectRef appends a ProjectReference ItemGroup pointing at
// the locally built dotnet SDK. Done inline rather than through
// opttest.DotNetReference because providertest applies DotNetReferences
// after `pulumi install`, which is too late.
func addCsharpProjectRef(t *testing.T, workdir, root string) {
	t.Helper()
	csprojPath := singleCsproj(t, workdir)
	localRefs, _ := filepath.Glob(filepath.Join(root, "sdk", "dotnet", "*.csproj"))
	if len(localRefs) == 0 {
		t.Fatalf("no .csproj under sdk/dotnet (run `make build_dotnet`)")
	}
	localCsproj, err := filepath.Abs(localRefs[0])
	if err != nil {
		t.Fatalf("abs sdk/dotnet csproj: %v", err)
	}
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		t.Fatalf("read csproj: %v", err)
	}
	injection := "  <ItemGroup>\n    <ProjectReference Include=\"" + localCsproj + "\" />\n  </ItemGroup>\n"
	out := strings.Replace(string(data), "</Project>", injection+"</Project>", 1)
	if err := os.WriteFile(csprojPath, []byte(out), 0o644); err != nil {
		t.Fatalf("write csproj: %v", err)
	}
}

// retargetCsprojFramework rewrites <TargetFramework>net8.0</TargetFramework>
// to whatever the host has installed.
func retargetCsprojFramework(t *testing.T, workdir string) {
	t.Helper()
	csprojPath := singleCsproj(t, workdir)
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		t.Fatalf("read csproj: %v", err)
	}
	tfm := detectDotnetTFM(t)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.Contains(line, "<TargetFramework>") {
			lines[i] = "    <TargetFramework>" + tfm + "</TargetFramework>"
		}
	}
	if err := os.WriteFile(csprojPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatalf("write csproj: %v", err)
	}
}

func singleCsproj(t *testing.T, workdir string) string {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(workdir, "*.csproj"))
	if len(matches) != 1 {
		t.Fatalf("expected exactly one .csproj in %s, got %d", workdir, len(matches))
	}
	return matches[0]
}

// detectDotnetTFM returns a TargetFrameworkMoniker matching the host's
// installed Microsoft.NETCore.App. Falls back to net8.0 if detection
// fails.
func detectDotnetTFM(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("dotnet", "--list-runtimes").CombinedOutput()
	if err != nil {
		t.Logf("dotnet --list-runtimes failed (%v); defaulting to net8.0", err)
		return "net8.0"
	}
	best := ""
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "Microsoft.NETCore.App ") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		major := strings.SplitN(parts[1], ".", 2)[0]
		if best == "" || major > best {
			best = major
		}
	}
	if best == "" {
		return "net8.0"
	}
	return "net" + best + ".0"
}

// runPulumiInstall invokes the pulumi CLI's `install` command in dir.
// Used for the python lane where pulumitest's auto-install path uses the
// bundled pip toolchain rather than honoring runtime.options.toolchain: uv.
func runPulumiInstall(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("pulumi", "install")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pulumi install in %s: %v\n%s", dir, err, string(out))
	}
}

// uvPipInstallEditable runs `uv pip install -e <pkg>` against the
// project's uv-managed venv. Replaces opttest.PythonLink which shells
// out to system python (PEP 668-blocked on Homebrew).
func uvPipInstallEditable(t *testing.T, projectDir, pkgPath string) {
	t.Helper()
	cmd := exec.Command("uv", "pip", "install", "-e", pkgPath)
	cmd.Dir = projectDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("uv pip install -e %s in %s: %v\n%s", pkgPath, projectDir, err, string(out))
	}
}

// setPythonToolchainUv rewrites Pulumi.yaml's `runtime: python` to the
// long form selecting uv:
//
//	runtime:
//	  name: python
//	  options:
//	    toolchain: uv
func setPythonToolchainUv(t *testing.T, workdir string) {
	t.Helper()
	path := filepath.Join(workdir, "Pulumi.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Pulumi.yaml: %v", err)
	}
	out := strings.Replace(string(data), "runtime: python\n", "runtime:\n  name: python\n  options:\n    toolchain: uv\n", 1)
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		t.Fatalf("write Pulumi.yaml: %v", err)
	}
}

// --- Per-example config helpers ------------------------------------------

func orgOnlyConfig() map[string]string {
	return map[string]string{
		"serviceOrg": ServiceProviderTestOrg,
	}
}

func webhookConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":  ServiceProviderTestOrg,
		"secretValue": "shhh-" + suffix,
		"hookSuffix":  suffix,
	}
}

func policyGroupsConfig() map[string]string {
	return map[string]string{
		"serviceOrg": ServiceProviderTestOrg,
		"groupName":  "v2-pg-" + generateRandomFiveDigits(),
	}
}

func environmentsConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":  ServiceProviderTestOrg,
		"projectName": "v2-envs-" + suffix,
		"envSuffix":   suffix,
	}
}

func rbacConfig() map[string]string {
	return map[string]string{
		"serviceOrg":      ServiceProviderTestOrg,
		"nameSuffix":      generateRandomFiveDigits(),
		"roleDescription": "Read-only access to stacks, created by the v2 rbac example.",
	}
}

func schedulesConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":   ServiceProviderTestOrg,
		"projectName":  "v2-schedules-" + suffix,
		"stackName":    "dev-" + suffix,
		"envName":      "v2-sched-env-" + suffix,
		"scheduleCron": "0 7 * * *",
	}
}

func stackConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":     ServiceProviderTestOrg,
		"projectName":    "v2-stack-" + suffix,
		"stackName":      "dev-" + suffix,
		"stackPurpose":   "demo",
		"executorImage":  "pulumi-cli",
	}
}

func teamsConfig() map[string]string {
	return map[string]string{
		"serviceOrg":      ServiceProviderTestOrg,
		"teamSuffix":      generateRandomFiveDigits(),
		"teamDescription": "A team created by the v2 example.",
	}
}

func oidcIssuerConfig() map[string]string {
	// RegisterOidcIssuer 409s if the issuer URL is already registered.
	// The pre-test sweeper deletes any prior pulumi_issuer/github_issuer
	// before the run, so the URLs are reusable. issuerSuffix kept for
	// per-test name uniqueness.
	return map[string]string{
		"serviceOrg":    ServiceProviderTestOrg,
		"issuerSuffix":  generateRandomFiveDigits(),
		"maxExpiration": "3600",
	}
}

func accessTokensConfig() map[string]string {
	return map[string]string{
		"serviceOrg":       ServiceProviderTestOrg,
		"tokenSuffix":      generateRandomFiveDigits(),
		"tokenDescription": "v2 access-tokens example",
	}
}

func agentPoolsConfig() map[string]string {
	return map[string]string{
		"serviceOrg":      ServiceProviderTestOrg,
		"poolSuffix":      generateRandomFiveDigits(),
		"poolDescription": "v2 example agent pool",
	}
}

func insightsAccountConfig() map[string]string {
	return map[string]string{
		"serviceOrg":          ServiceProviderTestOrg,
		"accountSuffix":       generateRandomFiveDigits(),
		"insightsEnvironment": "insights/credentials",
	}
}

func stackTagsConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":  ServiceProviderTestOrg,
		"projectName": "v2-stack-tags-" + suffix,
		"stackName":   "dev-" + suffix,
		"tagValue":    "v2-tag-" + suffix,
	}
}

func templateSourcesConfig() map[string]string {
	return map[string]string{
		"serviceOrg":     ServiceProviderTestOrg,
		"templateSuffix": generateRandomFiveDigits(),
		"sourceUrl":      "https://github.com/pulumi/examples",
	}
}

func organizationMembersConfig() map[string]string {
	return map[string]string{
		"serviceOrg":   ServiceProviderTestOrg,
		"memberLogin":  "pulumi-bot",
		"memberRole":   "member",
	}
}

func environmentConfigConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":  ServiceProviderTestOrg,
		"projectName": "v2-envcfg-" + suffix,
		"envName":     "v2-envcfg-env-" + suffix,
	}
}

func policyIssueConfig() map[string]string {
	return map[string]string{
		"serviceOrg": ServiceProviderTestOrg,
		"issueId":    "v2-issue-" + generateRandomFiveDigits(),
	}
}

func serviceCatalogConfig() map[string]string {
	return map[string]string{
		"serviceOrg":    ServiceProviderTestOrg,
		"serviceSuffix": generateRandomFiveDigits(),
	}
}

func stackConfigCatalogConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":  ServiceProviderTestOrg,
		"projectName": "v2-stack-config-" + suffix,
		"stackName":   "dev-" + suffix,
		"hookUrl":     "https://example.invalid/hooks/" + suffix,
		"envRef":      "organization/credentials",
	}
}

func taskConfig() map[string]string {
	return map[string]string{
		"serviceOrg": ServiceProviderTestOrg,
		"taskSuffix": generateRandomFiveDigits(),
		"taskID":     "v2-task-" + generateRandomFiveDigits(),
	}
}

func vcsIntegrationsConfig() map[string]string {
	suffix := generateRandomFiveDigits()
	return map[string]string{
		"serviceOrg":                    ServiceProviderTestOrg,
		"githubIntegrationId":           "gh-" + suffix,
		"githubEnterpriseIntegrationId": "ghe-" + suffix,
		"gitlabIntegrationId":           "gl-" + suffix,
		"bitbucketIntegrationId":        "bb-" + suffix,
		"azureDevOpsIntegrationId":      "ado-" + suffix,
	}
}

func customVcsConfig() map[string]string {
	return map[string]string{
		"serviceOrg": ServiceProviderTestOrg,
		"vcsSuffix":  generateRandomFiveDigits(),
		"baseUrl":    "https://git.example.invalid",
		"envRef":     "organization/vcs-credentials",
	}
}

func platformBootstrapConfig() map[string]string {
	return map[string]string{
		"serviceOrg":          ServiceProviderTestOrg,
		"suffix":              generateRandomFiveDigits(),
		"prodApprovalEnabled": "true",
		"slackWebhookUrl":     "https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap",
		"pagerDutyWebhookUrl": "https://events.pagerduty.com/v2/enqueue",
	}
}
