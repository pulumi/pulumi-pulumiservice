# v2 examples (hand-maintained per language)

End-to-end tests of the `pulumiservice:v2:*` resource surface. Each example
ships its own per-language program in:

```
examples/v2/<name>/
  yaml/        Pulumi.yaml + Main.yaml
  typescript/  Pulumi.yaml + index.ts + package.json + tsconfig.json
  python/      Pulumi.yaml + __main__.py + requirements.txt
  go/          Pulumi.yaml + main.go + go.mod
  csharp/      Pulumi.yaml + Program.cs + *.csproj
  java/        Pulumi.yaml + pom.xml + src/main/java/...
```

Each language directory is hand-maintained — there is no PCL source nor
code-generation step. We tried PCL with `pulumi convert` first; it
mishandles multi-module schemas (the `v2` module under the `pulumiservice`
package), emitting unquoted Python string literals, missing
`pulumi.String(...)` wrappers in Go, and invalid C# array initializers
for nested object lists. The fix needs to land upstream in
`pulumi/pulumi`'s PCL→language codegen; until then, the canonical
representation lives in the language files directly.

## Coverage

| Example | Resources | Mode |
|---|---|---|
| webhooks            | 2 × OrganizationWebhook | full E2E |
| oidc-issuer         | 2 × OidcIssuer | preview-only — RegisterOidcIssuer is non-idempotent |
| policy-groups       | 1 × PolicyGroup | preview-only — BatchUpdatePolicyGroup requires the group to exist first |
| policy-packs        | 1 × PolicyPack | preview-only — Create returns an upload URL we don't POST against |
| environments        | 1 × Environment_esc + 2 × RevisionTag_esc | preview-only — CreateEnvironment returns empty 200 (no /id) |
| approval-rules      | 1 × PolicyGroup | preview-only — same as policy-groups |
| rbac                | 1 × Role + 1 × Team | preview-only |
| schedules           | 1 × ScheduledDeployment + 1 × EnvironmentSchedule | preview-only — depends on DeploymentSettings idField |
| stack               | 1 × Stack | preview-only — CreateStack response is `{messages}`-only |
| deployment-settings | 1 × DeploymentSettings | preview-only — singleton, no /id in response |
| teams               | 1 × Team | full E2E (CreatePulumiTeam + DeleteTeam) |

The "preview-only" examples still:
- type-check the program in every language,
- resolve `pulumiservice:v2:*` tokens against the in-process provider,
- exercise the v2 Check + Diff path,

…they just stop short of calling Create against the live API because
the create surface for those resources is incomplete in v2 metadata.

## Run

```bash
PULUMI_TEST_OWNER=<your-org> \
PULUMI_BACKEND_URL=<your-pulumi-url> \
PULUMI_ACCESS_TOKEN=<your-token> \
go test -tags v2 -run TestV2 -v -timeout 30m ./examples
```

To exercise just one example or one lane:

```bash
go test -tags v2 -run "TestV2/webhooks/yaml"     -v ./examples
go test -tags v2 -run "TestV2/.+/typescript"     -v ./examples
```

## Local-SDK linkage

The hand-written programs reference `@pulumi/pulumiservice` /
`pulumi_pulumiservice` / the Go SDK module path / `Pulumi.PulumiService`
by name. The test harness patches each program at run time to point
those names at the local checkout's SDK builds:

- **typescript**: `opttest.YarnLink("@pulumi/pulumiservice")` — requires
  `make build_nodejs install_nodejs_sdk`.
- **python**: rewrites `Pulumi.yaml` to `runtime.options.toolchain: uv`
  (so pulumi creates a uv-managed venv that bypasses Homebrew Python's
  PEP 668 lock), runs `pulumi install` ourselves, then `uv pip install
  -e <repo>/sdk/python` to editable-install the local SDK.
- **go**: `go mod edit -require=… -replace=…` against the workspace's
  `sdk/` module before pulumitest's first `go mod tidy`.
- **csharp**: rewrites the rendered `.csproj` to drop the
  `PackageReference Include="Pulumi.PulumiService"` and add a local
  `<ProjectReference>` to `sdk/dotnet/Pulumi.PulumiService.csproj`.
  Also retargets `<TargetFramework>net8.0</TargetFramework>` to
  whatever the host's `dotnet --list-runtimes` reports.
- **java**: pom.xml's `<version>[0.0.0,)</version>` resolves to the
  local snapshot in `~/.m2`, populated by `make build_java` (which now
  runs `gradle publishToMavenLocal`). Requires `mvn` on `PATH` (`brew
  install maven`).

## Java post-generation patches

`make build_java` runs three sed/perl patches on the generated Java SDK
to work around upstream `pulumi-java-gen` bugs (all three are
idempotent on regen — see `Makefile` `.make/generate_java`):

1. Strip `include("lib")` from `settings.gradle` — pulumi-java-gen
   ships it but never creates the `lib/` subproject.
2. Replace the multi-line schema description in `build.gradle` with a
   one-line summary — Groovy can't parse the unescaped newlines.
3. Bump the `JavaLanguageVersion` toolchain pin from 11 to 17 — 11
   isn't preinstalled on macOS Adoptium, 17 is.

Then `make build_java` runs `gradle publishToMavenLocal` so the
example pom.xml's `[0.0.0,)` version range resolves to the local
snapshot.
