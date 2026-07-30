# POC: discriminated unions in the PSP schema — findings

Branch `claude/union-poc`. Everything below was produced and verified locally on
2026-07-30 with pulumi v3.242.0.

## Update: the generator now produces this automatically

`provider/pkg/rest/types.go` implements the demand-driven type registry inside
`BuildSchema`: named types for every reachable component schema, discriminated bases as
inline `oneOf` + `discriminator` over const-tagged flattened variants, 1-variant bases as
definite types, cycle-safe recursion, module assignment by reachability (shared types hoist
to `api`), resource-token collisions suffixed `Properties`, and a hard error on 2-member
object unions. The committed `schema.json` is now `pulumi package get-schema ./bin/...`
output, no hand patch, and the `POC_SCHEMA_FILE` hook is gone.

Mechanical run over the full spec:

- **146 named api types** across 9 modules, **19 discriminated union sites** (the hand
  patch had 11 types / 4 sites). New unions the patch never covered: 3-member
  `ApprovalRuleEligibility` on Gate, 4-member `AgentEntity` on Task, and the recursive
  **17-member `PermissionExpression`** tree.
- `Role.details` is byte-equal to the hand patch (modulo descriptions);
  `Gate.rule` correctly collapsed to the definite `ChangeGateApprovalRuleInput` (1-variant
  rule); previously-`Any` nested objects (`operationContext`, `cacheOptions`, ...) are now
  named types.
- Zero generator errors, zero guard trips; all existing provider tests pass (10 packages),
  including scaffold idempotency and example coverage.
- All 5 SDKs regenerate; TS/Go/.NET compile checked; the TS example previews with tags
  intact; `pulumi convert` binds the provider's own GetSchema with no override and
  reproduces the same per-language results (including the Go `__type` bug).

## What the POC is

`poc/patch_schema.py` rewrites the committed schema
(`provider/cmd/pulumi-resource-pulumiservice/schema.json`) to the azure-native union shape
for two properties:

- `pulumiservice:api:Role.details` → 6-variant recursive `PermissionDescriptor` union
  (tag `__type`) — the RBAC case.
- `pulumiservice:api/deployments:Settings.vcs` → 5-variant `DeploymentSettingsVCS` union
  (tag `provider`).

Variants are const-tagged named types with base properties flattened in. SDKs are generated
into `sdk-poc/`, example programs live in `poc/examples/<lang>`, convert output in
`poc/convert/`. The provider binary gained a POC-only `POC_SCHEMA_FILE` env override in
`withCloudApiSchema` so `pulumi convert` binds the patched schema.

## Result matrix

| Stage | TS | Python | Go | C# | Java |
|---|---|---|---|---|---|
| SDK compiles | PASS | PASS | PASS | PASS | PASS |
| Union slot typing | full union | `Union[...]` in / `Any` out | `interface{}` | `object` (6 > 2 members) | `Object` (6 > 2) |
| Typed variant classes usable | PASS (literal-typed tag; typo caught by tsc with "Did you mean") | **FAIL** (`__type` ctor arg name-mangled to `_PermissionDescriptorAllowArgs__type`; required yet value ignored) | **FAIL** (`__type` is an unexported struct field; cannot be set) | PASS (`__type` settable; `InputList<object>` initializer ambiguity papercut) | PASS (`.__type(...)` builder) |
| Program compiles | PASS (typed) | PASS (dict form) | PASS (raw `pulumi.Map`) | PASS (typed, `object[]` workaround) | PASS (typed builders) |
| `pulumi preview` — `__type` reaches the engine at both nesting levels | PASS | PASS | PASS | PASS | PASS |
| `pulumi convert` from PCL | PASS (dict form) | PASS (dict form) | **FAIL** — narrows to `PermissionDescriptorAllowArgs` but emits the unexported field: `cannot refer to unexported field __type in struct literal` | PASS (typed) | PASS (typed; imports all 6 variant classes, cosmetic) |
| `pulumi up` / `refresh` round-trip | pending (needs a test org) | pending | pending | pending | pending |

Also verified: `pulumi preview --diff` HIDES `__`-prefixed keys in its display (the plan
file proves they serialize) — cosmetic but confusing.

## Per-language failures to report upstream

All three hard failures are one root cause: **property names with leading underscores**
(`__type`, the Jackson discriminator convention) break identifier generation. Azure never
hits this because its discriminators are `type`/`kind`/`odataType`.

1. **pulumi Go SDK-gen + program-gen** (pulumi/pulumi): a schema property named `__type`
   becomes an unexported Go struct field (`__type pulumi.StringInput`). Users cannot set it;
   `pulumi convert` emits struct literals that do not compile
   (`cannot refer to unexported field __type`). Expected: mangle to an exported name
   (e.g. `Type_`) with the `pulumi:"__type"` tag preserved.
2. **pulumi Python SDK-gen** (pulumi/pulumi): a `__`-prefixed required property generates
   `def __init__(__self__, *, __type: ...)`; Python name-mangles the parameter to
   `_PermissionDescriptorAllowArgs__type`, so the documented name is uncallable. Bonus bug:
   for `const` properties codegen hardcodes the value (`pulumi.set(__self__, "__type",
   'PermissionDescriptorAllow')`) but still declares the parameter required — it should be
   optional/defaulted (TS side already treats const-with-required correctly via literal
   types).
3. **pulumi .NET SDK** (pulumi/pulumi-dotnet, minor): `InputList<object>` collection
   initializers are ambiguous for T=object (`Add(params Input<object>[])` vs
   `Add(InputList<object>)`); affects any union-typed array. Workaround: assign `object[]`.
4. Cosmetic (pulumi/pulumi CLI): `preview --diff` display omits `__`-prefixed keys, so the
   discriminator looks missing even though it serializes.

Known upstream context to link: l2-discriminated-union conformance work — Go #21829
(closed), .NET pulumi-dotnet#866 (closed), Python #21830/#21832 (open), Java
pulumi-java#2021 (open); named union types pulumi/pulumi#14674.

## What this means for the design

- The azure-native playbook works end to end today in TS, C#, and Java, and via dict/map
  forms in Python and Go. Nothing blocks phase 1/phase 2.
- The `__type` identifier problem is the single upstream fix that matters for PSP.
  Options while it is open: (a) ship unions anyway — dict/map forms work everywhere and are
  no worse than today's `Any`; (b) rename the Pulumi-side property (schema-level rename of
  the tag) — rejected: the runtime has no nested-rename machinery, and the wire needs
  `__type` verbatim.
- PCL discriminator narrowing works (converter picked the right variant from the tag),
  so the registry-examples pipeline is viable for phase 2. The `pulumi package gen-sdk`
  local invocation did not run example conversion in this environment (it strips the
  fence without attempting conversion, control included); CI's docs pipeline converts —
  chase the exact env wiring during phase 2, not now.

## Reproduce

```
cd .claude/worktrees/union-poc
python3 poc/patch_schema.py
pulumi package gen-sdk provider/cmd/pulumi-resource-pulumiservice/schema.json \
  --version 1.999.0-alpha.poc --language <lang> --out sdk-poc/
cd poc/examples/<lang> && pulumi preview   # PULUMI_BACKEND_URL=file://... PULUMI_CONFIG_PASSPHRASE=
cd poc/convert && POC_SCHEMA_FILE=.../schema.json pulumi convert --from pcl --language <lang> --out out-<lang>
```
