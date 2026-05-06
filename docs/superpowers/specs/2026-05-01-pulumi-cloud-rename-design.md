# Pulumi Cloud Provider — rename + REST-binding scaffolding

**Status:** Draft, awaiting review.
**Date:** 2026-05-01
**Branch:** `lward/pcp`

## Summary

Rename `pulumi-pulumiservice` → `pulumi-cloud` in place. Introduce a small REST-binding library that lets us declare new resources by referencing OpenAPI operation IDs plus hand-written `Args`/`State` Go types. Existing resources remain on whatever framework they currently use (legacy custom gRPC interface or `pulumi-go-provider/infer`); both already coexist via `MakeProvider`'s mux pattern.

## Goals

1. Repo name, plugin name, Go module path, schema name, and SDK package names all reflect "Pulumi Cloud."
2. Existing user stacks upgrade in place via type aliases (`cloud:index:Foo` ⇄ `pulumiservice:index:Foo`) — no resource churn.
3. New resources can be authored against `pulumi-go-provider/infer` with an HTTP backend driven by the Pulumi Cloud OpenAPI spec, without per-resource CRUD plumbing.
4. CI continues to release the (now-renamed) plugin from this repo. Legacy `pulumi-pulumiservice` versions stay frozen on registries; no further releases under that name from this repo.

## Non-goals

- Maintaining a parallel v1 codebase or dual-named releases.
- Migrating any existing resources to the new REST-binding library. Both styles coexist; existing code is touched only by the rename and alias additions.
- Designing or implementing the actual generated Cloud resources (deferred to a follow-up project).
- Designing handling for deeply polymorphic OpenAPI shapes (RBAC permission discriminators, etc.) — deferred.
- A user-facing helper SDK for predefined shapes — out of scope.

## Rename map

| Surface | Before | After |
|---|---|---|
| Repo | `pulumi/pulumi-pulumiservice` | `pulumi/pulumi-cloud` |
| Plugin binary | `pulumi-resource-pulumiservice` | `pulumi-resource-cloud` |
| Go module | `github.com/pulumi/pulumi-pulumiservice` | `github.com/pulumi/pulumi-cloud` |
| `cmd/` directory | `provider/cmd/pulumi-resource-pulumiservice/` | `provider/cmd/pulumi-resource-cloud/` |
| Schema `name` | `pulumiservice` | `cloud` |
| Schema `displayName` | `Pulumi Cloud` (already correct) | `Pulumi Cloud` |
| Resource type tokens | `pulumiservice:index:*` | `cloud:index:*` |
| npm | `@pulumi/pulumiservice` | `@pulumi/cloud` |
| PyPI | `pulumi_pulumiservice` | `pulumi_cloud` |
| Go SDK module | `…/pulumi-pulumiservice/sdk/go/pulumiservice` | `…/pulumi-cloud/sdk/go/cloud` |
| .NET package | `Pulumi.PulumiService` | `Pulumi.Cloud` |
| Java package | `com.pulumi.pulumiservice` | `com.pulumi.cloud` |
| ci-mgmt | `provider: pulumiservice` | `provider: cloud` |

Source-internal Go identifiers (`PulumiServiceResource`, `pulumiserviceProvider`, etc.) and the test org name (`service-provider-test-org`) are intentionally left as-is — they have no external surface.

## Aliases

Every resource currently in `manual-schema.json` (and every `infer.Resource` registration) gets a `pulumiservice:index:Foo` alias. The Pulumi engine resolves aliases at refresh; the provider receives all incoming RPCs under the new tokens, so no runtime dual-token registry is needed.

Schema-level alias entries:

```json
"aliases": [
  { "type": "pulumiservice:index:Stack" }
]
```

For `infer`-driven resources the alias is added via the resource's `Annotate(infer.Annotator)` method.

## Coexistence runtime

**Already in place.** `MakeProvider` in `provider/pkg/provider/provider.go` builds an `infer.NewProviderBuilder()` provider, wraps the legacy custom `pulumirpc.ResourceProviderServer` via `WithWrapped(rpc.Provider(...))`, and serves the union as a single muxed provider.

This design adds a new source of `infer.InferredResource` entries — the REST-binding library's outputs — and chains them into the existing `WithResources(...)` call. No mux redesign needed.

Disjointness invariant: every resource token must be unique across the legacy schema, the existing infer-driven resources, and the new REST-driven resources. `pulumi-go-provider`'s build step enforces this at provider startup.

## REST-binding library

### Shape

A small library at `provider/pkg/rest/`. Exposes a generic `Controller[A, S]` that implements `infer.Custom*[A, S]` interfaces and translates each call into an HTTP request against an operation in a parsed OpenAPI spec.

Each new Cloud resource is a tiny declaration in `provider/pkg/cloud/`:

```go
type Stack struct {
    *rest.Controller[StackArgs, StackState]
}

type StackArgs struct {
    OrganizationName string `pulumi:"organizationName" provider:"replaceOnChanges"`
    ProjectName      string `pulumi:"projectName"      provider:"replaceOnChanges"`
    StackName        string `pulumi:"stackName"        provider:"replaceOnChanges"`
}

type StackState struct {
    StackArgs
    StackID string `pulumi:"stackId"`
}

func newStack() *Stack {
    return &Stack{
        Controller: rest.New[StackArgs, StackState](spec, rest.Ops{
            Read:   "getStack",
            Create: "createStack",
            Update: "updateStack",
            Delete: "deleteStack",
        }),
    }
}
```

Promoted methods from the embedded `*rest.Controller` satisfy infer's reflection. The struct's *type name* (`Stack`) becomes the resource token name (`cloud:index:Stack`).

### What the library handles

- Parsing the OpenAPI spec and indexing operations by `operationId`.
- Translating an `Args` struct into an HTTP request (path parameter substitution from struct fields, body marshalling, header injection for auth).
- Decoding HTTP responses into a `State` struct.
- Surfacing common metadata via the registration: field renames (`pulumi name` → OpenAPI param name), secret fields, replace-on-change fields.
- Validating each registration against the loaded spec at startup — fields in `Args`/`State` not present in the operation, required spec fields without Go counterparts, missing operations. Failure means the provider exits before serving.

### What the library doesn't handle (yet)

- Polymorphic discriminated unions (RBAC permissions). Out of scope; resources requiring this stay hand-written.
- Pagination. Out of scope until needed.
- Custom `Diff` / `Check`. Default behavior comes from infer; resources with custom needs implement those methods directly on their wrapper struct.

### Spec management

- **Source.** A single OpenAPI 3 spec for the Pulumi Cloud API, fetched at codegen time from a URL (configurable via `PULUMI_CLOUD_OPENAPI_URL` env var; default TBD by user).
- **Storage.** Committed to repo at `provider/pkg/cloud/spec.json`. `//go:embed` pulls it into the binary; parsing happens once during `init()`.
- **Refresh.** A `make generate` target invokes a small fetcher tool (`provider/tools/openapi-fetch/`) which downloads, normalizes, and writes the spec. Re-run when upstream changes.
- **CI freshness check.** A workflow step runs `make generate` and fails if `git diff --exit-code provider/pkg/cloud/spec.json` shows changes. Stale specs fail PRs.

### Configuration sharing

Both the legacy custom server and the infer-driven resources receive the same configuration (access token, API URL) through the existing `Configure` flow.

The REST library exposes a small indirection — `rest.SetTransportResolver(func(ctx) (rest.Transport, error))` — that the provider wires once during init. The resolver typically reads `infer.GetConfig[config.Config](ctx)` and returns an authenticated HTTP client. This keeps the `rest` package free of any dependency on the project's `config` package.

## Build / CI

### Driven by ci-mgmt parameter flip

Flip `.ci-mgmt.yaml`:
```yaml
provider: cloud
env:
  PROVIDER: cloud
```
Run upstream regen — Makefile, goreleaser configs, and most workflows update mechanically.

### Net-new build wiring

- `go generate ./provider/pkg/cloud/...` invokes the spec fetcher (already wired via the `go:generate` directive in `provider/pkg/cloud/spec.go`).
- A `make generate` (or sibling) target should hook this in. The current Makefile is autogenerated by ci-mgmt, so the cleanest path is upstream support for an "extra generate" hook. Until that exists, contributors run `go generate` directly.
- A workflow step (in `lint.yml` or new `verify-generated.yml`) running the fetcher followed by `git diff --exit-code` against `provider/pkg/cloud/spec.json`. Stale specs fail PRs.
- Lint paths extend naturally; no other Makefile changes needed beyond the rename.

### Sequencing with ci-mgmt

1. PR to `pulumi/ci-mgmt` (if any code changes needed there). Likely none — templates are already parameterized off `provider:`.
2. `.ci-mgmt.yaml` flip + regen here, in one PR.
3. Source rename + Go module path update in a follow-up PR (touches every Go file).
4. SDK regen on top of the renamed schema in the same PR as #3.

## Implementation order

1. Add `provider/pkg/rest/` skeleton (this design).
2. Add `provider/pkg/cloud/` skeleton with one fixture resource for end-to-end validation.
3. Add `provider/tools/openapi-fetch/` + `go:generate` directive + `make generate` target.
4. Wire `cloud.All()` into `MakeProvider`'s `WithResources(...)`.
5. Add aliases to `manual-schema.json`.
6. Smoke test: build, dump schema, verify aliases and fixture token.
7. Rename: ci-mgmt flip → regen → Go module path → SDK package names.

Steps 1–6 ship in their own PR. Step 7 ships separately because it touches every file.

## Risks / open items

- **`pulumi-go-provider` mux primitives are already exercised here**, so the mux risk is low. Embedding `*rest.Controller[A, S]` in a per-resource named struct (so promoted methods + the right schema name both work) needs verification on a live build.
- **`ID` extraction policy.** REST library needs a way to know which response field is the resource ID. Initial plan: explicit `IDField string` in `rest.Ops`; if absent, fall back to a configurable convention (e.g. `id`, then the resource's primary key fields concatenated). Resolve during implementation.
- **OpenAPI spec source URL** is unknown to the design; user supplies. Fetcher tool reads from env var with no default — failing fast if unset.
- **Schema-level aliases on legacy resources** require editing `manual-schema.json`. Mechanical but bulky; the diff will be large in the rename PR.
- **Module path migration** touches every Go import in this repo plus `sdk/go/go.mod`. Mechanical via `gofmt -r` or `goimports`.
