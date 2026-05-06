# O3 + O4 — Plan for 2026-05-06

**Goals from the project doc:**
- **O3** Drift detected by the system. CI fails when an op lands without a decision (resource / function / method / `_excluded`).
- **O4** v0 → v2 customers upgrade without state loss. Aliases populated for every v0 resource that has a v2 counterpart.

**What's NOT in scope today:**
- Adding metadata-driven `functions:` and methods support (post-preview maintenance debt). Today's classifications use `_excluded` with reason taxonomy.
- Per-resource `availableOnCompletion` polling (deferrable to follow-up).

**Architecture:**
- O3 enforcement lands in the scaffolder. New `_excluded` schema accepts both prefix-based patterns and exact operationIds, each with a `reason`. Scaffolder fails if any op in spec.json isn't classified into one of: resource CRUD slot, deprecated, `_excluded`. CI gets a `go generate && git diff --exit-code` step.
- O4: extend metadata's `aliases` field to be auto-populatable from a small `_v0Aliases` map at the top of metadata.json. Scaffolder validates that every alias points at a real prior token. Hard-rename cases (Webhook → OrgWebhook, schedule unification) get documented as manual migration recipes in CHANGELOG, no auto-alias.

**Files touched:**
- `provider/pkg/rest/metadata.go` (schema for `_excluded` patterns)
- `provider/tools/scaffold-metadata/main.go` (enforcement + bootstrap + alias validation)
- `provider/pkg/cloud/metadata.json` (curated `_excluded`, `_v0Aliases`)
- `provider/pkg/rest/resource_test.go` or new `scaffold_metadata_test.go` (asserts every op accounted for)
- `.github/workflows/<existing>.yml` (CI step)
- `CHANGELOG.md` (manual migration recipes for hard cases)

**Sequencing:** A→B→C is critical path; D runs parallel to A/B; E gates on C and D.

---

## Phase A: O3 enforcement mechanism (~2 hrs)

### Task A1: Extend `_excluded` schema to support patterns

**Files:** `provider/tools/scaffold-metadata/main.go`

Today `_excluded` is `[]string`. Replace with a typed array supporting two entry shapes:

```json
"_excluded": [
  {"prefix": "List",    "reason": "data sources, planned as v2 functions"},
  {"id": "Capabilities", "reason": "service-introspection endpoint, not for IaC"}
]
```

Add to `metadataDoc`:

```go
type exclusionEntry struct {
    Prefix string `json:"prefix,omitempty"`
    ID     string `json:"id,omitempty"`
    Reason string `json:"reason"`
}

type metadataDoc struct {
    Version    int                         `json:"version"`
    Package    string                      `json:"package,omitempty"`
    Note       string                      `json:"_note,omitempty"`
    Excluded   []exclusionEntry            `json:"_excluded,omitempty"`
    V0Aliases  map[string]string           `json:"_v0Aliases,omitempty"`
    Resources  map[string]json.RawMessage  `json:"resources"`
}
```

Update `loadMetadata` to parse the new shape. Update `writeMetadata` to emit the new shape (replace the existing `_excluded` block writer).

**Verify:** `go build ./tools/scaffold-metadata/...` succeeds. Existing metadata.json (with empty `_excluded`) still parses.

### Task A2: Track classification of every op

**Files:** `provider/tools/scaffold-metadata/main.go`

In `derive()` (or wherever appropriate), record every op's fate into one of these buckets:
- `resource:<token>:<slot>` (create/read/update/delete) — already tracked via `byNoun`
- `deprecated` — already tracked via `stats.deprecated`
- `excluded:<entry>` — new
- `unaccounted` — new; final report

After the regen merge loop, walk every operationId in `parsedSpec.ops` and check it's in one of the first three buckets. If not:
- Compare against `_excluded` patterns (prefix match) and IDs (exact match); if matched, classify as `excluded`.
- Otherwise, append to `unaccounted` with its operationId, method, path.

If `len(unaccounted) > 0`, scaffolder fails non-zero with a list. Print up to 30 unaccounted ops with method+path; print count for the rest.

```go
if len(unaccounted) > 0 {
    fmt.Fprintf(os.Stderr, "\nscaffold-metadata: %d operationIds are not classified — every op must be a resource CRUD slot, deprecated, or in _excluded:\n", len(unaccounted))
    for i, op := range unaccounted {
        if i >= 30 {
            fmt.Fprintf(os.Stderr, "  ... and %d more\n", len(unaccounted)-30)
            break
        }
        fmt.Fprintf(os.Stderr, "  %-50s %s %s\n", op.id, op.method, op.path)
    }
    os.Exit(1)
}
```

**Verify:** running scaffolder against current spec exits non-zero with ~362 unaccounted ops. (Yes, the test of the enforcement is that it fires.)

### Task A3: CI step for generated artifacts

**Files:** `.github/workflows/lint.yml` (or `test.yml` — pick whichever runs on PR and is required for merge)

Add a step:

```yaml
- name: Verify generated metadata
  run: |
    cd provider/pkg/cloud
    go generate ./...
    cd "$GITHUB_WORKSPACE"
    if ! git diff --exit-code provider/pkg/cloud/metadata.json provider/pkg/cloud/spec.json; then
      echo "::error::metadata.json or spec.json drift detected; run 'go generate ./provider/pkg/cloud/...' locally"
      exit 1
    fi
```

**Verify by inspection:** the workflow file is syntactically valid (yaml lint), the step is in a job that runs on `pull_request`. Don't try to actually trigger it locally — that's what the green PR check will demonstrate.

---

## Phase B: O3 curation (~3 hrs, includes user decisions)

### Task B1: Bootstrap `_excluded` with default categories

**Files:** `provider/pkg/cloud/metadata.json`

Hand-edit the `_excluded` block with these starter patterns. Order matters — first match wins:

```json
"_excluded": [
  {"prefix": "List",     "reason": "list endpoints, planned as v2 functions (data sources)"},
  {"prefix": "Search",   "reason": "search endpoints, planned as v2 functions"},
  {"prefix": "Find",     "reason": "find endpoints, planned as v2 functions"},
  {"prefix": "Cancel",   "reason": "actions, planned as v2 methods"},
  {"prefix": "Trigger",  "reason": "actions, planned as v2 methods"},
  {"prefix": "Approve",  "reason": "actions, planned as v2 methods"},
  {"prefix": "Reject",   "reason": "actions, planned as v2 methods"},
  {"prefix": "Reset",    "reason": "actions, planned as v2 methods"},
  {"prefix": "Refresh",  "reason": "actions, planned as v2 methods"},
  {"prefix": "Restore",  "reason": "actions, planned as v2 methods"},
  {"prefix": "Validate", "reason": "validation, planned as v2 methods"},
  {"prefix": "Encrypt",  "reason": "ESC value actions, planned as v2 methods on Environment"},
  {"prefix": "Decrypt",  "reason": "ESC value actions, planned as v2 methods on Environment"},
  {"prefix": "Open",     "reason": "ESC session actions, planned as v2 methods or functions"},
  {"prefix": "Close",    "reason": "ESC/ChangeRequest session actions, planned as v2 methods"},
  {"prefix": "Complete", "reason": "deployment actions, planned as v2 methods"},
  {"prefix": "Append",   "reason": "deployment log actions, planned as v2 methods"},
  {"prefix": "Bulk",     "reason": "bulk ops, deferred to v2.x"},
  {"prefix": "Poll",     "reason": "polling endpoints, internal — clients use Read"},
  {"prefix": "Accept",   "reason": "invitation acceptance, planned as v2 method"}
]
```

These should cover ~150-200 of the dropped ops by prefix.

### Task B2: Resolve the unknown-verb tail

**Files:** `provider/pkg/cloud/metadata.json`

Run scaffolder. The remaining unaccounted-for list will be the unknown-verb ops + any leftovers. Categorize each into `_excluded` with `id` form. Categories I've previewed:

- ChangeRequest workflow (Apply, Approve, Unapprove, Close, Submit on /api/change-requests/...): planned as v2 methods on a future ChangeRequest resource; for now exclude with reason "ChangeRequest workflow, planned as v2 methods".
- OAuth/SSO setup (InitiateAzureDevOpsOAuth, StartGitHubSetup, AWSSetup, AWSSSOInitiate, AWSSSOSetup, AWSSSOListAccounts, AzureSetup, AzureListAccounts, GCPSetup, GCPListAccounts, InitiateOAuth): exclude with reason "interactive setup flows, not IaC-suitable".
- ESC actions (CheckYAML_esc, HeadEnvironment_*, CheckEnvironment_*, CloneEnvironment, RedeliverWebhookEvent_*, PingWebhook_*, ReassignEnvironmentOwnership, RotateEnvironment, PauseEnvironmentSchedule, ResumeEnvironmentSchedule, CheckEnvironment_*_versions, RetractEnvironmentRevision_*): exclude with reason "ESC actions, planned as v2 methods on Environment".
- Bare-named ops (`Get` GET /api/change-requests/{orgName}/{changeRequestID}, `Update` PATCH same path, `Token` POST /api/oauth/token): exclude with reason "underspecified operationId in upstream spec, may need upstream fix".
- Top-level introspection (Capabilities, Version, FetchRestSpecification, AITemplate): exclude with reason "service-introspection endpoint, not for IaC".
- Audit log actions (ExportAuditLogEventsHandlerV1/V2, ForceAuditLogExport, TestAuditLogExportConfiguration): exclude with reason "audit log actions, planned as v2 methods on AuditLogExportConfiguration".

Add these as `{"id": "<opId>", "reason": "<...>"}` entries.

**For each unaccounted op the scaffolder still flags after B1+B2:** add an exclusion entry with the most accurate reason. If unsure, use reason "unclassified — needs review before v1 GA" — visible flag, not silent.

**Verify:** scaffolder exits 0. Re-run regen — metadata.json updates only if exclusions change anything; spec.json should NOT change (it's pinned post-revert in Phase E).

### Task B3: Add a Go test asserting "no unaccounted ops"

**Files:** `provider/pkg/rest/metadata_test.go` (new) or extend existing test

```go
func TestEverySpecOpIsAccountedFor(t *testing.T) {
    spec, meta := loadFixtures(t)
    // ... compute accounted_for via the same logic as scaffolder ...
    // ... fail with detailed list of unaccounted ops ...
}
```

This is belt-and-suspenders alongside the CI scaffolder check: the test runs in `go test ./pkg/...` so a developer notices before pushing.

If the test logic largely duplicates scaffolder logic, consider exporting a helper (`rest.AccountedForOps(spec, meta) (accounted, unaccounted []string)`) and using it from both places.

**Verify:** `go test ./pkg/rest/...` passes.

---

## Phase C: O4 — populate aliases (~2 hrs, parallelizable with A/B)

### Task C1: Enumerate v0 resources + build mapping table

**Files:** `tasks/v0_to_v2_mapping.md` (working notes, not committed if you prefer)

Sources:
- `provider/pkg/provider/manual-schema.json` — 20 raw v0 resources (token list captured above in this plan)
- `provider/pkg/provider/provider.go` — infer-style v0 resources (search for `infer.Resource[`)

Produce a table:

| v0 token | v2 counterpart | Mapping confidence |
|---|---|---|
| pulumiservice:index:AgentPool | pulumiservice:v2:AgentPool | high (1:1) |
| pulumiservice:index:DeploymentSettings | pulumiservice:v2:DeploymentSettings | high |
| pulumiservice:index:OidcIssuer | pulumiservice:v2:OidcIssuer | high |
| pulumiservice:index:Stack | pulumiservice:v2:Stack | needs verify |
| pulumiservice:index:Environment | pulumiservice:v2:Environment | needs verify (v2 token may include esc/ module) |
| pulumiservice:index:Webhook | ??? | medium — split into OrgWebhook/StackWebhook? |
| pulumiservice:index:DeploymentSchedule | pulumiservice:v2:ScheduledDeployment | medium |
| pulumiservice:index:DriftSchedule | ??? | low — likely unified with ScheduledDeployment |
| pulumiservice:index:TtlSchedule | ??? | low — likely unified |
| pulumiservice:index:EnvironmentRotationSchedule | ??? | low |
| pulumiservice:index:EnvironmentVersionTag | pulumiservice:v2:EnvironmentTag_esc_environments? | medium |
| pulumiservice:index:OrgAccessToken | pulumiservice:v2:OrgToken | medium |
| pulumiservice:index:TeamAccessToken | pulumiservice:v2:TeamToken | medium |
| pulumiservice:index:AccessToken | pulumiservice:v2:PersonalToken | medium |
| (others) | (...) | (...) |

The `pulumiservice:v2:*` token list comes from `python3 -c "import json; ..."` against metadata.json.

**High-confidence pairs get aliased automatically.** Medium/low confidence pairs are documented as manual-migration recipes in Phase E.

### Task C2: Add `_v0Aliases` mechanism to scaffolder

**Files:** `provider/tools/scaffold-metadata/main.go`

Add `_v0Aliases` to `metadataDoc` (already done in Task A1's metadataDoc rewrite). Map shape: `{ "<v2 token>": "<v0 token>" }`.

In `mergeOperations` (or a sibling pass), if an entry's token (or metadata key) appears in `_v0Aliases`, ensure `aliases` includes the v0 token. Write-if-absent semantics — humans can also hand-set additional aliases.

```go
// After other derivations:
if v0, ok := v0Aliases[tok]; ok {
    addAlias(entry, v0)
}
```

**Verify:** `go build ./tools/scaffold-metadata/...` succeeds.

### Task C3: Populate `_v0Aliases` from C1's high-confidence pairs

**Files:** `provider/pkg/cloud/metadata.json`

Hand-edit the new `_v0Aliases` block with only the high-confidence 1:1 mappings:

```json
"_v0Aliases": {
  "pulumiservice:v2:AgentPool":         "pulumiservice:index:AgentPool",
  "pulumiservice:v2:DeploymentSettings": "pulumiservice:index:DeploymentSettings",
  "pulumiservice:v2:OidcIssuer":        "pulumiservice:index:OidcIssuer",
  "pulumiservice:v2:Environment":       "pulumiservice:index:Environment"
}
```

(Final list determined by Task C1's verification.)

Run scaffolder; aliases should auto-populate on the matching resources.

### Task C4: Validate aliases reference real prior tokens

**Files:** `provider/tools/scaffold-metadata/main.go`

The scaffolder can't easily verify v0 tokens exist (they're in a separate file the scaffolder doesn't read), but it CAN verify the v2 side: every key in `_v0Aliases` must appear in `doc.Resources`. Fail if not.

```go
for v2tok := range doc.V0Aliases {
    if _, ok := doc.Resources[v2tok]; !ok {
        fail("_v0Aliases references unknown v2 token %q", v2tok)
    }
}
```

**Verify:** scaffolder runs clean.

---

## Phase D: housekeeping (parallel with A/B/C)

### Task D1: Revert spec.json refresh from earlier session

**Files:** `provider/pkg/cloud/spec.json`

This file picked up an unrelated 333-line diff from the requireImport regen. Revert before starting today's work:

```bash
git checkout provider/pkg/cloud/spec.json
```

Run scaffolder (`go generate ./provider/pkg/cloud/...`) immediately to confirm metadata.json doesn't depend on the spec changes (if it does, regen and verify the metadata.json delta is acceptable).

---

## Phase E: integration + CHANGELOG

### Task E1: Document manual migration recipes for hard renames

**Files:** `CHANGELOG.md`

Add under the existing `## Unreleased > ### Breaking Changes`:

```markdown
- v0 → v2 migration: most resources alias automatically (you'll only need to update type tokens — `pulumiservice.Foo` → `pulumiservice.v2.Foo`). The following renames require an explicit `aliases` ResourceOption on first migration:
  - `pulumiservice.Webhook` → `pulumiservice.v2.OrgWebhook` or `pulumiservice.v2.StackWebhook` (split by scope)
  - `pulumiservice.DeploymentSchedule` / `DriftSchedule` / `TtlSchedule` / `EnvironmentRotationSchedule` → `pulumiservice.v2.ScheduledDeployment` (unified)
  - `pulumiservice.OrgAccessToken` → `pulumiservice.v2.OrgToken`
  - `pulumiservice.TeamAccessToken` → `pulumiservice.v2.TeamToken`
  - `pulumiservice.AccessToken` → `pulumiservice.v2.PersonalToken`
  
  Migration recipe (TypeScript example):
  ```ts
  new pulumiservice.v2.OrgWebhook("...", {...}, {
    aliases: [{ type: "pulumiservice:index:Webhook" }]
  });
  ```
```

(Refine after Task C1 confirms the actual hard-rename list.)

### Task E2: Final verification

**Verify:**
1. `cd provider && go test ./pkg/...` — all green.
2. `cd provider/pkg/cloud && go generate ./...` — exits 0, no diff in metadata.json or spec.json (everything stable).
3. `git diff --stat` is contained to: metadata.go (Task A1 schema), main.go (Tasks A2/C2/C4), metadata.json (Tasks B1/B2/C3), the lint.yml CI step, CHANGELOG.md, and the new test file.
4. Spot-check a few aliased resources: their generated schema includes the alias.

---

## Self-review

- **Spec coverage:** Every outcome (O3, O4) has tasks. The "no scope expansion" CLAUDE.md rule is honored — no metadata-driven functions/methods today, no per-resource availability hooks today.
- **Placeholders:** Tasks B2 and C1 contain decisions ("categorize each into the right reason," "build the mapping table"). These are explicit human-input checkpoints, not placeholders — the WORK has to happen, the values come from inspection.
- **Risk: spec.json drift mid-session.** Phase D1 pins it. If a fresh op lands during the session, scaffolder will catch it — that's the enforcement working.
- **Risk: hard-rename count larger than expected.** If C1 reveals >5 hard-renames, consider whether some can be aliased anyway by accepting input-shape coercion. Decide case-by-case.
- **Risk: CI workflow file is auto-generated.** Several `.github/workflows/*.yml` files are warned as "autogenerated by ci-mgmt." If `lint.yml` is in that set, the step will be wiped on next ci-mgmt run. Land the step in `v2-tests.yml` instead (which is hand-curated per its banner).

## Execution order recommendation

Sequential within phases; parallel between A/B/D and C is fine but avoid letting them both touch metadata.json simultaneously. Suggested order:

1. **D1** (revert spec.json) — 2 min, must be first
2. **A1, A2** — 1.5 hrs — scaffolder mechanism
3. **B1, B2** — 2-3 hrs — curate exclusions, iterate against scaffolder until it exits 0
4. **C1, C2, C3, C4** — 2 hrs — aliases (can also start during B if you have parallel attention)
5. **A3** (CI step), **B3** (Go test), **E1** (CHANGELOG), **E2** (final verify) — 1 hr

Total: 6-7 hours focused work.
