# Maintaining the Pulumi Service Provider

This document is the operational playbook for maintainers. It covers the
day-to-day reality of keeping the provider in sync with Pulumi Cloud as
the OpenAPI spec evolves — the cadence, the decision criteria, the
end-to-end workflow for each common change type.

If you're adding a single new resource or reacting to a spec diff, this
is the document to read. `CLAUDE.md` is the short-form version for
agents; this doc is the authoritative long-form.

## Mental model

The provider is **generated, not hand-written**. Three inputs compose
into the shipped artifacts:

```
┌─────────────────────────────┐   ┌─────────────────────┐
│ provider/spec/               │   │ provider/           │
│   openapi_public.json        │   │   resource-map.yaml │
│   (pinned Pulumi Cloud spec) │   │   (editable mapping)│
└──────────────┬───────────────┘   └──────────┬──────────┘
               │                              │
               └──────────────┬───────────────┘
                              ▼
            ┌────────────────────────────────────┐
            │ provider/cmd/pulumi-gen-pulumiservice│
            │ (generator: parse, coverage,       │
            │  schema emit, metadata emit)       │
            └───────────────┬────────────────────┘
                            │
         ┌──────────────────┼──────────────────────┐
         ▼                  ▼                      ▼
   bin/schema.json    bin/metadata.json     bin/coverage-report.md
         │                  │
         ▼                  ▼
   sdk/{nodejs,py,go,        embedded into
        dotnet,java}         pulumi-resource-pulumiservice
```

**Zero hand-coded resources.** Every supported resource is expressed
in `resource-map.yaml`; `provider/pkg/` has no `customresources/`
package. When a Pulumi Cloud API pattern doesn't fit the current
metadata shape, the right response is to extend the metadata schema
(see "Decision tree" below) — not to add per-resource Go.

**Never edit `bin/*.json` or `sdk/*` by hand.** They're generated
outputs; direct edits will be overwritten on the next
`make v2_provider` / `make build_sdks`.

## Cadence

**Recommended:** refresh the spec once per week, or whenever Pulumi
Cloud announces a release that exposes a new endpoint you need.

There's no CI drift job today; refresh is a manual cadence a
maintainer owns. Whoever merges a resource-map change has responsibility
to also bump the spec if it's gotten stale.

## The refresh workflow

This is the full loop for picking up new Pulumi Cloud spec operations.

### 1. Pull the latest spec

```bash
make update_spec
```

Requires a sibling checkout of `pulumi-service` at
`../pulumi-service/` with a current `main`. Copies
`pulumi-service/pkg/apitype/spec/openapi_public.json` into
`provider/spec/openapi_public.json` in this repo.

Review the diff:

```bash
git diff provider/spec/openapi_public.json | head -200
```

### 2. See what changed from the coverage gate's perspective

```bash
make coverage_report
head -40 bin/coverage-report.md
```

The report tells you:

- **Unmapped operations** — new operationIds the spec has that the map
  doesn't cover. Each one needs a decision (next section).
- **Duplicate claims** — usually benign (a shared read/delete endpoint,
  a list op used as both readVia and a function). Review to confirm.
- **TODO markers in the map** — placeholders you left on a previous
  iteration. Resolve or re-justify.
- **Mapped** / **Excluded** counts — bookkeeping.

The `coverage_report_strict` variant fails with non-zero exit on any
unmapped op. CI should run that one once we wire CI.

### 3. Triage each unmapped op

Use the decision tree in the next section. For each unmapped op,
either:

- Add it to a `resources:`/`functions:`/`methods:` block under the
  correct sub-module in `provider/resource-map.yaml`, OR
- Add it to `exclusions:` with a written reason (1 sentence).

Err on the side of **excluding with a reason** if unsure. Exposing a
resource is a user-visible commitment; excluding is a soft default that
a later maintainer can revisit.

### 4. Regenerate

```bash
make v2_provider   # generator → schema.json + metadata.json, rebuilds binary,
                   # and writes the binary's final schema back to
                   # provider/cmd/pulumi-resource-pulumiservice/schema.json

make build_sdks    # regenerates sdk/{nodejs,python,go,dotnet,java}
```

### 5. Verify

```bash
make coverage_report_strict       # must exit 0
cd provider && go test ./...      # unit tests
cd sdk/nodejs && yarn install && yarn run tsc --noEmit   # TS SDK compiles
```

Plus eyeball the TypeScript SDK classes for any resource you just
added — the generated doc comments and property shapes are what end
users will consume, and OpenAPI descriptions don't always translate
idiomatically. Iterate on the `doc:` / property-naming choices in the
map until the generated SDK reads well.

### 6. Commit

One commit per logical change. At minimum, one for the spec bump and
one per new resource added. Keep `bin/*` and `sdk/*` in commits where
they were regenerated, so reviewers can see the full effect.

## Decision tree: resource / function / method / exclude / extend-metadata

When a new operationId appears, pick one bucket.

### Is it a declarative thing with a lifecycle? → **Resource**

Criteria:
- Has a noun-shaped URL (`/api/…/foos/{fooId}`).
- Create/Read/Update/Delete (or a subset) map 1:1 to HTTP verbs.
- An end user would reasonably want to declare *N* of these in a
  Pulumi program and let Pulumi reconcile them.

Examples:
- `CreateAgentPool`/`GetAgentPool`/`PatchOrgAgentPool`/`DeleteOrgAgentPool`
  → `pulumiservice:orgs/agents:AgentPool`.
- `CreateWebhook_esc_environments` → one of the scopes of
  `pulumiservice:stacks/hooks:Webhook`.

Edge cases:
- **Some verbs missing.** If only create + delete exist, you can use
  `readVia: { operationId: ListFoos, filterBy: id }` to fall back to
  list-and-filter.
- **PATCH is the create path (upsert).** Map `create: UpdateFoo,
  update: UpdateFoo` and duplicate-claim the op in both slots.
- **Polymorphic** (one resource, multiple CRUD paths based on a
  discriminator, e.g., Webhook scope or Team type). Use the
  polymorphic operation form (`case: <field>, scopes: { a: {create:…,
  read:…}, b: {create:…} }`) and give the resource a `discriminator:`
  block.

### Is it a read-only lookup returning data? → **Function**

Criteria:
- GET only.
- Doesn't mutate anything.
- Returns info that an end user might want to inline into a program.

Examples:
- `ListAgentPools` → `pulumiservice:orgs/agents:listAgentPools`.
- `GetAuthPolicy` → `pulumiservice:orgs/oidc:getAuthPolicy`.

### Is it an imperative action on an existing resource? → **Method**

Criteria:
- POST (usually) to an action path like `/api/…/foos/{fooId}/scan`.
- Operates on a specific resource instance; takes no independent
  identity.
- Doesn't mutate the resource's declarative shape (so shouldn't be
  part of the resource's update).

Examples:
- `ScanAccount` → `Account.triggerScan` method on
  `pulumiservice:orgs/insights:Account`.
- `RegenerateThumbprints` → `Issuer.regenerateThumbprints` method on
  `pulumiservice:orgs/oidc:Issuer`.

### Is it engine-internal, admin-only, or interactive? → **Exclude**

Categories that belong under `exclusions:`:
- **Stack update lifecycle** (AppendUpdateLogEntry, CancelUpdate,
  RunDeployment, …) — managed by the Pulumi engine itself.
- **CLI/tooling endpoints** (BatchEncryptValue, OpenEnvironment,
  CheckEnvironment, …) — consumed by `pulumi` CLI, not by declarative
  IaC.
- **Interactive wizards** (AWSSetup, AzureSetup, AWSSSOSetup, …) —
  OAuth-style bring-up that doesn't fit CRUD.
- **Admin/internal** (admin backfill endpoints, internal reporting
  views).
- **Legacy/deprecated** (preview paths superseded by GA paths; v1 API
  versions).

Always attach a 1-sentence reason:

```yaml
exclusions:
  - { operationId: CancelUpdate_update, reason: "stack update lifecycle: engine-managed, not declarative" }
```

### Is it irreducible? → **Extend the metadata; don't add Go**

**v2 ships zero hand-coded resources.** There is no
`customresources/` package and no escape hatch. Every operational
pattern the v1 provider needed is now expressible in the metadata
schema; when the spec grows a new pattern that none of the primitives
cover, the right answer is to add a new primitive — not to add a
per-resource Go implementation.

The existing primitives (all declared in
`provider/pkg/runtime/metadata.go`) cover:

| Pattern | Primitive | Example user |
|---|---|---|
| Upsert (PUT/PATCH is both create and update) | Reuse operationId in `create` + `update` | `LogExport`, `Settings`, `TeamStackPermission` |
| Tombstone-style delete via the update op | `delete: { operationId: …, bodyOverride: {…} }` | `TeamStackPermission` (permission: 0) |
| Identity field that sits in body for POST, path for GET/PATCH/DELETE | Property `createSource:` (+ optional `createFrom:` for wire-level rename) | `Stack`, `Environment` |
| No per-item GET — read piggybacks on parent | `readVia: { operationId: …, extractField: …, keyBy: … }` | `Tag`, `Tags` |
| Batch delete iterates over a map property | `delete: { operationId: …, iterateOver: <prop>, iterateKeyParam: <path-placeholder> }` | `Tags` |
| Non-JSON request/response bodies | Property `source: rawBody` + op `rawBodyFrom:` / `rawBodyTo:` + `contentType:` | `Environment` (application/x-yaml) |
| Two-step create (POST then follow-on PATCH) | `postCreate:` on the resource | `Environment` (POST empty, PATCH YAML) |
| Polymorphic dispatch by discriminator | `case:` + `scopes:` in operations; `discriminator:` on the resource | `Webhook` (org/stack/esc scopes), `Team` (pulumi/github) |

**Adding a new primitive** (when the spec shows an 8th case the seven
above don't cover):

1. Extend `CloudAPIOperation` or `CloudAPIResource` (or `CloudAPIProperty`
   or `CloudAPIReadVia`) in `provider/pkg/runtime/metadata.go` with the
   new field.
2. Wire the parser in `provider/pkg/gen/metadata.go` — usually a
   one-line addition inside `buildOp` or `buildReadVia`.
3. If the field is a YAML key inside an operations block, register it
   in `provider/pkg/gen/parse.go`'s `metadataKeys` so the coverage
   gate ignores it correctly.
4. Implement the dispatch behavior in `provider/pkg/runtime/dispatch.go`.
5. Add a test exercising the primitive end-to-end.
6. Use it from `provider/resource-map.yaml` on the affected resource.

This is where engineering effort goes. The payoff is that the *next*
resource with the same pattern lands in five lines of YAML instead of
five hundred lines of Go.

## Adding a new resource — step by step

For a resource that fits the declarative pattern (the common case).

1. **Identify the operations** in the spec. Run the inventory commands
   in [Spec archaeology](#spec-archaeology) below.
2. **Pick a token.** `pulumiservice:<module>:<Name>`. Follow the
   anti-stutter rule — the module noun shouldn't appear in the type
   name. Read the fully-qualified token out loud; if it stutters,
   rename the type.
3. **Add the entry** to `provider/resource-map.yaml` under the right
   module:
   ```yaml
   orgs/foos:
     resources:
       Foo:
         doc: One-sentence description.
         operations:
           create: CreateFoo
           read:   GetFoo
           update: UpdateFoo
           delete: DeleteFoo
         id:
           template: "{organizationName}/{fooId}"
           params: [organizationName, fooId]
         forceNew: [organizationName]
         properties:
           organizationName:
             from: orgName
             type: string
             source: path
             required: true
             doc: Organization that owns this foo.
           name:
             type: string
             source: body
             required: true
           # … all user-facing + output-only properties
           fooId:
             from: id
             type: string
             source: response
             output: true
   ```
4. **Regenerate + verify:**
   ```bash
   make v2_provider
   make build_sdks
   cd provider && go test ./...
   ```
5. **Write a YAML example** under `examples/yaml-<name>/Pulumi.yaml`
   (per-resource smoke test) and/or extend a canonical example under
   `examples/canonical/*` if the resource fits an existing scenario.
6. **Add a test** in `examples/examples_yaml_test.go`.
7. **CHANGELOG entry** under `## Unreleased` or the current pre-release
   section, categorized under `### Improvements`.
8. **Spot-check the generated TS SDK.** Open
   `sdk/nodejs/<module>/<name>.ts` and confirm the shape reads
   idiomatically. If property names, required/optional flags, or doc
   strings feel wrong, iterate on the map and regenerate.

## Changing an existing resource

Most spec changes are additive (new optional field). Workflow:

1. Pull spec: `make update_spec`.
2. Coverage gate won't flag additive field changes — it only tracks
   operationIds, not schema shapes. So this one's on the human
   reviewer: read the spec diff in
   `provider/spec/openapi_public.json` for any resources you care
   about.
3. If new fields should be exposed, add them to the resource's
   `properties:` block. Set `source: body` / `required: <bool>` as
   appropriate.
4. Regenerate + verify + CHANGELOG.

**Breaking spec changes** (renamed field, removed field, type change)
are rarer. Apply the equivalent change to the map. If the old name
needs to stay available for v2.x users, use the `aliases:` field on
the property to accept both:

```yaml
properties:
  name:
    from: newName
    source: body
    aliases: [oldName]
```

## Removing a resource

When Pulumi Cloud retires an endpoint:

1. `make update_spec` — spec no longer has the operationId.
2. Coverage gate flags stale claim (coverage gate's stale-mapping
   detection is currently incomplete; for now grep the map for the
   old operationId and remove the mapping entry).
3. If the Pulumi resource itself is going away, mark it as a breaking
   change in the next major (3.x), not as a silent removal. For a
   patch release, hide the underlying behavior but keep the resource
   declared so existing programs still parse.

## Upstream migration (long-term)

`provider/resource-map.yaml` is a stopgap. The durable home for the
mapping is upstream in `pulumi-service/specification/src/` as Java
annotations that emit `x-pulumi-*` extensions into the spec during
`specification/generate_spec.sh`.

**Trigger for migration:** the map has been stable for a full minor
release (2.1.x → 2.2.0 ships without significant map churn) and the
remaining hand-authored fields in the map (property renames, doc
overrides) are few enough to fit in annotations rather than a side
file.

**When that lands:**
- Generator parses `x-pulumi-*` directly from the spec.
- `resource-map.yaml` collapses to overrides-only (or disappears).
- Cloud engineers own spec-plus-mapping in a single PR. Drift is
  impossible by construction.

Not in scope for any 2.x release without a separate planning pass.

## Versioning

- **2.x.y patch:** bug fix, doc change, spec refresh with no user-visible
  API change.
- **2.x.0 minor:** new resources, new properties on existing resources,
  new functions or methods. Preview-API dependencies may land here even
  if they're still under `/api/preview/` upstream.
- **3.0.0 major:** another sub-module repartition, renaming strategy
  change, or any other user-facing break at v2.x's scale. We don't plan
  these; if we're thinking about one, that itself needs a planning doc.

## Verification checklist

Before opening a PR that touches the provider:

- [ ] `make coverage_report_strict` exits 0.
- [ ] `cd provider && go test -count=1 ./...` is green.
- [ ] `make v2_provider && make build_sdks` completes cleanly for all
      five languages.
- [ ] TypeScript SDK compiles (`cd sdk/nodejs && yarn run tsc
      --noEmit`).
- [ ] For any new resource or property: YAML example added, test
      wired, CHANGELOG entry written.
- [ ] For any PR that bumps `provider/spec/openapi_public.json`: the
      coverage report shows every new operationId triaged (mapped or
      excluded with reason).
- [ ] No `TODO:` markers left in `resource-map.yaml` operations.
- [ ] Custom-resource count is still ≤5 (or you're explicitly
      justifying why it grew).

## Spec archaeology

Useful one-liners for investigating the spec.

**List all operations under a path prefix:**
```bash
python3 -c "
import json
with open('provider/spec/openapi_public.json') as f: d = json.load(f)
for path, methods in sorted(d['paths'].items()):
  if '/api/orgs/' in path:
    for m, op in methods.items():
      if m in ['get','post','put','patch','delete']:
        print(f'{m.upper():6} {path:80} -> {op.get(\"operationId\")}')
"
```

**Find operations whose ID contains a substring:**
```bash
python3 -c "
import json
with open('provider/spec/openapi_public.json') as f: d = json.load(f)
for path, methods in d['paths'].items():
  for m, op in methods.items():
    if m in ['get','post','put','patch','delete']:
      oid = op.get('operationId','')
      if 'Webhook' in oid:
        print(f'{m.upper():6} {path} -> {oid}')
"
```

**Show a component schema (for deciding property types):**
```bash
python3 -c "
import json
with open('provider/spec/openapi_public.json') as f: d = json.load(f)
print(json.dumps(d['components']['schemas'].get('AgentPool'), indent=2))
"
```

## Running the refresh with Claude Code

The spec refresh is the most repetitive maintenance work this provider
has. Hand the prompt below to Claude Code (or paste it into an
interactive session) and it will run the whole loop — pull the spec,
triage each new operation, populate property blocks, regenerate, and
run the verification gauntlet — leaving the changes in the working
tree for a human to review and commit.

This isn't a replacement for understanding the workflow — it's a tool
for not re-doing the mechanical parts by hand. The review step at the
end is still load-bearing.

### Refresh prompt (copy verbatim)

```
Refresh the Pulumi Service Provider against the latest Pulumi Cloud
OpenAPI spec. Follow docs/MAINTAINING.md for the authoritative
workflow; the steps below are the self-check.

PREREQUISITES
- A sibling checkout of pulumi-service at ../pulumi-service with the
  target revision on main.
- Clean working tree in this repo (git status --short shows nothing).
- make, pulumi CLI, yarn, python3, go all on PATH.

PROCEDURE
1. Run `make update_spec`. Inspect the diff in
   provider/spec/openapi_public.json — report the top-level summary
   (paths added, paths removed, paths changed) before proceeding.
2. Run `make coverage_report` and read bin/coverage-report.md. Note
   the counts (mapped / excluded / unmapped / duplicates / TODOs).
3. For every newly-unmapped operationId, use the decision tree in
   docs/MAINTAINING.md ("Decision tree: resource / function / method
   / exclude / custom") to classify:
     - Resource   → add under `resources:` in the correct sub-module
                    in provider/resource-map.yaml
     - Function   → add under `functions:`
     - Method     → add under `methods:` on the owning resource
     - Exclude    → add to top-level `exclusions:` with a 1-sentence
                    reason
     - Custom     → STOP; bring to the user (see DECISION AUTHORITY
                    below)
4. For any new resource, populate its full property block against the
   corresponding component schema in
   provider/spec/openapi_public.json. Set `source`, `required`,
   `secret`, `forceNew`, `default`, `enum`, and `doc` per property.
   Apply the anti-stutter naming rule from CLAUDE.md when choosing
   the type name (read the fully-qualified token out loud; if it
   stutters, rename).
5. Run `make v2_provider && make build_sdks`. Fix any generator or
   SDK-generation errors before proceeding.
6. Run `cd provider && go test -count=1 ./...`. All tests must pass.
7. Compile-check the TypeScript SDK:
     cd sdk/nodejs && yarn install && yarn run tsc --noEmit
8. For each newly-added resource, open
   sdk/nodejs/<module>/<name>.ts and eyeball the generated doc
   comments, property names, optional/required flags, and type
   shapes. If anything reads awkwardly, iterate on the resource-map
   entry (`doc:`, property renames via `from:`, secret flags) and
   regenerate.
9. Update CHANGELOG.md under the current unreleased section
   (or create a new `## Unreleased` block if none exists). One
   bullet per user-visible change. Skip pure-internal changes.
10. STOP. Do NOT commit. Report back with:
    - Coverage delta: before/after counts (total / mapped / excluded
      / unmapped / TODOs).
    - The operationIds you mapped, grouped by resource / function /
      method, each with its sub-module and the chosen Pulumi name.
    - The operationIds you excluded, each with its 1-sentence reason.
    - Any operationIds you couldn't confidently classify (these go
      to the user; see DECISION AUTHORITY).
    - Files modified (paths only, not the diff itself).

DECISION AUTHORITY — act autonomously on:
- Excluding administrative, internal, CLI-managed, engine-managed,
  interactive-wizard, or deprecated endpoints, with a written reason.
- Adding obvious CRUD resources where all four verbs have standard
  POST/GET/PATCH/DELETE endpoints under the same noun path.
- Adding obvious functions (GET-only, returns data, no mutation).
- Adding obvious methods (POST to an action path under an existing
  resource's URL, with no independent identity).

DECISION AUTHORITY — STOP and bring to the user for:
- Any pattern that needs a NEW metadata primitive (beyond the ones
  documented in "Is it irreducible?" above). Adding a primitive means
  touching runtime.CloudAPI* types, the generator, and dispatch.go —
  that's an architectural decision, not a routine mapping.
- Any resource with polymorphic operations (Webhook-style scope
  discriminator, Team-style type discriminator) — these need
  design review.
- Any operation whose classification is ambiguous (could reasonably
  be resource vs function vs method).
- Any spec change that would be breaking for existing v2 users
  (property renamed, property type changed, property removed, or a
  required field added to an existing resource).
- Any operation whose component schema has nested complex objects
  the map's `ref:`/`types:` mechanism can't obviously express.

HARD RULES
- Never hand-edit provider/cmd/pulumi-resource-pulumiservice/
  schema.json or metadata.json. They're generated by
  `make v2_provider`.
- Never hand-edit sdk/**. Regenerated by `make build_sdks`.
- Never leave a `TODO:` marker in a new resource-map entry without
  STOPping and reporting it.
- Never mark an exclusion without a 1-sentence reason.
- Never commit. Leave the working tree dirty for the user to review.
```

### When to use it

- **Weekly, or on-demand when Pulumi Cloud ships a release that
  exposes a new endpoint you need.** Run the prompt, review the
  triage, commit.
- **Not for targeted work.** If you just want to add one specific
  resource or debug one specific failure, a focused ad-hoc prompt is
  better — this prompt is optimized for "pull the spec and react to
  everything that changed."

### After running it

The prompt is deliberately hands-off on commits so a human can review
the triage decisions. Expected review pass:

1. Read the summary the agent reports.
2. Spot-check 2–3 of its classification decisions against the
   decision tree — especially any borderline "resource vs function
   vs method" calls.
3. Read the CHANGELOG entry.
4. Read the diff in `provider/resource-map.yaml`. It should be a
   clean additive change (new entries under modules, new entries in
   exclusions) plus CHANGELOG.
5. Commit in logical groups — spec bump + map additions + regenerated
   artifacts, plus a separate commit per "real" new resource with
   its example and test.

## Troubleshooting

- **Coverage gate fails with 'unmapped operation' right after
  `make update_spec`:** expected; triage per the decision tree.
- **`pulumi package gen-sdk` fails with "missing language block":**
  the emitter's `languageBlocks()` function is out of sync — check
  `provider/pkg/gen/provider.go`.
- **SDK generates but has empty class body for a resource:** the
  resource-map entry has no `properties:` block, or its `create`
  operation has a TODO marker. Fill in properties.
- **`pulumi preview` says "exactly one of X, Y, Z must be set" on a
  resource with no obvious check:** the declarative `checks:` block
  in the map is being enforced by the runtime `Check()` RPC. That's
  working as intended.
- **After adding a property, SDK has the property but a `pulumi up`
  rejects the input:** did you regenerate the SDK? `make build_sdks`.
- **Added metadata primitive but dispatch ignores it:** check
  `provider/pkg/runtime/dispatch.go` — the dispatch path for the verb
  in question must explicitly consult the new field. Parse + emit
  plumbing alone isn't enough.
- **New resource appears in metadata.json but not schema.json:** the
  resource's `create:` operationId must resolve in the OpenAPI spec;
  otherwise the schema emitter skips it. A `TODO:` marker produces the
  same outcome — resolve it before expecting the resource to emit.
