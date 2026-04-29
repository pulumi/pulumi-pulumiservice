# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working
with code in this repository. For the long-form maintainer playbook —
including the full refresh workflow, the decision tree for new
operations, and the anti-patterns to avoid — see
**[docs/MAINTAINING.md](docs/MAINTAINING.md)**.

## Project Overview

This is the Pulumi Service Provider (PSP), a Pulumi provider built on
top of the Pulumi Cloud REST API. It lets users manage Pulumi Cloud
resources (Stacks, Environments, Teams, Tokens, Webhooks, Deployment
Settings, etc.) from Pulumi programs.

**The provider is generated, not hand-coded. There is no escape
hatch — every supported resource lives in `resource-map.yaml`.** Two
inputs compose into the shipped binary + SDKs, both `//go:embed`-ed
directly into the runtime:

1. `provider/pkg/embedded/openapi_public.json` — pinned copy of the
   Pulumi Cloud OpenAPI 3.0.3 spec.
2. `provider/pkg/embedded/resource-map.yaml` — the editable mapping
   from operationIds to Pulumi resources/functions/methods (plus
   per-property metadata — renames, secrets, defaults, force-new,
   validation checks).

There is one binary: `pulumi-resource-pulumiservice`. `go build` is
all it takes. The runtime is built on
`github.com/pulumi/pulumi-go-provider`; it derives the runtime
metadata from the embedded inputs at startup, and serves the Pulumi
schema lazily via `GetSchema` (regenerated from the same embedded
inputs on first call). The CRUD dispatcher in `provider/pkg/runtime/`
consumes the metadata.

`GetSchema` errors loudly if the embedded resource-map is incomplete
(any operationId in the spec without a claim — resource, function,
method, or explicit exclusion). This is the same coverage gate that
`go test ./provider/pkg/embedded/...` runs in CI.

If a Pulumi Cloud API pattern can't be expressed in the current
metadata shape, the answer is to **extend the metadata schema** in
`provider/pkg/runtime/metadata.go` (plus parse + dispatch), not to add
per-resource Go. The current primitives that cover the non-trivial
cases:
- `createSource`/`createFrom` — per-verb property source/wire-name
  rename (`name` is body on POST, path on subsequent verbs).
- `pathName` — URL path placeholder name when it differs from the
  response/body wire name (AgentPool `id` ↔ `{poolId}`).
- `bodyFrom` — request-body wire name when it differs from the
  response wire name (LogExport `newEnabled` body ↔ `enabled` response).
- `source: pathAndBody` — identity field that goes in BOTH the URL
  path and the request body, with the API validating they agree
  (Webhook).
- `bodyAs` — operation-level: a single property's value IS the entire
  request body (StackTags PATCH expects `{tagName: value}` directly).
- `bodyOverride` — tombstone-style delete via update op
  (TeamStackPermission).
- `readVia.extractField`/`keyBy` — read as a field on a parent
  resource's GET.
- `iterateOver`+`iterateKeyParam` — delete one call per map key
  (Tags batch).
- `rawBodyFrom`/`rawBodyTo`+`contentType` — non-JSON bodies (ESC
  Environment YAML PATCH).
- `postCreate` — two-step create.
- Polymorphic scope inference — when a polymorphic resource (Webhook)
  has no explicit `discriminator:` field, the dispatcher picks the
  scope whose ID-template path placeholders are all satisfied by user
  inputs; output-only placeholders (server-assigned names) are
  ignored.

## Repository Structure

```
provider/
  cmd/
    pulumi-resource-pulumiservice/ provider binary (single binary; `go build`)
  pkg/
    embedded/                      //go:embed of openapi_public.json + resource-map.yaml
                                   (THE editable mapping lives here)
    gen/                           schema/metadata/coverage emitters as a library
    runtime/                       metadata-driven CRUD dispatcher
    provider/                      pulumi-go-provider Provider literal
    version/                       version variable injected via LDFLAGS

sdk/                               generated SDKs; never edit by hand

examples/
  canonical/                       12 end-to-end user-story programs
  yaml-<resource>/                 per-resource smoke-test programs
  examples_yaml_test.go            integration test registration
  examples_canonical_test.go       canonical-scenario test registration

docs/
  MAINTAINING.md                   long-form maintainer playbook
  UPGRADE-v1-to-v2.md              upgrade guide for v1 users
```

**Never edit `bin/*.json` or `sdk/*` by hand** — they're regenerated.
`CLAUDE.md` and `docs/MAINTAINING.md` are the two human-authored docs.

## Build Commands

### Provider binary

```bash
go build ./provider/cmd/pulumi-resource-pulumiservice  # plain `go build` works
make v2_provider                                       # the same plus pulumi package get-schema
                                                       # writes provider/cmd/.../schema.json for SDK gen
make coverage_report                                   # `go test` against the coverage gate
make coverage_report_strict                            # alias of coverage_report (the test is strict by default)
make update_spec                                       # refresh embedded spec from sibling ../pulumi-service/
```

### SDK generation + build

```bash
make build_sdks           # regenerate all five language SDKs from schema.json
make nodejs_sdk           # one language at a time
make build                # everything: provider + SDKs
```

### Testing

```bash
make test_provider        # unit tests in provider/pkg/...

# Integration tests require PULUMI_ACCESS_TOKEN + PULUMI_TEST_OWNER.
cd examples && go test -tags=yaml -v -timeout 3h ./...           # per-resource
cd examples && go test -tags=canonical -v -timeout 3h ./...      # canonical scenarios
cd examples && go test -tags=all -v -timeout 3h ./...            # everything
```

### Linting

```bash
make lint                 # golangci-lint in provider/, sdk/, examples/
```

When linting `examples/`, use `--build-tags all` so test files under
build tags are included:

```bash
cd examples && golangci-lint run --timeout 10m --build-tags all
```

## Environment Variables

Used by integration tests (local `.env` is gitignored):

- `PULUMI_ACCESS_TOKEN` — token for Pulumi Cloud.
- `PULUMI_TEST_OWNER` — test org (defaults to
  `service-provider-test-org`).
- `PULUMI_BACKEND_URL` — optional API URL override (consumed by the
  provider's `apiUrl` config).

## Adding or changing a resource

**Read [docs/MAINTAINING.md](docs/MAINTAINING.md) first.** Quick
reference:

### Declarative (common case)

Edit `provider/pkg/embedded/resource-map.yaml`, under the right sub-module:

```yaml
orgs/<module>:
  resources:
    ResourceName:
      doc: One-sentence description.
      operations:
        create: CreateOp      # spec operationId
        read:   GetOp
        update: UpdateOp
        delete: DeleteOp
      id:
        template: "{organizationName}/{fooId}"
        params: [organizationName, fooId]
      forceNew: [organizationName]
      properties:
        organizationName:
          from: orgName         # wire name if different from SDK name
          type: string
          source: path           # path | query | body | response
          required: true
          doc: …
        # output-only: set `output: true` and `source: response`
```

Then:

```bash
make v2_provider && make build_sdks
cd provider && go test ./...
```

Add a YAML example under `examples/yaml-<resource>/` and register it in
`examples/examples_yaml_test.go`. Update `CHANGELOG.md` under the
current unreleased section.

### What if the resource doesn't fit the metadata shape?

Extend the metadata schema — don't add hand-coded Go. Every known
irreducible case in v2 has been expressed declaratively:

- Upsert (no separate POST): reuse the same operationId for create +
  update. See `LogExport`, `Settings`, `TeamStackPermission`.
- Tombstone-style delete: `delete: { operationId: …, bodyOverride: {…} }`.
- Per-verb property source or wire rename: `createSource:` /
  `createFrom:` on the property.
- Child resources with no dedicated GET: `readVia: { operationId: …,
  extractField: …, keyBy: … }` — piggyback on the parent's GET.
- Batch-over-children delete: `delete: { operationId: …,
  iterateOver: <prop>, iterateKeyParam: <path-placeholder> }`.
- Non-JSON bodies: property `source: rawBody` + op `rawBodyFrom: <prop>`
  / `rawBodyTo: <prop>` / `contentType: …`.
- Multi-step create: `postCreate:` on the resource (runs a follow-on op
  with update-style input handling after the primary create).

If a new case comes up that none of those cover, the right move is to
add a new primitive — edit `provider/pkg/runtime/metadata.go`, wire it
through `provider/pkg/gen/metadata.go` and the dispatcher, add a test,
and use it. Do not reintroduce a `customresources/` package.

## Resource naming — no module stutter

In `pulumiservice:<module>:<Name>`, the module provides context. Don't
repeat it in the type name:

- ✅ `pulumiservice:orgs/oidc:Issuer` (not `OidcIssuer`)
- ✅ `pulumiservice:orgs/insights:Account` (not `InsightsAccount`)
- ✅ `pulumiservice:stacks/tags:Tag` (not `StackTag`)

Exception when the un-prefixed name would be too generic or collide:
- `pulumiservice:orgs/identity:IdentityProvider` — bare `Provider`
  would clash with `pulumi.Provider`.
- `pulumiservice:orgs/agents:AgentPool` — bare `Pool` is meaningless.

When adding a new resource, read the fully-qualified token out loud
before committing the name.

## Schema must NOT be hand-edited

`provider/cmd/pulumi-resource-pulumiservice/schema.json` and
`metadata.json` are regenerated outputs. The generator writes them
from `resource-map.yaml` + the OpenAPI spec; the provider's runtime
`GetSchema` further merges in custom-resource schema contributions.

If you need to change the schema, edit `resource-map.yaml` or the
custom resource's `Schema()` method. Never hand-edit the JSON.

## Copyright Headers

All new files use the current year in the range:

- In 2026: `// Copyright 2016-2026, Pulumi Corporation.`
- In 2027: `// Copyright 2016-2027, Pulumi Corporation.`

## CHANGELOG Updates

Update `CHANGELOG.md` when making code changes that are user-visible:

- `### Improvements` — new features, new resources, new properties.
- `### Bug Fixes` — bug fixes.
- `### Breaking Changes` — breaking changes (rare outside majors).

Format: `- Description of change [#<issue-or-pr>](<link>)`.

**Do NOT** add CHANGELOG entries for:
- Test-only changes.
- Doc updates (README, MAINTAINING.md, CLAUDE.md).
- CI/build tooling changes.

## Testing conventions

- Use `pulumitest` (`github.com/pulumi/providertest/pulumitest`) for
  new integration tests. Do not use the legacy `integration.ProgramTest`.
- Unit tests live alongside the code they test (`_test.go` suffix) in
  `provider/pkg/...`.
- Every new resource should have a YAML example that doubles as an
  integration test — either in `examples/yaml-<name>/` or as part of
  a canonical scenario under `examples/canonical/<n>-<scenario>/`.
- Example tests are tagged; use `-tags=yaml`, `-tags=canonical`, or
  `-tags=all` as appropriate.

## Sub-module naming

Resources live under sub-modules that mirror the Pulumi Cloud URL
hierarchy:

- `orgs`, `orgs/agents`, `orgs/teams`, `orgs/tokens`, `orgs/oidc`,
  `orgs/policies`, `orgs/templates`, `orgs/audit`, `orgs/cmk`,
  `orgs/members`, `orgs/roles`, `orgs/services`, `orgs/policypacks`,
  `orgs/insights`, `orgs/identity`
- `stacks`, `stacks/hooks`, `stacks/deployments`, `stacks/tags`,
  `stacks/permissions`
- `esc`, `esc/schedules`, `esc/versions`, `esc/permissions`,
  `esc/cloudsetup`
- `integrations`, `changegates`, `changerequests`

Tokens use `/` between nested modules (Pulumi-native convention):
`pulumiservice:orgs/teams:Team`, not `pulumiservice:orgs:teams:Team`.

## Release process

Releases are handled by Pulumi employees via the `#release-ops` Slack
channel. GitHub Actions build, test, and publish.

## Provider configuration

Two config variables (surfaced in every SDK's `Provider` resource):

- `accessToken` (env: `PULUMI_ACCESS_TOKEN`) — secret Pulumi Cloud
  token.
- `apiUrl` (env: `PULUMI_BACKEND_URL`, default
  `https://api.pulumi.com`) — API endpoint override for self-hosted
  Pulumi Cloud.

## Spec refresh

See [docs/MAINTAINING.md](docs/MAINTAINING.md) for the full workflow.
Short version:

```bash
make update_spec                   # pull from sibling pulumi-service
make coverage_report               # see what changed
# triage each unmapped op into resources/functions/methods/exclusions
make v2_provider && make build_sdks
cd provider && go test ./...
```

## Anti-patterns to avoid

- **Editing `provider/cmd/pulumi-resource-pulumiservice/schema.json`
  directly.** That file is regenerated by `make v2_provider` (it's
  the binary's `GetSchema` output, captured for `make build_sdks`).
  Edit `provider/pkg/embedded/resource-map.yaml` instead.
- **Editing `sdk/**` directly.** Same — they're generated.
- **Reintroducing a `pulumi-gen-pulumiservice` binary or a
  `provider/pkg/customresources/` package.** v2 ships as a single
  binary with `go build`; the metadata schema is the extension
  surface. If a new Pulumi Cloud pattern doesn't fit, extend
  `provider/pkg/runtime/metadata.go` (plus parse + dispatch).
- **Excluding a new operationId without a written reason.** Every
  `exclusions:` entry must have a 1-sentence reason so a later
  maintainer can revisit.
- **Copying a v1 resource name verbatim into v2.** Apply the
  anti-stutter rule; the sub-module gives context.
- **Shipping without regenerating SDKs.** `make build_sdks` is
  load-bearing; without it users see a new property in the schema but
  can't use it from their language.
