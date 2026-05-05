# Pulumi Cloud Provider v2 — Architecture

**Status:** Reference (current state on `lward/pcp`).
**Date:** 2026-05-05.
**Scope:** the v2 layer only — REST library, scaffolder, OpenAPI flow, `metadata.json` conventions. The legacy raw and outer `infer` layers appear briefly in §1 for context.

## 1. Context

### Goals

- New Pulumi Cloud resources can be added by editing `metadata.json` against the embedded OpenAPI spec — no Go code per resource.
- The OpenAPI spec is the source of truth for shapes; the spec file refreshes from upstream via `go generate` and is checked for staleness in CI.
- Existing user code continues to resolve unchanged. v2 lives under a sub-namespace (`pulumiservice:v2:*`) inside the same plugin and schema-name (`pulumiservice`) as v1.
- Schema construction and CRUD dispatch are runtime-only. Nothing in the v2 layer is generated at build time.

### Non-goals

- Replacing the v1 surface. v1 stays implicit at the package root (`pulumiservice:index:*`); v2 is additive.
- A typed Go API per v2 resource. v2 resource I/O is `property.Map` end-to-end.
- Polymorphic discriminated unions (e.g. RBAC permission shapes). The v2 schema builder degrades unsupported OpenAPI constructs to `pulumi.json#/Any`; resources that need first-class polymorphism stay on the outer `infer` layer.

### Where v2 sits

The plugin (`pulumi-resource-pulumiservice`) is a three-layer composition built in `provider/pkg/provider/provider.go:100`:

1. **Innermost — `legacyRaw`.** A `pulumirpc.ResourceProviderServer` for the resources defined in `manual-schema.json`. Tokens at `pulumiservice:index:*`.
2. **Middle — `dispatch.Wrap(legacyRaw, ...)`.** Routes per-URN CRUD calls for v2 tokens to `rest.DynamicResource` handlers built from `cloud.Spec()` and `cloud.Metadata()`. The v2 schema fragment is spliced into `GetSchema` responses by `withCloudV2Schema` (`provider/pkg/provider/provider.go:223`).
3. **Outermost — `infer.NewProviderBuilder().WithWrapped(...)`.** Adds modern `infer`-style resources (Team, OrganizationRole, etc.) at `pulumiservice:index:*` and stamps in package-level metadata (display name, language map, config schema).

The rest of this document concerns layer 2.

### Token shape

Two forms appear:

- **Bare:** `pulumiservice:v2:Foo` — used when the scaffolder didn't assign a module (the underlying URL prefix is a singleton).
- **Modular:** `pulumiservice:v2/<module>:Type` — emitted when at least two candidates share the same first-or-aliased URL segment. Example: `pulumiservice:v2/agents:Pool` for `AgentPool` derived from `/orgs/{org}/agent-pools`. As of this writing, 39 of 49 v2 resources carry an explicit module.

The modular form is stored as the `token` field on the metadata entry; the bare form is the metadata-key. The runtime uses `token` if present and falls back to the key otherwise (`provider/pkg/rest/resource.go:46`).

## 2. The two inputs

### `provider/pkg/cloud/spec.json`

- An OpenAPI 3 document downloaded from `https://api.pulumi.com/api/openapi/pulumi-spec.json`. Embedded into the binary via `//go:embed`. Parsed once at process start (`provider/pkg/cloud/spec.go:55`).
- The spec is the source of truth for operation paths, request/response schemas, and parameters.
- Purely machine-managed. No hand edits. Refresh via `go generate ./provider/pkg/cloud/...` (which runs `openapi-fetch` followed by `scaffold-metadata`).

The runtime `Spec` (`provider/pkg/rest/spec.go`) only models the subset BuildSchema and the dispatcher need: an operation index keyed by `operationId`, plus `components.schemas` for `$ref` resolution. Duplicate `operationId`s fail parsing.

### `provider/pkg/cloud/metadata.json`

The bridge between OpenAPI operations and Pulumi resources. File format is versioned (`version: 1`) and parsed by `rest.ParseMetadata` (`provider/pkg/rest/metadata.go:110`).

Top-level fields:

| Field        | Owner       | Purpose                                                                  |
| ------------ | ----------- | ------------------------------------------------------------------------ |
| `version`    | tooling     | Schema version. Currently `1`.                                           |
| `package`    | tooling     | Default package for tokens that don't carry a prefix. Set to `pulumiservice`. |
| `_note`      | human       | Free-form comment. Round-tripped, not consumed.                          |
| `_excluded`  | human       | Tokens the scaffolder must skip on regen. Empty today.                   |
| `resources`  | mixed       | Map of token → `ResourceMeta`.                                           |

Per-resource (`ResourceMeta`) fields:

| Field            | Owner       | Purpose                                                                                       |
| ---------------- | ----------- | --------------------------------------------------------------------------------------------- |
| `operations`     | tooling     | `{create, read, update, delete}` operation IDs. Replaced wholesale on regen.                  |
| `token`          | tooling     | User-facing Pulumi token (modular form). Written once; humans can override and the override survives. |
| `renames`        | tooling     | Pulumi-side → wire-side name map. Inferred by the scaffolder; humans may add entries.         |
| `outputsExclude` | tooling     | Response fields to drop (e.g. envelope keys that collide with the resource's own type name).  |
| `aliases`        | human       | Pulumi tokens treated as equivalent (used for in-place migration after renames).              |
| `fields`         | human       | Per-field overrides: `forceNew`, `secret`, `description`. Keyed by Pulumi-side name.          |
| `outputs`        | human       | Output allowlist (mutually exclusive with `outputsExclude`; allowlist wins).                  |
| `description`    | human       | Resource description override; falls back to the create op's description.                     |
| `examples`       | human       | PCL snippets rendered into the description as `## Example Usage` blocks.                      |

The dual ownership is preserved by the scaffolder via `json.RawMessage` round-tripping: the entry is decoded into a `map[string]any`, the auto-derived keys are rewritten, and the result is re-encoded with deterministic key ordering (`provider/tools/scaffold-metadata/main.go:514`). Hand-curated keys pass through untouched.

## 3. The codegen tools

Both tools live under `provider/tools/` and are wired into the `cloud` package via `//go:generate` directives in `provider/pkg/cloud/spec.go:41`:

```
go:generate go run ../../tools/openapi-fetch -out spec.json
go:generate go run ../../tools/scaffold-metadata -in spec.json -out metadata.json
```

### `openapi-fetch`

Downloads the spec, normalizes it (decoded then re-encoded with sorted keys, 2-space indentation, trailing newline), writes to disk. The normalization step exists so diffs against the committed spec are reviewable. URL is configurable via `-url` flag or `PULUMI_CLOUD_OPENAPI_URL`; default is the public spec URL.

### `scaffold-metadata`

Walks every operation in the spec, derives a candidate set of v2 resources, and merges them into `metadata.json`. Six steps in order:

**1. Verb/noun extraction.** Each `operationId` is split into `(verb, noun, slot)` where `slot` is one of `create | read | update | delete | ""`. The `verbPrefixes` table is matched longest-first so `BatchCreate` wins over `Create`; `matchPrefix` requires the next character to be uppercase so `Update` matches `UpdateStack` but not `Updater`. Non-CRUD verbs (`List`, `Cancel`, `Approve`, etc.) are recognized so noun extraction works on action ops, but they don't claim a slot.

**2. Slot grouping.** Operations group by extracted noun. Within a noun, the first operationId for each slot wins. Operations whose path carries the `x-pulumi-route-property.Visibility = "Deprecated"` marker are skipped.

**3. Scope folding.** Three passes that collapse name variants onto the canonical noun:

- *Underscore folding* — when a longer noun has the form `Shorter_…` and `Shorter` exists without its own create op, the longer noun's slots feed the shorter and the longer entry is removed. (No live spec entries currently trigger this; route-suffixed nouns like `Environment_esc_environments` survive because no bare `Environment` exists.)
- *Plural folding* — e.g. `Tasks` → `Task`: the plural's non-read slots feed the singular. Real example: `CreateTasks` creates the `Task` resource.
- *Scope-prefix stripping* — `OrgAgentPool` → `AgentPool`, `PulumiTeam` → `Team`. Strips a leading `Organization`, `Org`, `Pulumi`, `Team`, `Project`, `Stack`, or `User` prefix when the bare noun exists without its own create op.

**4. Rename inference.** For each path parameter on a non-create op that isn't already on the create op's path, the scaffolder applies four rules in order (`provider/tools/scaffold-metadata/main.go:407`):

1. **Server-id pattern** — create response has `id`, path uses a noun-prefixed form (e.g. `issuerId`): emit `{path-param: "id"}`.
2. **Suffix-strip pattern** — body field mirrors path under suffix removal (`{tagName}` ↔ body `name`): emit `{body-field: path-param}`.
3. **Catch-all** — bare path param matched by body `name`.
4. **Verbose body alias** — body field duplicates a path param under a longer alias (e.g. `organizationName` body ↔ `orgName` path): detected via suffix match plus stem prefix relationship.

**5. Module derivation.** Each candidate's URL prefix (after stripping `api`, `orgs`, `user`, `console`) is canonicalized via the `moduleAliases` table (e.g. `agent-pools` → `agents`, `oidc/issuers` → `auth`). When at least two candidates share a canonical prefix, both get an explicit `token` of the form `pulumiservice:v2/<module>:<Type>`. The type name is shortened by stripping a redundant module prefix (`AgentPool` in module `agents` → `Pool`) and a trailing route suffix (`Environment_esc_environments` in module `esc` → `Environment`).

**6. Merge + report.** Each candidate's auto-derived fields (`operations`, `renames`, `outputsExclude`, `token`) are written; hand-curated fields round-trip; orphans (entries in `metadata.json` no longer derivable from the spec) are reported on stderr. The output is written with deterministic key order (preferred fields first, alphabetical for the rest).

## 4. Runtime: schema + dispatch

Two entry points, both reading the same parsed spec and metadata:

- `rest.BuildSchema(spec, metadata, pkg)` — produces a `schema.PackageSpec`.
- `rest.Resources(spec, metadata)` — produces `map[token] *DynamicResource` for the dispatcher.

### `rest.BuildSchema`

Builds one `ResourceSpec` per metadata entry (`provider/pkg/rest/schema.go:62`):

- **Inputs** come from the create op: path/query parameters first, then the request body's flattened object schema. Path parameters default to `WillReplaceOnChanges = true` and are required.
- **Outputs** come from the read op's response, falling back to create's response if read has none. The `id` key is reserved by Pulumi and skipped from outputs.
- **Path params merged into outputs.** This is non-obvious. The merge exists so Delete (which sources path params from saved state, not inputs) can reconstruct its URL after a process restart.
- **Allow/deny.** `outputs` (allowlist) takes precedence; `outputsExclude` (denylist) is applied otherwise.
- **Secret heuristic.** `looksSecret` flags any field whose lowercased name contains `secret`, `tokenvalue`, `password`, `apikey`, `accesstoken`, or `ciphertext`. The OpenAPI spec doesn't carry an extension we can rely on for this today; the heuristic catches webhook signing secrets, the per-token `tokenValue`, and similar fields.
- **Examples.** Each PCL snippet in `examples` is appended to the description as a `pulumi`-tagged fenced block (`appendExamples`). Pulumi's SDK codegen runs `pulumi convert` per target language at gen time, producing per-language examples in TypeScript/Python/Go/.NET/Java docstrings.
- **Aliases.** Any `aliases` from metadata become schema-level `AliasSpec` entries on the resource.

Errors aggregate per-resource and surface as a single multi-line error from `BuildSchema` — never from process startup.

### `rest.DynamicResource`

Implements `mw.CustomResource` (`provider/pkg/rest/resource.go:58`). Operation IDs are resolved against the spec at call time, not at construction; broken operations surface on the first CRUD call to the affected resource.

- **Check** is a passthrough. `replaceOnChanges` tags drive engine-side replacement.
- **Diff** is structural inputs equality. Anything else returns `HasChanges: true`.
- **Create** substitutes path params into the create op's URL, JSON-encodes the inputs as the body, fires the request, decodes the response, populates path params into state, and synthesizes the resource ID.
- **Read** is optional. When the resource has no read op (token endpoints, tag list-only endpoints), Read returns prior state unchanged. Otherwise, `inputs` and prior `state` are merged so server-generated path params (e.g. `issuerId`) are reachable when constructing the URL.
- **Update** merges `inputs ∪ OldInputs ∪ State`, fires the update op, populates path params back into state.
- **Delete** merges `state ∪ OldInputs` so legacy stacks (created before path params round-tripped into outputs) can still be deleted.

Three sub-mechanisms worth naming:

**`synthesizeID`.** The resource ID is concatenated from path-parameter values of the most authoritative non-create op (read → update → delete → create), looked up in state first then inputs. After Create returns, the engine owns the ID and threads it back as `req.ID` on subsequent calls; nothing in DynamicResource re-synthesizes it.

**`populatePathParams`.** Many Pulumi Cloud endpoints return empty or minimal bodies (e.g. POST creating a stack returns `{}`). Without enrichment, a downstream resource referencing `${parent.projectName}` would fail because the parent's state never carried `projectName` forward, even though the schema declares it as an output. `populatePathParams` walks the path-param placeholders of create + read and copies each Pulumi-named input value into state when state doesn't already carry that key.

**`renameMapKeys`.** The runtime's response decoder translates wire-side keys to Pulumi-side names so that downstream `state.GetOk(pulumiName)` lookups work uniformly. Renames apply only at the top level of resource I/O, not nested.

### Schema delivery — `withCloudV2Schema`

The v2 schema is computed and merged at GetSchema time, not at startup. `withCloudV2Schema` (`provider/pkg/provider/provider.go:223`) wraps the underlying provider's `GetSchema`, parses the base schema returned by the inner layer, calls `rest.BuildSchema`, and merges the result via `mergeSpec` (existing tokens win on collision; `dst.Resources[k]` is set only if absent). The composed JSON is re-encoded and returned.

The implication: BuildSchema errors are surfaced through GetSchema responses. The provider boots and serves CRUD even if the metadata document has resources whose operation IDs don't resolve — those resources fail their first call rather than blocking the whole binary.

### Transport

`authedTransport` (`provider/pkg/provider/provider.go:198`) rewrites the scheme and host of every outgoing request onto the configured Pulumi Cloud base URL, then attaches `Authorization: token <PAT>`, `Accept: application/vnd.pulumi+8` (when not already set), and `X-Pulumi-Source: provider`. The DynamicResource layer constructs URLs against a sentinel host (`https://transport.invalid`) when the spec declares no servers; the transport overrides the host before sending. Resource code sets its own `Accept: application/json` for v2 calls, which beats the transport's vendored-media-type default.

Transport resolution is a process-global hook: `rest.SetTransportResolver` registers a `func(ctx) (Transport, error)`, which the dispatcher invokes via `resolveTransport(ctx)` on every CRUD call. The provider's `Configure` step is the natural place to register it. This keeps `rest` decoupled from the provider's `config` package.

## 5. Invariants worth pinning

These are facts about v2 that don't fall out of any single file and that future work needs to preserve.

1. **Everything is runtime.** The schema is recomputed on every GetSchema; CRUD reads `Spec` and `Metadata` directly. There is no `schema.json` artifact for v2.
2. **Dual ownership of `metadata.json`.** The scaffolder rewrites a known set of fields (`operations`, `token`, `renames`, `outputsExclude`) and round-trips the rest via `json.RawMessage`. Adding a new auto-derived field requires updating both the scaffolder and `mergeOperations`'s key list.
3. **Resource IDs are synthesized on Create only.** They're concatenated from path-param values (state ∪ inputs). After Create the engine owns the ID and threads it back on Read/Delete; DynamicResource never re-derives it.
4. **Renames map Pulumi-side → wire-side.** Response keys are translated to Pulumi-side via `renameMapKeys` before returning state, so every downstream `state.GetOk(...)` call uses the Pulumi name. Adding a rename requires both directions to keep working — verify the round-trip in tests.
5. **Path parameters round-trip through state.** The schema builder explicitly merges path params into outputs so Delete can reconstruct its URL from saved state. A future change that drops outputs-side path params will break Delete on the next process restart.
6. **`looksSecret` is a name-substring heuristic.** The OpenAPI spec does not carry a reliable secret marker today. New secret-shaped fields with non-matching names need a `fields[name].secret = true` override in `metadata.json` until the spec gains an `x-pulumi-secret` extension.
7. **Read is optional.** Resources without a per-instance GET endpoint (tokens, tags, memberships) have an empty `read` slot; refresh is a no-op and prior state is returned unchanged.
8. **Schema mapping errors are deferred.** They surface during GetSchema, not at startup. Half-broken metadata documents don't block the running provider; the affected resources fail at first use.
9. **Pulumi reserves `id`.** The schema builder skips any response field literally named `id` from outputs. Resources whose API returns server-generated identifiers under that name expose them through path-parameter renames (see rule 1 of `inferRenames`).
10. **The scaffolder excludes two classes.** `x-pulumi-route-property.Visibility = "Deprecated"` operations and any token in the top-level `_excluded` array. Both filters are silent; check the scaffolder's stderr summary if a derivation result is unexpected.
