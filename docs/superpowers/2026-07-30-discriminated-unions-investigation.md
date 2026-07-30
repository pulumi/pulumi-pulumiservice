# Discriminated unions in the PSP REST API generation — investigation notes

## 1. Status quo

`provider/pkg/rest/schema.go` `BuildSchema` emits **zero named types**. `out.Types` is
initialised (schema.go:32) and merged by `mergeSpec` (provider.go:300-307) but never populated.

`openAPIToType` (schema.go:620-652) degrades **every `$ref`** to `pulumi.json#/Any`:
```go
if _, ok := node["$ref"].(string); ok {
    return schema.TypeSpec{Ref: "pulumi.json#/Any"}
}
```

`flattenObjectSchema` (schema.go:558-563) hard-errors on `oneOf`/`anyOf`, but only at the
*root* of a request/response body. Discriminated types appear as nested properties, so they
never reach that check — they silently become `Any`.

Committed schema: `provider/cmd/pulumi-resource-pulumiservice/schema.json` has 41 types, all
`pulumiservice:index:*` (hand-written). The 45 `pulumiservice:api:*` resources contribute 0.
Example: `.resources["pulumiservice:api:Role"].inputProperties.details` = `{"$ref":"pulumi.json#/Any"}`
→ Go SDK `Details pulumi.AnyOutput` (sdk/go/pulumiservice/api/role.go:26).

## 2. What is actually in the OpenAPI spec

`provider/pkg/cloud/spec.json`: 738 component schemas.
- `oneOf`: 1 occurrence (EscSchemaSchema)
- `anyOf`: 1
- `discriminator`: 16
- `allOf`: 111

All 16 discriminators use the **Java/Jackson inheritance pattern**: base declares
`discriminator {propertyName, mapping}` with **no `oneOf`**; each subtype is
`allOf: [{$ref: base}, {extra props}]`.

### Reachable from declared resources (5 of 45 resources, 7 of 16 unions)

| Resource | Union | tag prop | variants |
|---|---|---|---|
| DeploymentSettings | DeploymentSettingsVCS | `provider` | 5 |
| ScheduledDeployment | DeploymentSettingsVCS | `provider` | 5 |
| Role | PermissionDescriptor | `__type` | 6 |
| Gate | ChangeGateRuleInput | `ruleType` | 1 |
| Gate | ChangeGateRuleOutput | `ruleType` | 1 |
| Gate | TargetEntity | `entityType` | 1 |
| Task | AgentEntity | `type` | 4 |
| Task | AgentUserEvent | `type` | 5 |

Two structurally different shapes:

**Fat base / thin variants** — `DeploymentSettingsVCS`: base has 8 props
(provider, repository, deployCommits, paths, installationId, previewPullRequests,
pullRequestTemplate, deployPullRequest). 4 of 5 variants add **nothing**; GitHub adds
`reviewStackLabels`. Flattening loses almost nothing.

**Thin base / fat variants** — `AgentEntity`, `PermissionDescriptor`: base is only the tag.
Variants carry disjoint fields **with colliding names**:
AgentEntityPolicyIssue{id,name} / AgentEntityPR{merged,number,repo} /
AgentEntityRepository{forge,host,name,org} / AgentEntityStack{name,project}.
Flattening is lossy and ambiguous here. A union earns its keep.

`PermissionDescriptor` is **recursive** (`PermissionDescriptorCondition.subNode` → `$ref`
back to `PermissionDescriptor`). `flattenObjectSchema` currently rejects cycles outright
(schema.go:536-540), so a named-type builder needs its own cycle handling.

## 3. Empirical codegen results (pulumi v3.242.0, `pulumi package gen-sdk`)

Probes in `scratchpad/union-probe`, `union-probe5`, `union-nested`, `flat-probe`, `edge`.

### A. Union with 2 members

| Language | Input | Output |
|---|---|---|
| TypeScript | `A \| B` | `A \| B` |
| Python | `Union[AArgs, BArgs]` | **`Output[Optional[Any]]`** |
| .NET | `InputUnion<A,B>` | `Output<Union<A,B>?>` |
| Java | `Either<AArgs,BArgs>` | `Output<Either<A,B>>` |
| Go | **`interface{}` / `pulumi.Input`** | **`pulumi.AnyOutput`** |

### B. Union with 5 members (the realistic case)

| Language | Input | Output |
|---|---|---|
| TypeScript | full 5-way union | full 5-way union |
| Python | full 5-way `Union[...]` | **`Output[Optional[Any]]`** |
| .NET | **`object?`** | **`Output<object?>`** |
| Java | **`Output<Object>`** | **`Output<Object>`** |
| Go | **`interface{}`** | **`pulumi.AnyOutput`** |

.NET `Union<T0,T1>` and Java `Either<L,R>` are two-slot only. **Beyond 2 members they
silently degrade to `object`.** No warning, no error.

**Mitigation:** the member types are still generated in every language
(VcsGitHub/VcsGitLab/... exist in Go pulumiTypes.go, .NET Api/Inputs+Outputs, Java inputs/outputs).
So even where the property is `object`, users get concrete constructible types.

### C. Flattened single named type (no union) — `flat-probe`

| Language | Input | Output |
|---|---|---|
| TypeScript | `DeploymentSettingsVcs` | `DeploymentSettingsVcs` |
| Python | `DeploymentSettingsVcsArgs` | `outputs.DeploymentSettingsVcs` |
| .NET | `Input<Inputs.DeploymentSettingsVcsArgs>` | `Output<Outputs.DeploymentSettingsVcs?>` |
| Java | `Output<DeploymentSettingsVcs>` | `Output<DeploymentSettingsVcs>` |
| Go | `DeploymentSettingsVcsPtrInput` | `DeploymentSettingsVcsPtrOutput` |

**A plain named type is strictly typed in all 5 languages.** A union is strictly typed in 1.

### D. Nested and recursive unions

Both work. A union inside a named type renders the same as at resource level
(`GateTarget.entityInfo` → `Either<TargetEnv,TargetStack>` in Java,
`Union<...>` in .NET, TS union). A self-referential union member
(`PermCond.subNode` → `PermAllow | PermCond`) generates cleanly in all languages.

### E. `discriminator` is INERT in codegen

Identical output with and without the `discriminator` block, in TS, .NET and Java:
- `nodisc` (oneOf only, 2 members) → .NET `Union<Outputs.A, Outputs.B>?`, Java `Either<A,B>`
- `union-probe` (oneOf + discriminator, 2 members) → .NET `Union<...>`, Java `Either<...>`

**The union shape comes from `oneOf` alone.** `discriminator` buys stricter schema validation
and nothing else in generated code.

`const` on a property IS honoured and is what makes TS narrowing work:
`{"type":"string","const":"github"}` → TS `provider: "github"` (a literal type).

### F. Schema validation edge cases

| Case | Result |
|---|---|
| `oneOf` with **1 member** | **REJECTED** by the meta-schema (`minItems: 2`) |
| `discriminator.mapping` → type not in the package | accepted, unvalidated |
| `discriminator.mapping` covering only some members | accepted, unvalidated |
| `discriminator.propertyName` naming a property no member has | accepted, unvalidated |
| `discriminator` with no `mapping` | accepted |
| `discriminator` without `oneOf` | accepted, silently dropped |
| `oneOf` mixing primitive + object | accepted (`string \| A` in TS) |
| `oneOf` without `discriminator` | accepted, identical codegen |

The **only** hard constraint is `oneOf` ≥ 2 members. The binder copies
`discriminator.mapping` verbatim into `UnionType.Mapping` and never resolves it
(`pkg/codegen/schema/bind.go:943-985`); `len(oneOf) < 2` is only a binder *diagnostic*, but
the metaschema (`pkg/codegen/schema/pulumi.json:337-378`) rejects it outright on `BindSpec`.

The 1-member rule matters: ChangeGateRuleInput/Output and TargetEntity each have exactly one
mapping entry, so they **cannot** be expressed as a `oneOf` at all. They must be emitted as
the single variant type (or the flattened base).

### G. Named union types are impossible

`ComplexTypeSpec` embeds `ObjectTypeSpec` and has **no `OneOf`/`Discriminator` field**
(`pkg/codegen/schema/schema.go:1610-1641`). `bindTypeDef` (`bind.go:754-805`) branches on
`spec.Type == "object"` and otherwise falls through to `bindEnumType`. The metaschema's
`complexTypeSpec` is `oneOf: [objectType, enumType]` (`pulumi.json:444-480`).

So a union can only appear **inline in a property's `TypeSpec`**. The `oneOf` block must be
duplicated at every use site. For `PermissionDescriptor` that means Role.details *and* the
recursive `subNode`/`entries`/`options` sites all repeat the same 6-member `oneOf`.

### H. Where the discriminator IS used: PCL and `pulumi import`

Not dead everywhere. `pcl/binder_schema.go:415-422` annotates the union only when a
discriminator is present; `pcl/binder_resource.go:144-186` (`resolveUnionOfObjects`) then
narrows a PCL object literal to the right member; `pkg/importer/hcl2.go:718-780`
(`reduceUnionType`) does the same for `pulumi import`.

**This matters directly for PSP.** `appendExamples` (schema.go:708-721) wraps each metadata
example in a ` ```pulumi ` fence, and SDK codegen runs `pulumi convert` on it. An example that
sets a union-typed property needs the discriminator to convert to the right member type.
`TestEveryApiResourceHasExample` means every resource has such an example.

Caveat: the mapping value format is parsed three different ways across consumers
(full token after stripping `#/types/`, bare `.Name()` comparison, and a `strings.Contains`
substring match in pulumi-java's program-gen). Expect friction.

## 4. Runtime implications

`BuildSchema` output is **write-only** at runtime. It has one non-test caller, the GetSchema
override (provider.go:265-289). The CRUD handlers hold only `meta` + `spec`
(resource.go:96-99) and never see a `schema.TypeSpec`.

- `buildRequestBody` (resource.go:902-918) reads the body schema for **top-level key names
  only**; values are discarded and passed through by `propertyValueToAny` with no coercion.
- `roundTrip` decode (resource.go:788-813) never consults `ResponseRef` — generic
  `json.Unmarshal` → `property.Map`.
- `Check` (resource.go:115-141) does no validation; it normalises enum case and sorts
  unordered arrays, **top level only**.
- `Diff` (resource.go:256-285) is structural `property.Value.Equals`, top-level keys only.

So enriching the schema changes what the engine and SDKs believe, and changes nothing about
what the provider sends. Consequences:

1. **Nested renames are the real gap.** `renameMapKeys` and `buildRequestBody` rename
   top-level keys only. Named types must therefore use **wire-side property names verbatim**
   for nested fields, or the declared shape and the wire shape diverge. (The spec is already
   camelCase, so this is natural.)
2. **Fail-open vs fail-closed.** `flattenObjectSchema` errors abort `BuildSchema`
   (schema.go:41-47) but are swallowed at resource.go:247 and resource.go:907, where they
   turn into "send the user's inputs unmodified". Any construct newly tolerated on the schema
   side needs a matching decision at those two sites.
3. **Nested secrets cannot be marked today** — `looksSecret` is applied to top-level names
   only. Named types would make nested secret marking possible for the first time.

## 5. Scope: how many named types is this?

Walking the `$ref` graph from every declared resource's request/response bodies:

- 86 root body schemas (these become resource input/output property sets today)
- **207** distinct component schemas reachable in total
- **121** candidate named types (nested, non-root) — this is the named-type prerequisite
- of those, **7** are discriminated bases; emitting their variants adds **22** more types
- worst single resource: `DeploymentSettings` pulls in 42 schemas; `ScheduledDeployment` 25

So roughly **~143 named types**. The union work is ~29 of them; the other ~121 is the
named-type foundation that must land first.

No prior attempts in git history. No test coverage for `Any` degradation or `oneOf` in
`provider/pkg/rest/schema_test.go`.

## 6. Runtime deserialisation of union outputs (RESOLVED — it is broken)

The schema `discriminator` is never read at runtime by any language SDK (zero hits in any
SDK runtime; it feeds only PCL program-gen and `pulumi import`). Union outputs deserialise by
**first-match structural probing**, and both typed implementations are broken for
object-vs-object unions:

- **.NET** (`pulumi-dotnet sdk Serialization/Converter.cs:291-311` `TryConvertOneOf`): tries
  T0 then T1. For `[OutputType]` classes conversion "succeeds" once the value is a struct —
  missing properties are tolerated, nested errors become warnings. **T0 always wins.** A
  payload for member B deserialises as member A with nulls. Silent.
- **Java** (`pulumi-java sdk Converter.java:504-541` `tryConvertOneOf` + `canBeCast:480-487`):
  probes `isInstance` against the raw `ImmutableMap` before ever building the POJO, so both
  probes fail for object members. **`Either<ObjectA,ObjectB>` outputs are always null**, with
  only a log warning.
- TS/Go/Python are safe (structural / untyped pass-through).

Consequences for the design:
1. An object-vs-object union on an **output** property is broken in .NET (silently wrong
   member) and Java (null) at ANY arity=2; at arity 3+ it degrades to `object`/`Object`,
   which is ugly but not wrong.
2. Input-side unions are fine everywhere (inputs are serialised, not deserialised).
3. `defaultType` is honoured only by Python and Node output paths; it does not rescue
   .NET/Java.
4. For PSP's reachable unions (arity 1, 4, 5, 6 — never 2) the broken `Union`/`Either` code
   path is never generated. But any future 2-variant union would silently step on it, so the
   generator must either never emit output-side object unions or clamp arity ≥ 3.

Upstream context: pulumi/pulumi#14674 (named union types in `types` — open feature request),
pulumi/pulumi#10995 (importer union reduction).

## 7. Precedent: pulumi-aws-native (measured locally, v1.31.0 schema)

- 9,711 types, 1,187 resources; **125 `oneOf` sites, 0 `discriminator`**.
- 107 of 125 are object-vs-object unions; members are auto-numbered variants
  (`FormFieldPosition0Properties`, `FormStyleConfig1Properties`, ...) — i.e. synthesised
  variant names, not semantic ones.
- Unions appear on **both** sides: 23 directly on `inputProperties`, 24 on output
  `properties`, plus all the shared `types` sites.
- **35 are 2-member object unions** — exactly the arity where .NET silently picks T0 and
  Java deserialises to null. aws-native ships this and lives with it (or the breakage goes
  unnoticed).

## 8. Spec provenance: Pulumi owns the whole pipeline

The discriminator blocks originate in hand-written Java specification models in
pulumi-service (`specification/src/main/java/com/pulumi/model/*.java`), e.g.
`DeploymentSettingsVCS.java` carries Jackson `@JsonTypeInfo(use = Id.NAME, property =
"provider")` + `@JsonSubTypes(...)`. Pulumi's own `pulumi-codegen` tool renders these to
`openapi.json` (Makefile `openapi_spec`, service repo), which PSP embeds as `spec.json`.

Implications:
- The **wire contract is tag-based Jackson polymorphic deserialisation** — the service
  *requires* the discriminator property in request bodies. Any Pulumi-side shape must keep
  the tag as a real field the runtime sends (both flatten and union do, since PSP's runtime
  is passthrough).
- The pattern is deliberate, stable, and generated — not incidental spec noise.
- There are internal extension hooks (`SwaggerVendorExtensions`: `x-pulumi-subtypes`, etc.,
  and `x-pulumi-model-property` with `enumTypeName` appears 166× in spec.json). If the
  provider generator needs richer hints (variant display names, flatten-vs-union
  preferences), the upstream spec generator can emit them — an option third-party providers
  never have.
- `x-pulumi-model-property.enumTypeName` is also the natural source for **named enum types**
  during the named-type foundation work.

## 9. Precedent: pulumi-azure-native (the OpenAPI-discriminator playbook)

azure-native (commit 9236e100b698, 2026-07-30) is the only production provider that ships
`discriminator` in its schema. Its rules, all in `provider/pkg/gen/`:

- `genDiscriminatedType` (types.go:627-688): base schema has a discriminator → emit
  `TypeSpec{OneOf, Discriminator{PropertyName, Mapping}}`. **The base type is never
  registered** — base properties are flattened (duplicated) into every variant.
- **0 subtypes → ignore the discriminator. 1 subtype → emit the definite type, no union**
  (types.go:666-671).
- Each variant carries the discriminator property with **`const: "<tag>"`**, marked
  required, plus a description hint "Expected value is 'X'." (properties.go:246-271).
- Emitted on **both** input and output positions (111 input / 114 output resource sites,
  877 sites in types; read-only discriminators suppress the input side only).
- **Resource-level** discriminated PUT bodies become one resource per variant ("resource
  variants", schema.go:857-876) because a Pulumi resource cannot be a oneOf. (Not needed for
  PSP — all our unions are property-level.)
- At scale: DataFactory LinkedService is a **121-member** union — TS gets the full union,
  .NET/Java get `object`, and azure-native accepts that.
- Runtime (its typed converter, convert/sdkInputsToRequestBody.go:112-126): first-match
  trial conversion where a variant's `const` mismatch rejects the branch — i.e. the
  discriminator drives member selection indirectly via const. Missing tag → first member
  wins. (PSP's passthrough runtime needs none of this.)
- Real-world failure mode (issue #2671): Go users pass a wrong untyped shape, the tag never
  reaches the wire, Azure rejects. TS would have caught it at compile time. The lesson:
  unions improve safety where typed, and the service must reject bad payloads — which the
  Pulumi service already does (Jackson).

## 10. Precedent: the rest of the ecosystem

- **aws-native**: 223 `oneOf`, **0 discriminator**. Inline anonymous unions become
  numeral-suffixed member types (`ApiSchema0Properties`; open issue #994 complains). Named
  object schemas with oneOf are **merged** into one object when branches don't conflict
  (`mergeObjectUnionProperties`, jsschema_util.go:44-72 — "Cloud Control remains
  responsible for enforcing the union constraint"), dropping branch-specific requireds.
  Object-bearing type lists collapse to Any.
- **kubernetes**: 69 `oneOf` sites, all primitive-vs-object or primitive-vs-primitive
  (IntOrString, JSONSchemaPropsOrBool...). **Zero object-vs-object unions.** Quantity is
  deliberately collapsed to plain string.
- **terraform-bridge**: unions only via hand-authored `AltTypes`, always with a primitive
  `type` default, applied to inputs only; sets `DisableUnionOutputTypes` for Node.
- **google-native, awsx**: zero unions.
- **No provider besides azure-native ships `discriminator`.**

## 11. Upstream trajectory (2026): discriminated unions are being fixed

A conformance test `l2-discriminated-union` (object-vs-object union + discriminator) now
exists in pulumi/pulumi (`pkg/testing/pulumi-test-language/tests/l2_discriminated_union.go`).
Status verified via gh 2026-07-30:

- Go: pulumi/pulumi#21829 **CLOSED** (fixed)
- .NET: pulumi-dotnet#866 **CLOSED** (implemented)
- Python: pulumi/pulumi#21830 **OPEN** (still failing; also #21832 loose output types)
- Java: pulumi-java#2021 **OPEN** (unimplemented)

So the per-language degradation and deserialisation bugs in §3/§6 are point-in-time, and
upstream investment is active. Named union types remain unsupported (pulumi/pulumi#14674
open, no movement).

## 12. Not established

- **PCL example conversion with a union is untested.** My probe produced an empty
  `## Example Usage`, but so did the plain-named-type control, so example conversion simply
  was not running for a synthetic package in this sandbox. The result is inconclusive, not
  negative. Validate this during implementation — it is the highest-risk unknown, because
  `TestEveryApiResourceHasExample` forces every resource to carry a PCL example.
  (azure-native's example pipeline handles union-typed properties in practice, which is
  weak positive evidence.)
- Whether pulumi-dotnet#866 changed the runtime `Converter.TryConvertOneOf` first-match
  behaviour or only the conformance/codegen layer (the deserialisation review found main's
  Converter.cs unchanged vs v3.101.0).

## 13. Recommendation (post-research)

Adopt the azure-native playbook, uniformly, as phase 2 on top of the named-type foundation:

1. **1-variant bases** (ChangeGateRuleInput/Output, TargetEntity): emit the definite variant
   type. No union. (azure-native rule; also forced by the metaschema's `minItems: 2`.)
2. **Multi-variant bases** (DeploymentSettingsVCS, PermissionDescriptor, AgentEntity,
   AgentUserEvent — arity 4-6): emit inline `oneOf` over named variant types +
   `discriminator{propertyName, mapping}`; flatten base properties into each variant; tag
   property gets `const` + required on each variant. Do not register the base type.
3. **Guard**: fail generation on a 2-member object union in an output position until the
   upstream .NET/Java deserialisers are proven fixed (no current PSP union has arity 2, so
   the guard is free today and forces an explicit decision if the spec ever adds one).
4. Runtime: no change needed — property-level unions pass through `buildRequestBody`
   verbatim, and the service enforces the tag via Jackson.
5. Uniform rule beats per-union special-casing (flatten-the-fat-base was considered):
   one code path, matches ecosystem precedent, faithful to future spec evolution, and the
   variant types are generated in every language even where the property degrades.
