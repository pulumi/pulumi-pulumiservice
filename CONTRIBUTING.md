# Contributing to Pulumi Service Provider

First, thanks for contributing to Pulumi Service Provider and helping make it better. We appreciate the help!
This repository is one of many across the Pulumi ecosystem and we welcome contributions to them all.

## Code of Conduct

Please make sure to read and observe our [Contributor Code of Conduct](./CODE-OF-CONDUCT.md).

## Communications

You are welcome to join the [Pulumi Community Slack](https://slack.pulumi.com/) for questions and a community of like-minded folks.
We discuss features and file bugs on GitHub via [Issues](https://github.com/pulumi/pulumi-pulumiservice/issues) as well as [Discussions](https://github.com/pulumi/pulumi-pulumiservice/discussions).

### Issues

Feel free to pick up any existing issue that looks interesting to you or fix a bug you stumble across while using Pulumi Service Provider. No matter the size, we welcome all improvements.

### Feature Work

For larger features, we'd appreciate it if you open a [new issue](https://github.com/pulumi/pulumi-pulumiservice/issues/new) before investing a lot of time so we can discuss the feature together.
Please also be sure to browse [current issues](https://github.com/pulumi/pulumi-pulumiservice/issues) to make sure your issue is unique, to lighten the triage burden on our maintainers.
Finally, please limit your pull requests to contain only one feature at a time. Separating feature work into individual pull requests helps speed up code review and reduces the barrier to merge.

## Developing

Here's a quick list of helpful make commands:

1. `make ensure`, which restores/installs any build dependencies
1. `make build`, which generates models from provider's `schema.json`, builds the provider and builds all SDKs into the `sdk` folder
1. `make install`, which installs Pulumi Service Provider

## Testing

This provider uses a comprehensive testing strategy with unit tests and integration tests. Every change should include appropriate tests, and every new resource must have both unit tests and integration test examples.

For detailed testing guidance, see the sections below.

### Overview of Testing Strategy

We maintain two layers of testing:

1. **Unit Tests** - Fast, isolated tests using mocks for resource implementations
2. **Integration Tests** - End-to-end tests using real Pulumi programs against the Pulumi Cloud API

Our goals:

- Fast feedback during development (unit tests run in seconds)
- High confidence in production behavior (integration tests verify real API interactions)
- Prevent regression (all bugs get tests to prevent recurrence)
- Consistent patterns across all resources

### Test Types

#### Unit Tests

- **Location**: `provider/pkg/resources/*_test.go` and `provider/pkg/pulumiapi/*_test.go`
- **Purpose**: Test resource implementations and API client methods in isolation
- **Speed**: Fast (<10 seconds per resource)
- **Dependencies**: None (uses mocks)
- **Coverage**: Individual CRUD operations, error handling, property validation
- **Run Command**:

  ```bash
  cd provider/pkg && go test -short -v -count=1 -cover -timeout 2h -parallel 4 ./...
  ```

#### Integration Tests

- **Location**: `examples/examples_yaml_test.go`
- **Purpose**: Test end-to-end functionality against real Pulumi Cloud API
- **Speed**: Slow (3 hours total for full suite, tests run in parallel)
- **Dependencies**: `PULUMI_ACCESS_TOKEN`, test organization access
- **Coverage**: Full resource lifecycle, idempotency, updates, imports
- **Run Command**:

  ```bash
  cd examples && go test -tags=yaml -v -count=1 -timeout 3h -parallel 4
  ```

- **Run Single Test**:

  ```bash
  cd examples && go test -v -run TestYamlTeamsExample -tags yaml -timeout 10m
  ```

#### Regression Tests

- **Purpose**: Prevent known bugs from recurring
- **Location**: Embedded in unit tests or integration tests
- **Marked with**: Comment referencing the original issue (e.g., `// Regression test for #123`)
- **Example**: See `TestYamlTeamsExample` which includes a regression test for issue #73

### Writing Unit Tests

Every new resource must include unit tests covering all CRUD operations.

#### Required Test Cases

For each resource, implement tests for:

1. **Read** - Resource found, not found, API error
2. **Create** - Success, validation errors, API errors
3. **Update** - Success, no-op (no changes), API errors
4. **Delete** - Success, resource not found, API errors
5. **Diff** - No changes detected, changes detected
6. **Check** - Valid inputs, invalid inputs

#### Pattern 1: Mock-Based Testing

Use mocks to isolate the resource implementation from API calls:

```go
package resources

import (
    "context"
    "testing"

    "github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
    pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
    "github.com/stretchr/testify/assert"
)

// Define mock client interface
type getTeamFunc func() (*pulumiapi.Team, error)

type TeamClientMock struct {
    getTeamFunc getTeamFunc
}

func (c *TeamClientMock) GetTeam(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error) {
    return c.getTeamFunc()
}

// Implement other interface methods as needed...

func TestTeamRead(t *testing.T) {
    t.Run("Read when the resource is not found", func(t *testing.T) {
        mockedClient := &TeamClientMock{
            getTeamFunc: func() (*pulumiapi.Team, error) {
                return nil, nil // Not found
            },
        }

        provider := PulumiServiceTeamResource{
            Client: mockedClient,
        }

        req := &pulumirpc.ReadRequest{
            Id:  "abc/123",
            Urn: "urn:123",
        }

        resp, err := provider.Read(req)

        assert.NoError(t, err)
        assert.Equal(t, "", resp.Id)
        assert.Nil(t, resp.Properties)
    })

    t.Run("Read when the resource is found", func(t *testing.T) {
        mockedClient := &TeamClientMock{
            getTeamFunc: func() (*pulumiapi.Team, error) {
                return &pulumiapi.Team{
                    Name:        "test-team",
                    DisplayName: "Test Team",
                }, nil
            },
        }

        provider := PulumiServiceTeamResource{
            Client: mockedClient,
        }

        req := &pulumirpc.ReadRequest{
            Id:  "org/test-team",
            Urn: "urn:123",
        }

        resp, err := provider.Read(req)

        assert.NoError(t, err)
        assert.Equal(t, "org/test-team", resp.Id)
        assert.NotNil(t, resp.Properties)
    })
}
```

#### Pattern 2: Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name        string
        input       map[string]interface{}
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid team name",
            input: map[string]interface{}{
                "name":             "valid-team",
                "organizationName": "my-org",
            },
            expectError: false,
        },
        {
            name: "missing organization name",
            input: map[string]interface{}{
                "name": "valid-team",
            },
            expectError: true,
            errorMsg:    "organizationName is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateInput(tt.input)
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### Best Practices for Unit Tests

- Use descriptive test names that explain what is being tested
- Test error cases thoroughly (not just happy paths)
- Use `t.Run()` subtests for organization
- Mock external dependencies (API clients)
- Test edge cases (empty strings, nil values, etc.)
- Never make real API calls in unit tests
- Never commit hardcoded tokens or credentials

### Writing Integration Tests

Integration tests use YAML examples with the `integration.ProgramTest` framework from Pulumi.

#### Important: YAML Only

**We only write integration tests for YAML examples.** This is a deliberate choice to:

- Reduce test execution time (one language instead of five)
- Simplify maintenance (fewer test files to update)
- Provide clear, language-agnostic examples
- Leverage Pulumi's `pulumi convert` for other languages

Users can convert YAML examples to their preferred language using:

```bash
pulumi convert --language python --from yaml --out ./python-example
```

#### Required: YAML Example for Every Resource

When adding a new resource, you **must** create a YAML example:

1. Create a `yaml-<resource>` directory in `examples/`
2. Add a `Pulumi.yaml` file with the resource configuration
3. Create a `README.md` that links to the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
4. Register the test in `examples/examples_yaml_test.go`

Example `README.md`:

```markdown
# YAML Team Example

This example demonstrates how to create and manage a Team resource.

## Converting to Other Languages

You can convert this YAML example to your preferred language using:

\`\`\`bash
pulumi convert --language python --from yaml --out ./python-example
\`\`\`

See the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) for more details.
```

#### Using the integration.ProgramTest Framework

```go
//go:build yaml || all
// +build yaml all

package examples

import (
    "testing"

    "github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestYamlTeamExample(t *testing.T) {
    integration.ProgramTest(t, &integration.ProgramTestOptions{
        Dir:   "yaml-team",
        Quick: true,
        ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
            // Validate outputs
            require.NotNil(t, stack.Outputs["teamName"])
            assert.Equal(t, "my-team", stack.Outputs["teamName"])
        },
    })
}
```

#### Test Lifecycle

A typical integration test follows this flow:

1. **Setup**: Create resources via `pulumi up`
2. **Verify**: Check outputs using `ExtraRuntimeValidation`
3. **Update** (optional): Modify resources and run `pulumi up` again via `EditDirs`
4. **Verify**: Check that updates are idempotent
5. **Cleanup**: Destroy resources via `pulumi destroy`

Example with updates:

```go
func TestYamlResourceUpdate(t *testing.T) {
    integration.ProgramTest(t, &integration.ProgramTestOptions{
        Dir:   "yaml-resource-initial",
        Quick: true,
        EditDirs: []integration.EditDir{
            {
                Dir:      "yaml-resource-updated",
                Additive: true, // Overlay on top of previous
                ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
                    // Verify updated state
                },
            },
        },
    })
}
```

### Local Testing Setup

#### Environment Variables

Create a `.env` file in the project root (gitignored) with:

```bash
PULUMI_ACCESS_TOKEN=pul-xxxxxxxxxxxxx
PULUMI_TEST_OWNER=your-test-org
```

- `PULUMI_ACCESS_TOKEN`: Personal access token for Pulumi Cloud
- `PULUMI_TEST_OWNER`: Organization name for tests (defaults to `service-provider-test-org` if not set)

The test framework automatically loads these variables when the `.env` file is present.

#### Running Tests Locally

```bash
# Run all unit tests
cd provider/pkg && go test -short -v -count=1 -cover -timeout 2h -parallel 4 ./...

# Run all integration tests (requires .env setup)
cd examples && go test -tags=yaml -v -count=1 -timeout 3h -parallel 4

# Run specific integration test
cd examples && go test -v -run TestYamlTeamExample -tags yaml -timeout 10m

# Run tests for specific resource
cd provider/pkg/resources && go test -v -run TestTeam
```

### Debugging Tests

#### Common Issues

##### Unit tests fail with "unexpected method call"

- **Cause**: Mock is missing implementation for a method
- **Solution**: Implement all interface methods in your mock, even if they just return nil

##### Integration test times out

- **Cause**: Resource creation is stuck or very slow
- **Solution**: Check Pulumi Cloud console for stack state, increase timeout, or check for API rate limiting

##### "Resource already exists" error

- **Cause**: Previous test run didn't clean up properly
- **Solution**: Manually destroy the test stack: `pulumi stack select <stack> && pulumi destroy`

##### Environment variables not loaded

- **Cause**: `.env` file not in project root or malformed
- **Solution**: Verify `.env` file location and format (no quotes around values)

#### Getting Test Output

```bash
# Verbose output
go test -v

# With test coverage
go test -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test with extra logging
go test -v -run TestTeam -count=1
```

#### Debugging Integration Tests

```bash
# Set Pulumi log level
export PULUMI_LOG_LEVEL=debug

# Keep test stack for inspection
cd examples && go test -v -run TestYamlTeamExample -tags yaml -timeout 10m

# After test, inspect stack in Pulumi Cloud console
```

### CI/CD Pipeline

#### Test Stages

Our CI pipeline runs tests in two stages:

1. **Unit Tests** (~2 minutes)
   - Runs on every PR
   - Must pass before merge
   - Fast feedback for development

2. **Integration Tests** (~3 hours)
   - Runs on every PR
   - Tests are sharded across 6 parallel jobs
   - Each shard runs a subset of examples

#### Test Sharding

Integration tests are divided into 6 shards to reduce total execution time:

- Tests are distributed based on test name hash
- Each shard runs independently in parallel
- Sharding is automatic via `go test` parallelism and CI matrix strategy

#### Flaky Tests

If a test fails intermittently:

1. **Investigate root cause**: Is it timing-related? API throttling? Race condition?
2. **Add retries if appropriate**: Use `integration.RetryOn` for known transient failures
3. **Increase timeout if needed**: Some resources take longer to provision
4. **Report the issue**: Create a GitHub issue with failure logs

**DO NOT** skip flaky tests without investigation.

### Manual Testing

You should also test changes manually using a Pulumi program that uses the updated SDKs.

#### Testing with Local SDK

**Node.js/TypeScript:**

```bash
# Build and install SDK
make install_nodejs_sdk

# In your test program
npm link @pulumi/pulumiservice
```

**Python:**

```bash
# Build SDK
make python_sdk

# In your test program
pip install -e /path/to/pulumi-pulumiservice/sdk/python
```

**Go:**

```bash
# Build SDK
make go_sdk

# In your test program, use replace directive in go.mod
replace github.com/pulumi/pulumi-pulumiservice/sdk/go/v2 => /path/to/pulumi-pulumiservice/sdk/go
```

**.NET:**

```bash
# Build SDK
make dotnet_sdk

# In your test program
dotnet add package Pulumi.PulumiService -s /path/to/pulumi-pulumiservice/sdk/dotnet/bin/Debug/ -v X.XX.XX
```

**Java:**

```bash
# Build and install SDK locally
make install_java_sdk

# SDK will be available in local Maven repository
```

### Best Practices Summary

- Write unit tests for every resource implementation
- Create YAML integration tests for every new resource (not TypeScript, Python, etc.)
- Test error cases and edge conditions
- Use descriptive test names
- Mock external dependencies in unit tests
- Add regression tests for bug fixes with issue references
- Never commit `.env` files or hardcoded credentials
- Never make real API calls in unit tests
- Never ignore flaky tests without investigation

### Reference Examples

- **Unit test example**: `provider/pkg/resources/team_test.go`
- **Integration test example**: `examples/examples_yaml_test.go`
- **Mock pattern**: `provider/pkg/resources/team_test.go`
- **Testing epic**: [#572](https://github.com/pulumi/pulumi-pulumiservice/issues/572)

## Submitting a Pull Request

For contributors we use the [standard fork based workflow](https://gist.github.com/Chaser324/ce0505fbed06b947d962): Fork this repository, create a topic branch, and when ready, open a pull request from your fork.

We require a changelog entry for pretty much all PRs. Add a line in `CHANGELOG.md` describing your change and link to an issue. In rare cases where your PR is a minor change, like formatting or a typo fix, apply `impact/no-changelog-required` label to your PR instead.

### Pulumi employees

Pulumi employees have write access to Pulumi repositories and should push directly to branches rather than forking the repository. Tests can run directly without approval for PRs based on branches rather than forks.

Please ensure that you nest your branches under a unique identifier such as your name (e.g. `pulumipus/cool_feature`).

## Creating a Release

This section is for Pulumi employees only.

To release a new version of the provider, follow steps below:

- Trigger a release in #release-ops
- Github Actions will automatically build, test and then publish the new release to all the various package managers
- Once that is done, you will see your version in [Releases](https://github.com/pulumi/pulumi-pulumiservice/releases)

## Getting Help

We're sure there are rough edges and we appreciate you helping out. If you want to talk with other folks in the Pulumi community (including members of the Pulumi team) come hang out in the `#contribute` channel on the [Pulumi Community Slack](https://slack.pulumi.com/).
