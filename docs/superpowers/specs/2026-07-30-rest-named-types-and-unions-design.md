# Named types and discriminated unions for the `api` namespace

Status: approved design, pending implementation plan.
Reference: [../2026-07-30-discriminated-unions-investigation.md](../2026-07-30-discriminated-unions-investigation.md) for all evidence and citations.

## Problem

`rest.BuildSchema` emits no named types. `openAPIToType` (provider/pkg/rest/schema.go)
degrades every `$ref` and anonymous object to `pulumi.json#/Any`, so nested inputs and
outputs (e.g. `Role.details`, `DeploymentSettings.operationContext`) are untyped `any` in
every SDK. The OpenAPI spec also models 16 discriminated (tagged-union) types, 7 reachable
from declared resources; these degrade to `Any` as well.

The loudest instance is RBAC: `Role.details` is the recursive `PermissionDescriptor`
union, so customers hand-write `__type`-tagged maps with no object model (issue #929;
the #project-psp-from-openapi "any for complex resources" thread). Phase 2 types that tree
structurally — the "translate inheritance to a discriminated union of objects" approach
from that thread. Issue #938 (Role drift not detected) is a runtime refresh problem and is
NOT addressed by this design.

## Decisions (made 2026-07-30)

1. **Phase 0 first: a POC of discriminated unions in all 5 languages** (added after the
   internal call). Hand-edit the committed schema, generate and compile every SDK, run real
   programs against the locally built provider, and record what works and what breaks per
   language. Failures become concrete reports to the Pulumi core team. No generator changes;
   the POC isolates the upstream surface (codegen, SDK runtimes, convert, import).
2. **Two production phases after the POC.** Phase 1 lands the named-type foundation with
   unions still degrading to `Any`. Phase 2 adds discriminated unions on top. Each phase is
   separately reviewable.
3. **Phase 2 follows the azure-native playbook, uniformly.** No per-type special-casing.

## Phase 0: POC (discriminated unions, all languages)

Vehicle: `Role.details` as the recursive `PermissionDescriptor` union (the RBAC case), plus
`DeploymentSettings.vcs` (the fat-base case). Steps:

1. Patch `provider/cmd/pulumi-resource-pulumiservice/schema.json` by script: add named
   variant types (const-tagged, base flattened in), rewrite the two properties to inline
   `oneOf` + `discriminator` on input and output sides. Wire names verbatim.
2. `pulumi package gen-sdk` for all 5 languages; compile each SDK.
3. One example program per language building a nested descriptor tree (Allow inside
   Condition/Compose) with the typed variant classes where the language offers them, a raw
   map where it does not. Run `pulumi preview` (and `up`/`refresh`/`destroy` against a test
   org) with the locally built provider binary. The stock runtime passes unions through, so
   no provider code changes are needed.
4. `pulumi convert` of a PCL example that sets a union property, to all 5 languages —
   retires the examples risk before phase 2 commits to it.
5. Optional: `pulumi import` of the created Role (exercises `reduceUnionType`).

Success criteria per language: SDK compiles; program expresses the union (typed or raw);
create body carries the tag; refresh round-trips; convert emits compiling code. Output: a
findings table plus a drafted upstream issue for each language-level failure.

## Phase 1: named types

### Scope

Emit a named Pulumi type for every component schema reachable from a declared resource's
operation bodies, excluding root body schemas (their properties already flatten into the
resource) and excluding schemas that declare a `discriminator` (those stay `Any` until
phase 2, so phase 2 is purely additive and never removes a phase-1 type). That is **114
types today** (121 nested reachable, minus 7 discriminated bases; measured via `$ref` walk).
Do not emit unreachable schemas (738 total in spec.json).

### Rules

- **Token**: `pulumiservice:api:<ComponentSchemaName>` in `PackageSpec.Types`. If the token
  collides with a resource token, append `Properties` to the type name (two known cases
  today: `CustomVCSRepository`, `ServiceItem` → `CustomVCSRepositoryProperties`,
  `ServiceItemProperties`; language codegen would otherwise emit two classes with one name
  in the same module). Any remaining collision fails `BuildSchema` (aggregate errors, same
  pattern as today).
- **Property names**: wire-side names verbatim. No nested renames. The runtime renames
  top-level keys only (`renameMapKeys`, `buildRequestBody`), so nested declared names must
  equal wire names or the schema lies. The spec is already camelCase.
- **allOf**: flatten chains into the named type (reuse `flattenObjectSchema` semantics).
- **Cycles**: recursive `$ref`s (e.g. `PermissionDescriptorCondition.subNode`) are legal for
  named types; the type builder needs its own visited-set that emits a token reference
  instead of erroring (today's `flattenObjectSchema` rejects cycles).
- **`$ref` in a property position**: emit `{"$ref": "#/types/pulumiservice:api:<Name>"}`
  instead of `pulumi.json#/Any`.
- **Anonymous nested objects**: keep today's behavior (free-form object of Any). Only named
  component schemas become named types. (YAGNI; revisit if the spec grows anonymous shapes.)
- **oneOf/anyOf/discriminator**: phase 1 keeps today's degradation to `Any` for any
  property referencing a discriminator-bearing schema (see Scope), and
  `flattenObjectSchema` keeps its `oneOf`/`anyOf` errors.
- **Secrets**: apply `looksSecret` and `FieldMeta.Secret` to named-type properties. This is
  the first time nested secrets can be marked; verify no known-sensitive nested field is
  missed.
- **Enums**: where a property carries `enum` (+ optional
  `x-pulumi-model-property.enumTypeName`), keep inline string typing in phase 1. Named enum
  types are a candidate phase 3: `PermissionDescriptorAllow.permissions` items carry the
  full 191-value `RbacPermission` enum with `enumTypeName`, so named-enum emission would
  directly answer the enum half of issue #929. Not in phase-1 scope.
- **Metadata**: no new metadata.json fields required for phase 1. `fields` metadata continues
  to apply to top-level resource properties only.

### Runtime

No behavior change. The runtime is passthrough (`buildRequestBody` sends top-level keys;
values serialize verbatim; decode is generic JSON). Two consistency notes to carry into
implementation:

- The fail-open sites (`resource.go:247`, `resource.go:907`) swallow `flattenObjectSchema`
  errors that fail `BuildSchema`. Phase 1 must not widen the gap: any construct the schema
  path newly tolerates must behave identically at those sites.
- Diff stays top-level structural; nested typed objects still collapse to one property diff.
  Acceptable; unchanged from today.

### Testing

- Unit tests in `rest/schema_test.go`: named type emission, token mapping, cycle handling,
  collision failure, secret marking on nested properties.
- Golden: regenerate `provider/cmd/pulumi-resource-pulumiservice/schema.json`; review the
  diff (expect ~121 new types and many `Any` → `$ref` property changes).
- Existing guards must stay green: `TestScaffoldMetadataIdempotent`,
  `TestEveryApiResourceHasExample`.
- SDK build for all 5 languages (`make build_sdks`) to catch name collisions in generated
  code (e.g. a type name colliding with a resource class in the same module).

## Phase 2: discriminated unions

### Rules (azure-native playbook)

For each component schema with a `discriminator` block, reachable from a declared resource:

- **Arity 0** (no mapping entries): ignore the discriminator; emit the base as a plain named
  type.
- **Arity 1** (ChangeGateRuleInput, ChangeGateRuleOutput, TargetEntity): emit the single
  variant as the definite type at the use site. No `oneOf` (metaschema requires >= 2).
- **Arity 2+** (DeploymentSettingsVCS 5, PermissionDescriptor 6, AgentEntity 4,
  AgentUserEvent 5): at every position that references the base (property, array `items`,
  map `additionalProperties`), emit inline `oneOf: [<variant refs>]` +
  `discriminator: {propertyName, mapping}` with mapping values
  `#/types/pulumiservice:api:<VariantName>`. Adds ~22 variant types.
- **Variants**: emit each variant as a named type with the base's properties flattened in.
  The tag property gets `const: "<tag>"`, is required, and gets "Expected value is '<tag>'."
  appended to its description. The base type itself is not registered.
- **Guard**: `BuildSchema` fails on a 2-member object union in an output position. The .NET
  and Java runtime deserializers pick wrong/null at that arity (verified upstream); no
  current union has arity 2, so the guard is free and forces an explicit decision if the
  spec adds one. Re-evaluate when pulumi/pulumi#21830 and pulumi-java#2021 close.
- **Body-root unions**: `flattenObjectSchema` keeps erroring on `oneOf`/`anyOf` at the body
  root (none exist in reachable operations). Resource-level polymorphism (azure-native's
  "resource variants") is out of scope.

### Runtime

No change. Union-typed property values pass through verbatim; every variant carries its tag
as a required field, and the service enforces it (Jackson `@JsonTypeInfo`). `Check` remains
non-validating for union members.

### Expected SDK rendering (verified empirically, pulumi v3.242.0)

- TypeScript: full union on inputs and outputs.
- Python: `Union[...]` inputs, `Any` outputs.
- Go: untyped slot (`pulumi.Input` / `pulumi.AnyOutput`); variant types generated.
- .NET / Java: `object` slot at arity 3+; variant types generated.
- Slots tighten as upstream l2-discriminated-union conformance fixes land (Go closed, .NET
  closed, Python and Java open); regenerating SDKs picks them up with no schema change.

### Validation risk to retire first

`TestEveryApiResourceHasExample` forces a PCL example per resource, and SDK docgen runs
`pulumi convert` on them. Convert of a union-typed property is unverified (sandbox probe was
inconclusive). First implementation step of phase 2: hand-write one PCL example setting
`DeploymentSettings.vcs` and run it through the real schema + `pulumi convert` for all
languages. If convert cannot narrow the member, fix or scope examples before generalizing.

### Testing

- Unit tests: arity 0/1/2+ paths, const injection, mapping emission, output-arity-2 guard,
  recursive union (PermissionDescriptor) termination.
- Golden schema diff review; SDK builds for all 5 languages.
- One end-to-end example per multi-variant union in metadata examples, converted in docs.

## Out of scope

- Runtime type coercion, nested renames, nested `Check` validation, detailed nested diffs.
- Named enum types from `x-pulumi-model-property`.
- Upstream spec-generator changes (available as an escape hatch; Pulumi owns the pipeline).
- The other Phase-0 known issues (query params, update-clear, output allowlist at runtime).
