# v2 Provider — Architecture & Migration Design

**Status:** Draft, soliciting feedback
**Author:** Luke Ward
**Date:** 2026-04-30
**Related:** PR [#763](https://github.com/pulumi/pulumi-pulumiservice/pull/763),
[`docs/MAINTAINING.md`](./MAINTAINING.md),
[`docs/UPGRADE-v1-to-v2.md`](./UPGRADE-v1-to-v2.md)

This doc captures the architecture choices behind the v2 rewrite and
the migration story for existing users. It exists to gather review
feedback before the wholesale cut-over lands. It deliberately stays
high-level — `MAINTAINING.md` covers the maintainer playbook and
`UPGRADE-v1-to-v2.md` covers the user-facing recipe; this doc is for
the *why*, not the *how*.

## 1. Context & Motivation

v1 is hand-coded. Every supported Pulumi Cloud resource lives as a Go
file with bespoke schema, CRUD methods, and tests. Each new endpoint
needs a PR with boilerplate, and the provider drifts from the Cloud
API as new spec fields are quietly missed.

Three things changed that make a generated approach feasible:

1. **Pulumi Cloud has a maintained OpenAPI spec.** The same spec that
   backs the public REST docs is consumable by tools.
2. **`pulumi-go-provider` is a viable runtime.** It supports
   runtime-built schema and arbitrary CRUD dispatchers, which
   removes the need for a separate codegen tool.
3. **The maintenance backlog forced the issue.** Hand-coding a Go
   resource per new Cloud endpoint can't keep up with the rate of
   spec changes.

**Goal:** make the provider generated rather than hand-coded; reduce
"add a resource" from a Go PR to a YAML edit; eliminate silent drift
between the spec and the provider.

## 2. Architecture

```
              ┌─────────────────────────────────────┐
              │  provider/pkg/embedded/             │
              │                                     │
              │  openapi_public.json   ◄── pinned   │
              │  resource-map.yaml     ◄── edited   │
              └────────────────┬────────────────────┘
                               │ //go:embed
                               ▼
            ┌──────────────────────────────────────┐
            │  pulumi-resource-pulumiservice       │
            │                                      │
            │  gen/      → schema, metadata        │
            │  runtime/  → CRUD dispatcher         │
            │  provider/ → pulumi-go-provider lit  │
            └──────────────────────────────────────┘
                               │
                ┌──────────────┴──────────────┐
                ▼                             ▼
         GetSchema (lazy)           CRUD calls (HTTPS)
                                              │
                                              ▼
                                   Pulumi Cloud REST API
```

Two inputs, both `//go:embed`-ed into the runtime binary:

- `provider/pkg/embedded/openapi_public.json` — pinned copy of the
  Pulumi Cloud OpenAPI 3.0.3 spec.
- `provider/pkg/embedded/resource-map.yaml` — editable mapping from
  operationIds to Pulumi resources/functions/methods, plus
  per-property metadata (renames, secrets, defaults, force-new,
  validation checks).

One binary: `pulumi-resource-pulumiservice`. Plain `go build` produces
it. There is no separate codegen step and no committed `schema.json`.

The runtime (`provider/pkg/runtime/`) is a metadata-driven CRUD
dispatcher: at startup it parses the embedded inputs into typed
metadata, and at request time it routes Pulumi RPCs (Create/Read/
Update/Delete/Check/Diff) through that metadata to HTTPS calls
against the Cloud API. `GetSchema` is lazy — the Pulumi schema is
emitted from the same embedded inputs on first call and cached.

A coverage gate (`go test ./provider/pkg/embedded/...`) fails CI if
any operationId in the spec is unclaimed in `resource-map.yaml`. The
same check fires from `GetSchema` at runtime. This is the property
that prevents drift.

The extension point — when a Cloud API pattern doesn't fit the
current shapes — is `provider/pkg/runtime/metadata.go` (plus its
parser and dispatcher). There is no `customresources/` package and
no per-resource Go.

## 3. Key Decisions

### 3.1 Single binary, no generator/runtime split

**Considered:** a `pulumi-gen-pulumiservice` codegen tool that emits
`schema.json` + `metadata.json` at build time, consumed by a separate
runtime binary. This is the conventional Pulumi provider layout.

**Chose:** embed both inputs in the runtime binary; emit schema
on-demand via `GetSchema`.

**Why:** `go build` "just works" — no Make ordering, no codegen step
to forget, no two-binary release. It also opens the door to provider
parameterization (different spec → different schema for self-hosted
Pulumi Cloud) without a rebuild. The cost is a small first-call
latency on `GetSchema`; in practice it's cached after first use and
not on the hot path.

### 3.2 `resource-map.yaml` as the editable mapping

**Considered:** extend the OpenAPI spec with `x-pulumi-*` annotations
so the spec is self-describing.

**Chose:** a separate YAML file owned by this repo, alongside the
spec.

**Why:** keeps `openapi_public.json` pristine and re-pullable from
`pulumi/pulumi-service`. Resource modeling (renames, force-new,
secrets, defaults) is a consumer concern, not a producer concern;
the Pulumi Cloud team shouldn't have to know about provider quirks
to ship spec changes.

### 3.3 Metadata schema is the only extension point

**Considered:** a `customresources/` package as an escape hatch for
resources that don't fit the metadata shape — the v1 pattern,
inverted.

**Chose:** no escape hatch. Every irreducible Cloud-API quirk lands
as a metadata primitive: `bodyOverride`, `iterateOver`,
`rawBodyFrom`/`rawBodyTo`, `bodyAs`, `postCreate`,
`readVia.extractField`, polymorphic scope inference, and so on.

**Why:** an escape hatch becomes a junk drawer — the easy answer
when something's awkward, the place patterns go to die. Forcing each
new pattern through the metadata schema keeps the abstraction honest
and reusable across resources.

### 3.4 Coverage gate

**Decision:** every operationId in the spec must be claimed in
`resource-map.yaml` (resource, function, method, or explicit
exclusion with a one-sentence reason). CI enforces this and so does
runtime `GetSchema`.

**Why:** the only reliable way to prevent silent drift. Spec refresh
PRs that fail the gate force an explicit triage decision rather than
letting new endpoints quietly disappear.

**Cost:** spec refreshes can't ship without classifying every new
operationId. We accept that — see open question on refresh-PR policy
in §5.

### 3.5 Sub-module token namespacing + anti-stutter naming

**Decision:** v2 type tokens use sub-modules that mirror the Cloud
API URL hierarchy (`pulumiservice:orgs/teams:Team`,
`pulumiservice:stacks/tags:Tag`), and don't repeat the module in the
type name.

**Why:** v1's flat `pulumiservice:index:*` namespace doesn't scale
past a few dozen resources without name collisions, and "what module
does this belong to" stops being answerable. Sub-modules mirror the
Cloud API URL so the structure is discoverable. The anti-stutter rule
(no `OidcIssuer` under `orgs/oidc` — just `Issuer`) keeps fully-qualified
tokens readable.

**Cost:** every v1 user's type tokens change. See §4.

### 3.6 Wholesale cut-over, with SDK compat bridge

**Decision:** v2 ships as a clean replacement for v1. There is no
gradual coexistence at the provider level. To soften the code
migration, v2 GA also ships an SDK-only `v1compat` namespace that
re-exports v1-shaped types as thin shims around the v2 resources.

**Considered: gradual layer on top of v1** (per Ian's PR #763
comment). v2 sits as a layer over v1, users migrate piece by piece.
**Rejected** — doubles the maintenance surface; every spec refresh
has to be reconciled against both the v1 hand-coded shapes and the
v2 generated ones, and `resource-map.yaml` stops being the single
source of truth.

**Considered: provider-level type-token aliasing** (per Joe's PR #763
comment, in its strongest form). The provider accepts v1 type tokens
and routes them to v2 resources internally; existing state files
"just work." **Rejected for v2 GA** — adds permanent complexity to
the dispatcher (every resource lookup checks both old and new
tokens), and the aliasing has to live forever. Can be revisited as a
v2.x add-on if state-migration friction proves too high in practice.

**SDK compat bridge** (per Joe's other suggestion) is a thinner
version of the above. It ships a `v1compat` SDK namespace
(e.g., `@pulumi/pulumiservice/v1compat/Team`) where each type is a
shim that constructs the v2 resource. User code keeps compiling
without rewriting imports. **The compat bridge does *not* avoid state
migration** — see §4 — but it does collapse the code-change cost to
near zero for the common case.

> **Open question for confirmation in review:** the design assumes
> the compat bridge is SDK-only (a per-language shim that constructs
> v2 resources under the hood). It does not propose provider-level
> type-token aliasing. If Joe meant the stronger version, that's a
> different architecture and we should discuss before merging.

## 4. State Migration

For users with existing v1 stacks, v2 imposes two migrations:

### 4.1 Code migration — cheap, optional with the compat bridge

Without the bridge, users rewrite imports/type names in their
Pulumi programs.

With the bridge (default v2 GA plan), they import from
`v1compat/...` and existing code keeps compiling. The compat shim is
a deprecation surface — users are expected to migrate to canonical
names within the v2 line. See §5 for the proposed deprecation
window.

### 4.2 State migration — mandatory, recipe-based

Even with the compat bridge, every existing resource's URN type
token must be rewritten in the state file. The compat shim
constructs the new (v2) token under the hood; without state rewrite,
`pulumi up` will plan to **delete every v1 resource and create every
v2 resource** because the URNs don't line up.

The recipe (full version in [`UPGRADE-v1-to-v2.md`](./UPGRADE-v1-to-v2.md)):

1. `pulumi stack export > state.json`
2. **URN rewrite** — replace v1 type tokens with v2 type tokens
   (rename table in the upgrade guide).
3. **ID rewrite** — for the four resources whose ID format changes
   (table below), also rewrite the `id` field in each resource entry.
4. `pulumi stack import < state.json`
5. `pulumi preview` to validate — should show no `delete+create`
   plans.

#### ID-format-changing resources

| Resource          | v1 ID                                | v2 ID                  |
|-------------------|--------------------------------------|------------------------|
| `OrgAccessToken`  | `org/name/tokenId`                   | `org/tokenId`          |
| `TeamAccessToken` | `org/team/name/tokenId`              | `org/team/tokenId`     |
| `StackTags`       | `org/proj/stack/name/tags`           | `org/proj/stack/name`  |
| `ApprovalRule`    | `environment/org/proj/env/ruleId`    | `org/gateId`           |

These four are why `pulumi state rename --type-token` alone isn't
enough. The recipe needs an ID-rewrite step.

### 4.3 Risks

- **Scale.** Users with 500+ resources need the recipe to be
  scriptable and idempotent. `sed` works for type tokens but is
  fragile for the structured ID rewrites. We probably need a small
  helper tool — see §5.
- **Validation.** No dry-run mode today. The user's only feedback
  loop is `pulumi preview` after `stack import`, and a wrong rewrite
  surfaces as a `delete+create` plan, which is alarming.
- **Edge cases.** The recipe hasn't been validated against very
  large state files (deep nesting, many parents/dependencies).
  Worth a stress test against a synthetic 500-resource stack before
  GA.

## 5. Open Questions

- **Confirm compat-bridge interpretation.** §3.6 reads Joe's
  suggestion as an SDK-only shim. If he meant the stronger
  provider-level aliasing form (also rejected in §3.6), state
  migration changes shape — re-litigate before merging.
- **Migration tooling.** Ship a `pulumi-pulumiservice-migrate`
  helper in this repo, or rely on docs + `sed`? Lean toward
  shipping a script — gives a tested path, an issue tracker for
  bugs, and a place to add a dry-run mode.
- **Coverage-gate policy on net-new endpoints.** When a spec refresh
  introduces a new operationId, the gate fails until it's claimed.
  Default policy proposal: refresh PRs that fail the gate are
  triage-only (add resources/functions/methods/exclusions, no other
  changes), and resource-modeling work follows in separate PRs.
  Otherwise refreshes get blocked on resource design.
- **Compat bridge deprecation window.** How many minor versions does
  `v1compat` stick around? Proposal: deprecation warnings start
  v2.1; removal in v3. Want explicit sign-off before committing.
- **Outstanding implementation issues** (architecture-adjacent, not
  architecture; listed for completeness — these block GA but
  don't change the design):
  - Call/Invoke RPCs not wired (`Account.triggerScan`,
    `listAccounts`).
  - OIDC Issuer `policies` `sortOnRead: true` — re-introduces the
    spurious-diff bug fixed in #542.
  - Access-token `value` dropped from state on first refresh —
    list endpoints don't return `tokenValue`, and the metadata
    schema has no preserve-from-state primitive.
  - SDK regen lagging the latest `resource-map.yaml` (HEAD is a
    "WIP checkpoint").
