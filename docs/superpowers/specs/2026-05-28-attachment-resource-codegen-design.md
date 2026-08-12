# Attachment Resource Codegen — Design

**Date:** 2026-05-28
**Status:** Draft (pending review)
**Author:** lward

## Problem

The Pulumi Service Provider exposes two `PolicyGroup` implementations:

- **`index:PolicyGroup`** — hand-written (`provider/pkg/resources/policy_group.go`). Takes
  declarative `stacks` / `accounts` / `policyPacks` arrays. In `Create` it issues
  `CreatePolicyGroup` (empty shell) → `BatchUpdatePolicyGroup([...all adds...])` → re-read.
  It bridges the imperative membership API into a declarative resource by diffing the desired
  lists and emitting batched add/remove calls.
- **`api:PolicyGroup`** — generated from `provider/pkg/cloud/spec.json` via the
  `dispatch.Wrap` + `rest.Resource` machinery. Each CRUD verb maps 1:1 to a single OpenAPI
  operation. Its create op `NewPolicyGroup` only accepts `name`/`entityType`/`mode`/`agentPoolId`,
  so the generated resource **cannot manage membership at all**.

The membership API is **imperative single-edge**: `addStack` / `removeStack` (and the account /
policy-pack equivalents) are *fields in the body* of one `UpdatePolicyGroup` PATCH operation —
not distinct operations. Generic codegen maps operations 1:1, so it cannot synthesize the
list-diffing the hand-written resource does.

## Goal

Let the generated (`api`) resources manage membership **via code generation**, so that the
spec-refresh pipeline keeps producing the membership surface for free, instead of hand-writing
and hand-maintaining resources like `index:PolicyGroup`.

## Non-goals

- Declarative membership *arrays* on the parent resource (that is what `index:PolicyGroup`
  already does and requires bespoke list-diffing — explicitly rejected).
- Performing mutations via Pulumi **functions/invokes** (functions are read-only by design;
  they run on every preview and have no lifecycle — wrong tool for attach/detach).
- Retiring `index:PolicyGroup`. Deprecation is a separate, later decision.

## Key insight

An imperative single-edge API (`addX` / `removeX`) maps perfectly onto an **attachment
resource** (one membership edge = one resource), AWS-`RolePolicyAttachment`-style:

- **Create** → `addX`
- **Delete** → `removeX`
- **Read** → GET the parent, check whether this edge is present in the parent's membership list
- **Update** → none (all inputs `forceNew` / replace-on-change)

Crucially, an attachment needs **no list-diffing** — the hard part of composite-create — which
is exactly why it is the cleanest thing to generate.

## Decision summary

- **Approach:** Attachment resources (Approach A), produced by code generation (not hand-written).
- **Scope:** Build the mechanism to be **general across the whole spec** — a first-class
  `attachment` resource *kind* in the runtime plus a spec-wide detector in the scaffolder.
- **Auto-emit is conservative; curation covers the messy cases.** The detector only
  auto-emits attachments it can derive unambiguously; asymmetric or identity-ambiguous cases
  are declared via hand-curated metadata the scaffolder preserves.

## Where the work is

| Layer | File(s) | Change | Size |
|---|---|---|---|
| Metadata schema | `provider/pkg/rest/metadata.go` | Add `Attachment *AttachmentMeta` to `ResourceMeta` | small |
| Runtime | `provider/pkg/rest/resource.go` | Attachment branch in Create/Delete/Read; body nesting; membership read; id handling | **medium (load-bearing)** |
| Schema gen | `provider/pkg/rest/schema.go` | Emit membership inputs instead of the raw update-request body shape | small–medium |
| Scaffolder | `provider/tools/scaffold-metadata/main.go` | `detectAttachmentPairs` + emit pass (~`main.go:223`), reusing `flattenedProps` + `mergeOperations` | small |
| Tests | `provider/pkg/cloud/*_test.go`, `examples/` | Idempotency stays green; example + coverage per new token | medium |

The scaffolder already introspects request bodies (`flattenedProps(spec, ref)` flattens `allOf`
and is already used for other inferences), so detection is additive. The runtime is the real
investment: today every metadata entry runs the **same** `rest.Resource` path with two
assumptions attachments break — (1) one operation per verb and `buildRequestBody`
(`resource.go:870`) emits *every* matching top-level field (so it would send `addStack` **and**
`removeStack` together), and (2) Read assumes the GET returns the resource object directly
(`resource.go:768`) with no membership-list support. There is no resource "kind" discriminator
today; this introduces the first one.

## Detailed design

### 1. `AttachmentMeta` descriptor (`rest/metadata.go`)

```go
type ResourceMeta struct {
    // ...existing fields...
    Attachment *AttachmentMeta `json:"attachment,omitempty"` // when set, this entry is an attachment kind
}

type AttachmentMeta struct {
    MutationOp       string   `json:"mutationOp"`       // shared op for create+delete, e.g. "UpdatePolicyGroup"
    AddField         string   `json:"addField"`         // body field for Create, e.g. "addStack"
    RemoveField      string   `json:"removeField"`      // body field for Delete, e.g. "removeStack"
    ReadOp           string   `json:"readOp"`           // parent GET, e.g. "GetPolicyGroup"
    MembershipField  string   `json:"membershipField"`  // list field in ReadOp response, e.g. "stacks"
    MatchKey         []string `json:"matchKey"`         // fields identifying the edge within MembershipField
    ParentIDParams   []string `json:"parentIdParams"`   // path params identifying the parent, e.g. ["organizationName","policyGroup"]
}
```

When `Attachment != nil`, the runtime takes the attachment branch (below) instead of the
default 1:1 CRUD path. `Operations` may be left empty for attachments (the ops live in
`AttachmentMeta`), or we mirror `mutationOp` into `Operations.Create`/`Delete` for tooling that
expects them — to be decided in planning.

### 2. Detection algorithm (scaffolder)

A new pass, run after the candidate loop (~`main.go:223`), over every operation that is an
update (PATCH/PUT) on a known parent candidate:

1. `props := flattenedProps(spec, op.RequestRef)`.
2. For each property `addX`, look for a sibling `removeX` (same stem after the `add`/`remove`
   prefix).
3. **Match the read list by TYPE, not name.** Find the field in the parent read-op response
   whose element `$ref` equals `addX`'s `$ref`. (Names are unreliable: `addPolicyPack` →
   `appliedPolicyPacks`, not `policyPacks`.)
4. **Auto-emit only when the edge identity is unambiguous** — i.e. the add-field's type, the
   remove-field's type, and the membership element type are the **same** `$ref`, and that
   type's required fields *are* the identity (no extra non-identity payload). Today that is
   `addStack`/`removeStack` ↔ `stacks` (`AppPulumiStackReference` = `{name, routingProject}`,
   both required, nothing else).
5. Anything that fails the unambiguity test (asymmetric types, asymmetric field names,
   identity is a subset of a heavy object) is **not** auto-emitted. It is `log()`'d as a
   "candidate attachment requiring curation" so it is visible, never silently dropped.
6. Emit each auto-detected attachment through `mergeOperations` so hand-curation is preserved
   and the idempotency test (`TestScaffoldMetadataIdempotent`) stays byte-stable.

### 3. Curation for asymmetric cases

Asymmetric attachments are authored by hand as an `attachment` block on a metadata entry; the
scaffolder's write-if-absent merge leaves them untouched. This covers:

- **PolicyGroup insights accounts** — `addInsightsAccount`/`removeInsightsAccount` take a full
  `InsightsAccount` object in the spec (required `id`/`ownedBy`/`provider`/…), but the response
  `accounts` is `[]string`. Curated descriptor: input is the account **name** (string),
  `membershipField: "accounts"`, `matchKey: ["name"]`, with the body shaped as the reference
  the API actually accepts.
- **PolicyGroup policy packs** — symmetric by type but the identity is a subset of a heavy
  object (`config`, `environments`, `version`). Needs a curated `matchKey` (e.g.
  `["name","version"]`); applying a pack also carries config, so this is a config-bearing edge,
  not a pure membership edge. Treat with care or defer.
- **Team permissions** — `addStackPermission`/`removeStack` and
  `addEnvironmentPermission`/`removeEnvironment` have asymmetric field names *and* types; fully
  curated if pursued (note: `index:`-side already has `TeamStackPermission` /
  `TeamEnvironmentPermission` resources — overlap to reconcile before generating these).

### 4. Runtime semantics (`rest/resource.go`)

- **Create:** body = `{ <AddField>: <inputs-as-object> }`, PATCH the parent (`MutationOp`).
  Requires new body construction that **nests** the resource inputs under `AddField` —
  `buildRequestBody` today only copies matching top-level fields and cannot nest.
- **Delete:** body = `{ <RemoveField>: <inputs-as-object> }`, PATCH the parent. 404 → success
  (already handled).
- **Read:** GET parent (`ReadOp`); locate the element in `response[MembershipField]` whose
  `MatchKey` fields equal this edge's inputs. Present → state; **absent → treat as gone**
  (synthesize the 404-equivalent so refresh removes/recreates). This membership check is
  net-new in `fetchState`/Read.
- **Diff / Update:** no update op; all inputs `forceNew` (replace-on-change). Reuses existing
  `replaceTriggeringFields`.
- **ID / import:** composite `idFormat` from `ParentIDParams` + `MatchKey`, e.g.
  `{organizationName}/{policyGroup}/{routingProject}/{name}` for stacks,
  `{organizationName}/{policyGroup}/{name}` for accounts. Note object-valued keys do **not**
  round-trip through the current scalar `idFormat` (`propertyValueToString`,
  `compileIDFormatRegex`) — composite/flattened keys are required.

### 5. Schema generation (`rest/schema.go`)

The generated attachment resource must expose **membership inputs** — the `AddField` object's
properties (flattened) plus `ParentIDParams` — rather than the raw `UpdatePolicyGroupRequest`
body shape. E.g. `PolicyGroupStackAttachment` inputs: `organizationName`, `policyGroup`,
`name`, `routingProject`. A schema-gen branch keys off `Attachment != nil`.

### 6. Token / naming

`pulumiservice:<module>:<ParentNoun><EdgeStem>Attachment`, e.g. `PolicyGroupStackAttachment`,
`PolicyGroupAccountAttachment`. Module placement follows the parent generated resource's module
(confirm `api` vs `index` surfacing during planning). The parent (`PolicyGroup`) entry is never
mutated by the attachment pass — it stays a normal, membership-free generated resource.

## Inventory (today's spec)

| Edge | Disposition |
|---|---|
| `PolicyGroup` `addStack`/`removeStack` ↔ `stacks` | ✅ auto-emit (clean symmetric, identity = whole object) |
| `PolicyGroup` `addInsightsAccount`/`removeInsightsAccount` ↔ `accounts` | ⚠️ curate (object-in / string-out; spec over-specifies) |
| `PolicyGroup` `addPolicyPack`/`removePolicyPack` ↔ `appliedPolicyPacks` | ⚠️ curate (heavy object; identity is a subset; config-bearing) |
| `Team` `addStackPermission`/`removeStack`, `addEnvironmentPermission`/`removeEnvironment` | ⚠️ curate if pursued (asymmetric; overlaps existing `index:` resources) |
| `AgentEntityDiff` `add`/`remove` | ⛔ skip (not wired to any endpoint) |

So the fully-automatic win **today** is `PolicyGroupStackAttachment`; the framework + curation
path delivers the rest and any future symmetric pairs the spec grows.

## Risks & edge cases

- **Insights-account child auto-expansion.** Server-side, adding a parent account expands to
  child accounts (the legacy resource has special "don't remove children whose parent is
  present" logic). An edge-per-resource model must tolerate server-added extras in `accounts`
  on Read and must not remove children when a parent edge is deleted. This is why accounts are
  curated and likely phase-2.
- **Spec-vs-reality type mismatch.** `addInsightsAccount`'s spec type (full `InsightsAccount`)
  ≠ what the API needs (a reference) ≠ the response type (`string`). Codegen must not trust the
  spec type blindly here; curation pins the real shape.
- **Object-valued ID keys** don't round-trip through the scalar `idFormat` machinery — needs
  composite/flattened handling.
- **Idempotency + coverage tests.** `TestScaffoldMetadataIdempotent` requires regenerate +
  commit to be byte-stable; `TestEveryApiResourceHasExample` requires a YAML example (or
  waiver) per new token.
- **First-ever resource "kind".** Introduces a discriminator into a runtime that treats all
  entries identically — keep the attachment branch well-isolated and the default path unchanged.

## Testing strategy

- Unit: scaffolder detector (symmetric auto-emit; asymmetric → curation log, no emit);
  metadata round-trip; `buildRequestBody` nesting; membership Read present/absent.
- Idempotency: regenerate `metadata.json`, assert byte-stable.
- Integration: YAML example for `PolicyGroupStackAttachment` exercising
  up → preview → refresh (membership drift) → destroy; import via composite id.
- Coverage: satisfy `TestEveryApiResourceHasExample` for each emitted token.

## Suggested sequencing (within the general build)

1. `AttachmentMeta` + runtime attachment branch + schema-gen, validated end-to-end on a
   hand-written `PolicyGroupStackAttachment` metadata entry (proves the runtime).
2. Scaffolder `detectAttachmentPairs` + emit pass; confirm it auto-emits exactly
   `PolicyGroupStackAttachment` and logs the asymmetric candidates.
3. Curated `PolicyGroupAccountAttachment` (the originally-motivating case), with child
   auto-expansion handling.
4. Policy packs / Team only if/when we decide to reconcile with existing resources.

## Open questions for review

1. **Module/token surfacing** for attachments (`api` vs `index`) — match parent? Confirm the
   parent generated `PolicyGroup`'s public token.
2. **Policy packs**: model as an attachment at all, given they carry config and overlap the new
   `PolicyPack` resource? Or leave out of the membership mechanism entirely.
3. **Team permissions**: in scope, or explicitly out (kept on the hand-written `index:` side)?
4. **`Operations` mirroring**: leave `operations` empty for attachments, or mirror `mutationOp`
   into create/delete for tooling that reads `Operations`?
