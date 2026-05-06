# Rename v2 `preview` module → `escPreview` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the generic `preview` module name with the domain-scoped `escPreview` for the four `preview/environments`-derived ESC resources, eliminating the awkward `v2.preview.*` doubling and making the relationship to the production `esc` module explicit.

**Architecture:** A one-line change to the scaffolder's `moduleAliases` table flips the canonical module from `preview` to `escPreview`. Re-running `go generate` rewrites the four `*_preview_environments` entries' `token` fields in `metadata.json`; the metadata keys (which are derived from operationIds) stay untouched. We then add Pulumi `aliases` arrays by hand on the four entries so any in-flight stack state pinned to `pulumiservice:v2/preview:*` rebinds cleanly. Embedded `provider/cmd/pulumi-resource-pulumiservice/schema.json` and all five language SDKs regenerate from the new metadata via `make provider` + `make build_sdks`. Old `sdk/*/v2/preview/` directories are removed by hand because codegen leaves obsolete dirs in place.

**Tech Stack:** Go 1.24+ (scaffold-metadata, runtime), Pulumi schema codegen for SDK regeneration, JSON for metadata.

**Name choice:** This plan commits to `escPreview`. To swap in a different name (e.g. `escBeta`, `escUnstable`, `esc2`), do a global find-and-replace on this file before executing — the only literal that matters in code is the value of the map entry in Task 2 and the strings in Task 1's expected tokens.

---

## File map

- **Modify** `provider/tools/scaffold-metadata/main.go` (single map entry on line 126).
- **Auto-regenerate** `provider/pkg/cloud/metadata.json` (`token` field on 4 entries).
- **Hand-edit** `provider/pkg/cloud/metadata.json` (add `aliases` array on those 4 entries).
- **Create** `provider/pkg/cloud/metadata_test.go` (pin new tokens + aliases).
- **Auto-regenerate** `provider/cmd/pulumi-resource-pulumiservice/schema.json` (the embedded schema artifact; rebuilt by `make provider`).
- **Auto-regenerate** `sdk/{nodejs,python,go,dotnet,java}/...` (delete obsolete `v2/preview/` dirs by hand).
- **Verify (probable no-op)** `provider/pkg/provider/provider.go` language map.
- **Skip** `provider/pkg/pulumiapi/insights_accounts.go` (its `"preview"` literals are URL path strings for the v1 InsightsAccount resource, not module names).
- **Skip** `provider/pkg/provider/manual-schema.json:582` and `provider/cmd/pulumi-resource-pulumiservice/schema.json:808` (those `"preview"` literals are enum values for deployment scheduling, unrelated).

---

## Task 1: Pin desired end-state with a unit test

**Files:**
- Create: `provider/pkg/cloud/metadata_test.go`

- [ ] **Step 1: Write the failing test**

```go
// Copyright 2016-2026, Pulumi Corporation.
package cloud_test

import (
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/cloud"
)

// TestPreviewEnvironmentsTokensUseEscPreviewModule pins the user-facing token
// shape for the four preview/environments-derived resources. The metadata
// keys keep their _preview_environments suffix (canonical OpenAPI-derived
// form), but the user-facing `token` field maps them into the escPreview
// module — see scaffold-metadata/main.go moduleAliases.
func TestPreviewEnvironmentsTokensUseEscPreviewModule(t *testing.T) {
	md := cloud.Metadata()

	cases := map[string]string{
		"pulumiservice:v2:Environment_preview_environments":    "pulumiservice:v2/escPreview:Environment",
		"pulumiservice:v2:EnvironmentTag_preview_environments": "pulumiservice:v2/escPreview:EnvironmentTag",
		"pulumiservice:v2:RevisionTag_preview_environments":    "pulumiservice:v2/escPreview:RevisionTag",
		"pulumiservice:v2:Webhook_preview_environments":        "pulumiservice:v2/escPreview:Webhook",
	}
	for key, want := range cases {
		rm, ok := md.Resources[key]
		if !ok {
			t.Errorf("missing metadata entry %q", key)
			continue
		}
		if rm.Token != want {
			t.Errorf("metadata[%q].token = %q, want %q", key, rm.Token, want)
		}
	}
}

// TestPreviewEnvironmentsAliasOldToken ensures stack state under the old
// pulumiservice:v2/preview:* tokens rebinds via Pulumi schema aliases.
func TestPreviewEnvironmentsAliasOldToken(t *testing.T) {
	md := cloud.Metadata()
	cases := map[string]string{
		"pulumiservice:v2:Environment_preview_environments":    "pulumiservice:v2/preview:Environment",
		"pulumiservice:v2:EnvironmentTag_preview_environments": "pulumiservice:v2/preview:EnvironmentTag",
		"pulumiservice:v2:RevisionTag_preview_environments":    "pulumiservice:v2/preview:RevisionTag",
		"pulumiservice:v2:Webhook_preview_environments":        "pulumiservice:v2/preview:Webhook",
	}
	for key, want := range cases {
		rm, ok := md.Resources[key]
		if !ok {
			t.Errorf("missing metadata entry %q", key)
			continue
		}
		var found bool
		for _, a := range rm.Aliases {
			if a == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("metadata[%q].aliases missing %q (got %v)", key, want, rm.Aliases)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd provider/pkg && go test ./cloud/ -run TestPreviewEnvironments -v`

Expected: both tests FAIL.
- `TestPreviewEnvironmentsTokensUseEscPreviewModule`: tokens still resolve to `pulumiservice:v2/preview:*`.
- `TestPreviewEnvironmentsAliasOldToken`: `Aliases` slices are empty.

- [ ] **Step 3: Commit the failing test**

```bash
git add provider/pkg/cloud/metadata_test.go
git commit -m "test: pin escPreview module name + alias for preview environments"
```

---

## Task 2: Update `moduleAliases` in scaffold-metadata

**Files:**
- Modify: `provider/tools/scaffold-metadata/main.go:120-131`

- [ ] **Step 1: Edit the alias table**

In `provider/tools/scaffold-metadata/main.go`, replace this block:

```go
var moduleAliases = map[string]string{
	"agent-pools":          "agents",
	"auth/policies":        "auth",
	"esc/environments":     "esc",
	"oidc/issuers":         "auth",
	"preview/agents":       "agents",
	"preview/environments": "preview",
	"preview/insights":     "insights",
	"saml":                 "auth",
	"stacks/deployments":   "deployments",
	"teams/tokens":         "tokens",
}
```

with:

```go
var moduleAliases = map[string]string{
	"agent-pools":          "agents",
	"auth/policies":        "auth",
	"esc/environments":     "esc",
	"oidc/issuers":         "auth",
	"preview/agents":       "agents",
	"preview/environments": "escPreview",
	"preview/insights":     "insights",
	"saml":                 "auth",
	"stacks/deployments":   "deployments",
	"teams/tokens":         "tokens",
}
```

The change is a single value: `"preview"` → `"escPreview"`.

- [ ] **Step 2: Verify the file compiles**

Run: `cd provider/tools/scaffold-metadata && go build ./...`

Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add provider/tools/scaffold-metadata/main.go
git commit -m "scaffold-metadata: rename preview/environments module to escPreview"
```

---

## Task 3: Regenerate `metadata.json`

**Files:**
- Auto-modify: `provider/pkg/cloud/metadata.json` (4 `token` fields)

- [ ] **Step 1: Run `go generate` for the cloud package**

Run: `cd provider/pkg/cloud && go generate ./...`

Expected stderr (approximately):
```
scaffold-metadata: metadata.json
  operations in spec:        ...
  candidates derived:        49
  added new entries:         0
  updated existing entries:  4
  ...
```

The line `updated existing entries: 4` confirms the alias change took effect on exactly the four affected entries. Anything else (added entries, orphans) means the upstream spec drifted independently — abort and resolve that first.

- [ ] **Step 2: Inspect the diff**

Run: `git diff --stat provider/pkg/cloud/metadata.json && git diff provider/pkg/cloud/metadata.json | grep -E '^[-+].*token|^@@' | head -40`

Expected: only `"token"` lines change. Each affected entry shows:
```
-      "token": "pulumiservice:v2/preview:Environment",
+      "token": "pulumiservice:v2/escPreview:Environment",
```
(and similarly for `EnvironmentTag`, `RevisionTag`, `Webhook`).

If `provider/pkg/cloud/spec.json` ALSO changed in this regen, the upstream OpenAPI spec drifted. Stash the spec.json changes into a separate commit before continuing — this PR should be the rename only.

- [ ] **Step 3: Commit the regen**

```bash
git add provider/pkg/cloud/metadata.json
git commit -m "cloud: regenerate metadata for escPreview module rename"
```

---

## Task 4: Add Pulumi `aliases` to the four affected entries

**Files:**
- Hand-edit: `provider/pkg/cloud/metadata.json` (add `aliases` array on 4 entries)

The scaffolder writes everything except `operations` as write-if-absent, so the manual `aliases` we add here will survive future regenerations.

- [ ] **Step 1: Add an `aliases` array to each of the four entries**

For each of the four entries, add an `aliases` array immediately after the `token` field. The scaffolder's `encodeStable` ordering puts `aliases` after `token`, so this matches what regen would produce.

Apply these four edits in `provider/pkg/cloud/metadata.json`:

(a) `pulumiservice:v2:Environment_preview_environments`:

Find:
```json
      "token": "pulumiservice:v2/escPreview:Environment",
```

Replace with:
```json
      "token": "pulumiservice:v2/escPreview:Environment",
      "aliases": [
        "pulumiservice:v2/preview:Environment"
      ],
```

(b) `pulumiservice:v2:EnvironmentTag_preview_environments`:

Find:
```json
      "token": "pulumiservice:v2/escPreview:EnvironmentTag",
```

Replace with:
```json
      "token": "pulumiservice:v2/escPreview:EnvironmentTag",
      "aliases": [
        "pulumiservice:v2/preview:EnvironmentTag"
      ],
```

(c) `pulumiservice:v2:RevisionTag_preview_environments`:

Find:
```json
      "token": "pulumiservice:v2/escPreview:RevisionTag",
```

Replace with:
```json
      "token": "pulumiservice:v2/escPreview:RevisionTag",
      "aliases": [
        "pulumiservice:v2/preview:RevisionTag"
      ],
```

(d) `pulumiservice:v2:Webhook_preview_environments`:

Find:
```json
      "token": "pulumiservice:v2/escPreview:Webhook",
```

Replace with:
```json
      "token": "pulumiservice:v2/escPreview:Webhook",
      "aliases": [
        "pulumiservice:v2/preview:Webhook"
      ],
```

(Apply each `Find`/`Replace with` independently. Each `token` value is unique, so each `Find` matches exactly once.)

- [ ] **Step 2: Verify the JSON parses cleanly**

Run: `jq '.resources["pulumiservice:v2:Environment_preview_environments"].aliases' provider/pkg/cloud/metadata.json`

Expected output:
```json
[
  "pulumiservice:v2/preview:Environment"
]
```

Run the same probe for the other three keys, substituting the metadata key and expected alias each time.

- [ ] **Step 3: Verify the scaffolder preserves the aliases on regen**

Run: `cd provider/pkg/cloud && go generate ./... && cd ../../.. && git diff --stat provider/pkg/cloud/metadata.json`

Expected stderr from scaffold-metadata: `updated existing entries: 0`. Expected diff: empty.

If the diff is non-empty, the scaffolder dropped or re-ordered the aliases. Investigate `mergeOperations` in `scaffold-metadata/main.go` and the `encodeStable` key order before continuing.

- [ ] **Step 4: Run the unit tests from Task 1**

Run: `cd provider/pkg && go test ./cloud/ -run TestPreviewEnvironments -v`

Expected: both tests PASS.

- [ ] **Step 5: Commit**

```bash
git add provider/pkg/cloud/metadata.json
git commit -m "cloud: alias old preview tokens for escPreview rename"
```

---

## Task 5: Run the full provider test suite

**Files:**
- (No edits expected; if tests reference the old token literal, this task adds a small fix.)

- [ ] **Step 1: Run all provider tests**

Run: `cd provider/pkg && go test -short -count=1 -timeout 30m ./...`

Expected: all PASS. The Pulumi alias makes most tests resilient — a test that exercises the schema dispatch via either the old or the new token should pass under both.

- [ ] **Step 2: Search for any hard-coded references to the old token**

Run:
```bash
cd /Users/lukeward/Workspace/psp/pulumi-pulumiservice
grep -rn 'v2/preview:' provider/pkg provider/cmd examples 2>/dev/null | grep -v 'metadata.json\|spec.json' | grep -v '\.git/'
```

Expected: empty output. If hits appear, they fall into two cases:
- **Test asserts a specific token literal** (e.g. checking BuildSchema returns it): update the literal to `pulumiservice:v2/escPreview:<Type>` and add a sibling test asserting that the alias is present.
- **Documentation or comments**: update the literal to the new token.

- [ ] **Step 3: Commit any test/doc updates**

If any files were touched in Step 2:
```bash
git add provider/ examples/
git commit -m "test: align v2 preview->escPreview token references"
```

(Skip if no changes.)

---

## Task 6: Rebuild the embedded schema and all SDKs

**Files:**
- Auto-modify: `provider/cmd/pulumi-resource-pulumiservice/schema.json`
- Auto-modify: `sdk/nodejs/`, `sdk/python/`, `sdk/go/`, `sdk/dotnet/`, `sdk/java/`

- [ ] **Step 1: Rebuild the provider binary (regenerates the embedded schema.json)**

Run: `make provider`

Expected: `bin/pulumi-resource-pulumiservice` rebuilt; `provider/cmd/pulumi-resource-pulumiservice/schema.json` regenerated.

- [ ] **Step 2: Verify schema.json reflects the new tokens**

Run:
```bash
grep -E 'v2/(preview|escPreview):' provider/cmd/pulumi-resource-pulumiservice/schema.json | sort -u
```

Expected: lines containing `pulumiservice:v2/escPreview:Environment` (and the other three types). No lines containing `pulumiservice:v2/preview:`. The aliases applied in Task 4 surface inside resource bodies under `aliases` keys rather than as top-level resource entries, which is correct.

- [ ] **Step 3: Regenerate all language SDKs**

Run: `make build_sdks`

Expected: each language SDK regenerates. The new module appears at `sdk/nodejs/v2/escPreview/`, `sdk/python/pulumi_pulumiservice/v2/escpreview/`, `sdk/go/pulumiservice/v2/escPreview/`, `sdk/dotnet/V2/EscPreview/` (or similar — see Task 7 for the casing check), `sdk/java/.../v2/escpreview/`.

- [ ] **Step 4: Remove obsolete `v2/preview/` SDK directories**

Pulumi codegen does not delete obsolete module directories. Remove them:

```bash
find sdk -type d -path '*v2/preview' -print
```

Expected: a list under each language. For each path printed, run `git rm -r <path>`. After:

```bash
find sdk -type d -path '*v2/preview' 2>/dev/null
```

Expected: empty.

Also check `sdk/nodejs/tsconfig.json` for any `v2/preview/*.ts` entries (they were in the search results) and confirm the SDK regen replaced them with `v2/escPreview/*.ts`. If `tsconfig.json` still lists `v2/preview/*` paths after `make nodejs_sdk`, the codegen didn't update it — open the file, replace `v2/preview/` with `v2/escPreview/` on the affected lines.

- [ ] **Step 5: Commit the SDK regen**

```bash
git add provider/cmd/pulumi-resource-pulumiservice/schema.json sdk/
git commit -m "sdk: regenerate for escPreview module rename"
```

---

## Task 7: Verify the .NET namespace casing

**Files:**
- Verify (probable no-op): `provider/pkg/provider/provider.go:144-154`

The csharp namespace map currently lists `pulumiservice → PulumiService` and `v2 → V2`. Pulumi's codegen falls back to PascalCase for unmapped path components, so `escPreview` should auto-render as `EscPreview`. We confirm this empirically.

- [ ] **Step 1: Confirm the dotnet SDK got the expected casing**

Run:
```bash
ls sdk/dotnet/V2/EscPreview/ 2>/dev/null && echo OK || echo MISSING
```

If `OK`: nothing to do, skip Steps 2–3.

If `MISSING`: the casing differs. Inspect `sdk/dotnet/V2/` to see what directory the codegen produced (e.g. `escpreview`, `Escpreview`, `EscPREVIEW`).

- [ ] **Step 2: Add an explicit namespace mapping (only if Step 1 reported MISSING)**

Edit `provider/pkg/provider/provider.go`. Find:

```go
			"csharp": map[string]any{
				"namespaces": map[string]any{
					"pulumiservice": "PulumiService",
					"v2":            "V2",
				},
```

Replace with:

```go
			"csharp": map[string]any{
				"namespaces": map[string]any{
					"pulumiservice": "PulumiService",
					"v2":            "V2",
					"escPreview":       "EscPreview",
				},
```

Then rebuild: `make provider && make dotnet_sdk`.

Re-run Step 1 to confirm the directory is now at `sdk/dotnet/V2/EscPreview/`.

- [ ] **Step 3: Commit (only if Step 2 ran)**

```bash
git add provider/pkg/provider/provider.go provider/cmd/pulumi-resource-pulumiservice/schema.json sdk/dotnet/
git commit -m "provider: map escPreview->EscPreview in csharp namespaces"
```

---

## Task 8: Sanity sweep

**Files:**
- (No edits.)

- [ ] **Step 1: Run lint**

Run: `make lint`

Expected: no new errors. (Pre-existing warnings unrelated to this PR may be present; review them but don't fix here.)

- [ ] **Step 2: Confirm no `v2/preview` token appears in any committed artifact**

Run:
```bash
git ls-files | xargs grep -l 'v2/preview:' 2>/dev/null | grep -v '\.lock$\|\.sum$'
```

Expected: only `provider/pkg/cloud/metadata.json` (the four `aliases` entries) and `provider/cmd/pulumi-resource-pulumiservice/schema.json` (alias entries inside resource bodies). Anywhere else means a stale reference; investigate before opening the PR.

- [ ] **Step 3: Re-run the schema dispatch test**

Run: `cd provider/pkg && go test -count=1 ./provider/... -run TestProvider`

Expected: PASS. (The test at `provider/pkg/provider/provider_test.go` exercises GetSchema and the v2 dispatch routing.)

---

## Task 9: CHANGELOG entry

**Files:**
- Modify: `CHANGELOG.md`

Per project policy, this rename is user-facing because anyone who has imported `pulumiservice.v2.preview.*` types would see the resource type token change. The Pulumi alias migrates state, but SDK names move — that's a CHANGELOG-worthy change.

- [ ] **Step 1: Add an `Unreleased` entry**

If `## Unreleased` doesn't exist yet, add it directly under the `# CHANGELOG` header. Add this bullet under `### Improvements`:

```
- Renamed the `v2/preview` module to `v2/escPreview` for the four `preview/environments`-derived ESC resources (`Environment`, `EnvironmentTag`, `RevisionTag`, `Webhook`). Existing stacks rebind via Pulumi resource aliases; SDK consumers should update imports from `pulumiservice.v2.preview.*` to `pulumiservice.v2.escPreview.*`.
```

- [ ] **Step 2: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs: changelog entry for escPreview rename"
```

---

## Task 10: Update auto-memory

**Files:**
- Modify or add: `~/.claude/projects/-Users-lukeward-Workspace-psp-pulumi-pulumiservice/memory/project_pulumicloud_v1_v2.md` and `MEMORY.md`

- [ ] **Step 1: Append a short note to the existing v1/v2 architecture memory**

In `~/.claude/projects/-Users-lukeward-Workspace-psp-pulumi-pulumiservice/memory/project_pulumicloud_v1_v2.md`, add this paragraph under the existing "How to apply" section:

```
**Module naming for preview-API resources:** When a candidate URL is under
`/api/preview/...` and produces a CRUD-shaped resource that would collide
with an existing module on bare type names, give it a domain-scoped module
suffix (`escPreview` for ESC; pattern: `<existing-module>Next`) rather than a
generic `preview` module. Generic `preview` was the previous convention
and was retired because it stacked awkwardly on the existing `v2`
namespace and obscured the relationship to the production module. See
`tasks/2026-05-06-rename-v2-preview-to-escPreview.md`.
```

(No update needed to `MEMORY.md` — the existing entry still points to the right file.)

---

## Task 11: Open the PR

- [ ] **Step 1: Push the branch**

```bash
git push -u origin lward/pcp
```

(If the branch is already tracking, `git push` is enough.)

- [ ] **Step 2: Open the PR**

Use `gh pr create` with this body (per project CLAUDE.md: terse, no customer info):

```
gh pr create --title "v2: rename preview module to escPreview" --body "$(cat <<'EOF'
## Summary
- Renames the four `preview/environments`-derived v2 resources from `pulumiservice:v2/preview:*` to `pulumiservice:v2/escPreview:*` (Environment, EnvironmentTag, RevisionTag, Webhook).
- Adds Pulumi resource aliases so existing state under the old tokens rebinds without churn.
- Single-entry change in the scaffolder's `moduleAliases` table; everything else is auto-regenerated.

## Test plan
- [x] `provider/pkg/cloud/metadata_test.go` pins both the new tokens and the aliases.
- [x] `make lint` clean.
- [x] `cd provider/pkg && go test -short ./...` passes.
- [x] `make provider && make build_sdks` regenerates the embedded schema and all five language SDKs; obsolete `sdk/*/v2/preview/` dirs removed.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 3: Verify the PR shows only the expected files**

Open the PR's Files Changed view. Expected files:
- `provider/tools/scaffold-metadata/main.go` (one-line alias change)
- `provider/pkg/cloud/metadata.json` (4 token + 4 aliases changes)
- `provider/pkg/cloud/metadata_test.go` (new file)
- `provider/cmd/pulumi-resource-pulumiservice/schema.json` (regen — substantial diff under the affected resources only)
- `sdk/nodejs/v2/escPreview/...`, `sdk/python/.../v2/escpreview/...`, `sdk/go/.../v2/escPreview/...`, `sdk/dotnet/V2/EscPreview/...`, `sdk/java/.../v2/escpreview/...` (new dirs)
- `sdk/*/v2/preview/...` deletions
- `CHANGELOG.md` (one bullet)
- (conditional) `provider/pkg/provider/provider.go` if Task 7 Step 2 ran

Anything else means a stray edit slipped in — investigate before merging.
