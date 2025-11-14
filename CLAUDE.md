# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the Pulumi Service Provider (PSP), a Pulumi provider built on top of the Pulumi Cloud REST API that allows managing Pulumi Cloud resources (Stacks, Environments, Teams, Tokens, Webhooks, Deployment Settings, etc.) using Pulumi programs.

The provider is written in Go and generates SDKs for multiple languages (Node.js/TypeScript, Python, Go, .NET, Java).

## Repository Structure

- `provider/` - Core provider implementation in Go
  - `provider/cmd/pulumi-resource-pulumiservice/` - Main provider binary and schema.json
  - `provider/pkg/provider/` - Provider server implementation (gRPC)
  - `provider/pkg/pulumiapi/` - HTTP client wrappers for Pulumi Cloud REST API
  - `provider/pkg/resources/` - Resource implementations (teams, stacks, webhooks, etc.)
  - `provider/pkg/util/` - Utility functions for property handling, diffs, secrets
- `sdk/` - Generated SDKs for each language (dotnet, go, java, nodejs, python)
- `examples/` - Example programs in various languages demonstrating provider usage

## Build Commands

### Essential Commands

```bash
# Restore/install build dependencies
make ensure

# Build provider binary only
make provider

# Build all SDKs (requires provider to be built first)
make build_sdks

# Build everything (provider + all SDKs)
make build

# Run provider tests
make test_provider

# Run linting (runs golangci-lint in provider, sdk, and examples directories)
make lint
```

### SDK-Specific Commands

```bash
make nodejs_sdk    # Build Node.js/TypeScript SDK
make python_sdk    # Build Python SDK
make go_sdk        # Build Go SDK
make dotnet_sdk    # Build .NET SDK
make java_sdk      # Build Java SDK
```

### Installation Commands

```bash
make install              # Install provider binary and Node.js/dotnet SDKs
make install_nodejs_sdk   # Link Node.js SDK locally via yarn
make install_java_sdk     # Publish Java SDK to local Maven
```

## Environment Variables

- **Integration Tests**: This project uses `.env` file for local testing credentials
  - `PULUMI_ACCESS_TOKEN`: Token for authenticating with Pulumi Cloud
  - `PULUMI_TEST_OWNER`: Organization name for tests (defaults to `service-provider-test-org` if not set)
  - These variables are loaded automatically by the test framework when present in `.env`
  - The `.env` file is gitignored for security

### Testing

```bash
# Run provider unit tests
cd provider/pkg && go test -short -v -count=1 -cover -timeout 2h -parallel 4 ./...

# Run example tests (integration tests)
cd examples && go test -tags=all -v -count=1 -timeout 3h -parallel 4

# Run specific example test
cd examples && go test -v -run TestYamlStackTagsPluralExample -tags yaml -timeout 10m
```

- Integration tests are located in `examples/` directory
- Tests are tagged: use `-tags yaml`, `-tags nodejs`, `-tags python`, etc.
- The `.env` file allows running integration tests locally against a custom organization
- Every new resource should have unit tests in `provider/pkg/` with `_test.go` suffix
- Add example programs in `examples/` for each new resource (examples serve as integration tests)
- Examples are organized by language: `ts-*`, `py-*`, `go-*`, `cs-*`, `java-*`, `yaml-*`

## Development Workflow

### Copyright Headers

**IMPORTANT**: All new files must have correct copyright year ranges:
- In 2025: Use `// Copyright 2016-2025, Pulumi Corporation.`
- In 2026: Use `// Copyright 2016-2026, Pulumi Corporation.`
- General rule: Always use current year as the end year for new files

### CHANGELOG Updates

**IMPORTANT**: Always update `CHANGELOG.md` when making code changes that affect users:

1. Add an entry under the `## Unreleased` section
2. If `## Unreleased` doesn't exist, create it at the top of the file (after the `# CHANGELOG` header)
3. Categorize changes appropriately:
   - `### Improvements` - New features, enhancements, new resources
   - `### Bug Fixes` - Bug fixes, corrections
   - `### Breaking Changes` - Breaking changes (rare, requires major version bump)
4. Format: `- Description of change [#issue_or_pr_number](link_to_issue_or_pr)`
5. Examples:
   - `- Added StackTags resource for managing multiple stack tags [#61](https://github.com/pulumi/pulumi-pulumiservice/issues/61)`
   - `- Fixed OIDC Issuer policies order to prevent accidental drifts [#542](https://github.com/pulumi/pulumi-pulumiservice/pull/542)`

**EXCEPTION**: **DO NOT** add CHANGELOG entries for:

- Test-only changes (adding tests, fixing tests, test infrastructure)
- Documentation updates (README, CLAUDE.md, code comments)
- CI/build configuration changes
- Development tooling changes

These changes are important but not user-facing, so they don't belong in the
CHANGELOG.

### Schema Changes

The provider schema is **manually maintained** at `provider/pkg/provider/manual-schema.json`. When adding/modifying resources:

1. Update `manual-schema.json` with the new resource/property definitions
2. Run `make provider` to regenerate the provider binary (uses `go generate`)
3. Run `make build_sdks` to regenerate all language SDKs from the schema
4. The schema is embedded into the provider binary at build time

#### Schema JSON Syntax Guidelines

**Important**: The schema JSON must follow strict syntax rules. Common patterns:

**Complex Nested Objects**:
- ❌ **NEVER** define inline objects directly in array `items`:
  ```json
  "items": {
    "type": "object",
    "properties": { ... }
  }
  ```
- ✅ **ALWAYS** define complex objects as separate types in the `types` section and reference them:
  ```json
  "items": {
    "$ref": "#/types/pulumiservice:index:YourTypeName"
  }
  ```

**Additional Properties**:
- ❌ **NEVER** use boolean values: `"additionalProperties": true`
- ✅ **ALWAYS** use type specifications:
  ```json
  "additionalProperties": {
    "$ref": "pulumi.json#/Any"
  }
  ```
  or
  ```json
  "additionalProperties": {
    "type": "string"
  }
  ```

**Type References**:
- Use `"$ref": "#/types/pulumiservice:index:TypeName"` for custom types
- Use `"$ref": "pulumi.json#/Any"` for generic object types
- Custom types must be defined in the `types` section with full descriptions

### Adding a New Resource

1. Add resource definition to `provider/pkg/provider/manual-schema.json`
2. Create API client methods in `provider/pkg/pulumiapi/` (e.g., `teams.go`, `webhooks.go`)
3. Create resource implementation in `provider/pkg/resources/` implementing `PulumiServiceResource` interface
4. Register the resource in `provider/pkg/provider/provider.go`
5. Rebuild provider and SDKs: `make build`
6. Add examples in `examples/` directory
7. **IMPORTANT**: Always create a YAML example for the new resource:
   - Add a `yaml-*` example directory in `examples/`
   - Create a `README.md` in the example directory that includes a link to the `pulumi convert` documentation (https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) for converting the example to other programming languages
   - Register the example test in `examples/examples_yaml_test.go`
   - Test the YAML example before completing: `cd examples && go test -v -run TestYaml<ResourceName>Example -tags yaml -timeout 10m`

### Resource Interface

All resources must implement the `PulumiServiceResource` interface:

```go
type PulumiServiceResource interface {
    Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error)
    Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error)
    Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error)
    Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error)
    Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error)
    Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error)
    Name() string
}
```

## API Client Architecture

The `pulumiapi` package provides HTTP client wrappers for the Pulumi Cloud REST API:

- `Client` - Base HTTP client with authentication and standard headers
- Individual files for each resource type (e.g., `teams.go`, `stack.go`, `webhooks.go`)
- All API calls use the standard headers: `X-Pulumi-Source: provider`, `Accept: application/vnd.pulumi+8`
- Authentication via Bearer token in `Authorization` header

## Release Process

Releases are handled by Pulumi employees via `#release-ops` Slack channel. GitHub Actions automatically builds, tests, and publishes new releases to all package managers.

## Configuration

Provider accepts two configuration options:
- `accessToken` (env: `PULUMI_ACCESS_TOKEN`) - Pulumi Service access token
- `apiUrl` (env: `PULUMI_BACKEND_URL`) - Custom API URL for self-hosted instances

## Version Management

- Provider version is set via `PROVIDER_VERSION` environment variable or defaults to `1.0.0-alpha.0+dev`
- Version is injected into provider binary via LDFLAGS at build time
- The Pulumi CLI binary version used for SDK generation is automatically synced with the `github.com/pulumi/pulumi/pkg/v3` dependency version

## Linting

The Makefile references `.golangci.yml` but it may not exist in the repository root. The `make lint` command runs golangci-lint in three directories: `provider`, `sdk`, and `examples` with a 10-minute timeout.

**Important**: When linting the `examples/` directory, you must use the `--build-tags all` flag because the test files use build tags (yaml, nodejs, python, etc.):

```bash
cd examples && golangci-lint run --timeout 10m --build-tags all
```

Without the build tags, golangci-lint may report false positives about unused functions that are actually used in files with specific build tags.

## Debugging CI Failures

When investigating CI test failures:

1. **Use GitHub CLI to fetch logs**:
   ```bash
   gh run view <run_id> --log-failed
   gh api repos/pulumi/pulumi-pulumiservice/actions/jobs/<job_id>/logs | grep "error:"
   ```

2. **Check for API validation errors**: The Pulumi Cloud API may reject values that are defined in the schema but not yet implemented (e.g., "runner" token type in OIDC policies)

3. **Check for outdated test data**: Examples may contain hardcoded values like TLS certificate thumbprints that become stale when certificates are rotated

4. **Look for related issues across examples**: If one language example fails, check all language examples (ts-, py-, go-, cs-, yaml-) for the same issue

## Lessons Learned

### Test Coverage and Example Management

**ALWAYS audit existing test coverage before creating new tests or examples:**

1. **Check for composite examples first**: Some examples intentionally test multiple related resources together:
   - `yaml-schedules` tests DeploymentSchedule, DriftSchedule, TtlSchedule, and EnvironmentRotationSchedule
   - `yaml-environments` tests Environment, EnvironmentVersionTag, and TeamEnvironmentPermission
   - `yaml-deployment-settings*` tests different DeploymentSettings configurations
   - Don't create duplicate examples for resources already covered in composite examples

2. **Understand the intent behind composite examples**: They exist to:
   - Show how related resources work together
   - Reduce maintenance burden (one example to update vs. many)
   - Test realistic usage patterns (resources are rarely used in isolation)
   - Avoid test duplication and longer CI times

3. **When to create a new example directory**:
   - ✅ Resource has NO existing example coverage
   - ✅ Resource represents a distinct use case not covered by composite examples
   - ❌ Resource is already tested in a composite example
   - ❌ Creating individual examples would duplicate existing coverage

4. **How to add test coverage for already-covered resources**:
   - Add test functions that point to existing composite examples
   - Example: Add `TestYamlDriftSchedule` that uses `yaml-schedules` directory
   - Don't create `yaml-drift-schedule/` if `yaml-schedules` already tests it

5. **Review existing test files**:
   - Check `examples/examples_yaml_test.go` for existing YAML tests
   - Check other language test files (nodejs, python, go, dotnet, java)
   - Use `grep -r "TestYaml.*Schedule" examples/` to find related tests

### CHANGELOG Best Practices

**DO NOT add CHANGELOG entries for test-only changes:**

- ❌ "Added YAML integration tests for 9 resources"
- ❌ "Fixed flaky test in TeamEnvironmentPermission"
- ❌ "Migrated tests from integration.ProgramTest to pulumitest"
- ✅ "Added StackTags resource for managing multiple stack tags" (new user-facing feature)
- ✅ "Fixed drift in OIDC Issuer policies order" (user-facing bug fix)

**Rationale**: Tests are infrastructure for maintainers, not features for users. The CHANGELOG documents user-facing changes only.

### Interpreting Issue Requirements

**Don't take issue descriptions literally without understanding context:**

1. **"Create dedicated example"** might mean:
   - "Add a dedicated test function" (not a new directory)
   - "Add focused documentation" (not a separate example)
   - "Extract into separate example" (only if current example is too complex)

2. **Always ask**: "Is this adding value or just duplication?"
   - If the resource is already tested elsewhere, link to it instead
   - If the example is already comprehensive, reference it instead
   - If creating work that will need ongoing maintenance, question if it's necessary

3. **Check with the team**: If an issue seems to request duplication, ask for clarification before implementing

### Code Review Feedback

**When reviewers question changes, they're often seeing duplication you missed:**

- "How is this different from X?" usually means "This duplicates X, please remove it"
- "Why are we changing Y?" usually means "This change seems unnecessary"
- Don't be defensive - reviewers have broader context and catch things you missed

**Respond to review feedback by**:

1. Acknowledging the feedback
2. Analyzing why you made the change
3. Fixing the issue (removing duplicates, reverting unnecessary changes)
4. Learning the pattern to avoid repeating the mistake

## Unit Testing Best Practices

### What to Unit Test

**Good candidates for unit testing:**
- ✅ Pure transformation functions (data structure conversions like `ToPropertyMap()`)
- ✅ Stateless logic (diff calculations, validation like `Diff()`, `Check()`)
- ✅ Simple methods with no external dependencies (like `Name()`)
- ✅ Input parsing and validation logic

**Poor candidates for unit testing:**
- ❌ CRUD operations requiring HTTP clients (test these as integration tests in `examples/`)
- ❌ Methods that primarily delegate to external services
- ❌ Complex orchestration logic across multiple services
- ❌ Anything requiring a real `pulumiapi.Client` instance

**Rationale**: Unit tests should be fast, isolated, and test logic you own. Integration tests in `examples/` already provide coverage for end-to-end workflows.

### Mocking Guidelines

**Don't mock what you don't own:**
- If a dependency doesn't provide an interface, it's a signal that unit testing at that level may not be appropriate
- The `pulumiapi.Client` is a concrete struct - don't try to mock it in resource tests
- Client/HTTP testing belongs in the `pulumiapi/` package, not in resource tests
- Resource tests should focus on resource-specific logic only (transformations, diffs, validation)

**Avoid over-engineering:**
- ❌ No helper functions that just wrap struct literals (e.g., `buildMock(...)` that returns `&Mock{...}`)
- ❌ No complex mocking infrastructure with embedding and type conversions
- ❌ If you need `unsafe.Pointer` conversions, you're going the wrong way
- ✅ Inline mock creation when actually needed
- ✅ Keep test setup simple and obvious

### Test Quality Indicators

**Signs of good tests:**
- ✅ Each test has a distinct, clear purpose
- ✅ Table-driven tests for variations on the same behavior
- ✅ Simple setup with inline test data, clear assertions
- ✅ Fast execution (< 1s for entire test suite)
- ✅ Tests focus on behavior, not implementation details
- ✅ No redundant test cases

**Signs of "AI slop" or over-engineered tests:**
- ❌ Redundant test cases testing the same thing multiple ways
- ❌ Unused test structure fields (e.g., `expectError bool` when method never errors)
- ❌ Helper functions that don't actually reduce complexity
- ❌ Over-complicated extraction/assertion logic
- ❌ Tests that duplicate what integration tests already cover
- ❌ Complex mocking that fights the type system

### Example: Stack Resource Tests

See `provider/pkg/resources/stack_test.go` for a well-architected test file that follows these principles:

**What it tests:**
- `Name()` - simple getter
- `ToPropertyMap()` - pure transformation
- `ToPulumiServiceStackTagInput()` - input parsing
- `Diff()` - diff calculation logic
- `Check()` - input validation
- `Update()` - ensures it errors (stacks are immutable)

**What it doesn't test:**
- Create/Read/Delete operations (require real `pulumiapi.Client`)
- HTTP client behavior (belongs in `pulumiapi/` package)
- End-to-end workflows (covered by integration tests in `examples/`)

**Results:**
- 6 test functions, 17 test cases
- < 50ms execution time
- No mocking complexity
- No external dependencies
- Clear, maintainable code

### Common Pitfalls

**Trying to mock `pulumiapi.Client`:**
- The Client is a concrete struct, not an interface
- Don't try embedding, type conversion tricks, or unsafe pointers
- If you need the client, you need an integration test

**Testing the wrong thing:**
- Don't test that HTTP calls work (that's the client's job)
- Don't test Pulumi SDK behavior (trust the SDK)
- Test YOUR logic: transformations, validation, diff calculation

**Over-complicated test structures:**
```go
// ❌ Bad: Unused fields that make tests confusing
tests := []struct {
    name          string
    input         Foo
    expected      Bar
    expectError   bool  // Never true in any test case
    errorContains string // Never used
}{...}

// ✅ Good: Only the fields you actually use
tests := []struct {
    name     string
    input    Foo
    expected Bar
}{...}
```

### When to Skip Unit Tests

It's OK to not have unit tests for:
- Simple CRUD resources with no transformation logic
- Resources that are thin wrappers around API calls
- Code that's already covered by integration tests

**Focus unit testing effort on:**
- Complex validation logic
- Non-trivial transformations
- Diff calculation algorithms
- Edge cases that are hard to reproduce in integration tests
