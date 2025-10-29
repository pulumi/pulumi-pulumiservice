# TeamEnvironmentPermission Resource Example

This example demonstrates the TeamEnvironmentPermission resource in the Pulumi Service Provider.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Provider built and installed locally (for local testing)

## What This Example Does

This example creates a TeamEnvironmentPermission resource that grants a team specific permissions to access an environment. It demonstrates resource dependencies by creating both a Team and an Environment, then granting the team read permission to the environment.

## Running the Example

The `digits` parameter is automatically set by the test framework to ensure unique resource names.

If running manually:

```bash
pulumi config set digits 12345
pulumi up
```

## Converting to Other Languages

See [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) for converting this YAML example to TypeScript, Python, Go, C#, or Java.

## Resource Details

The TeamEnvironmentPermission resource requires:
- **organization**: The Pulumi organization
- **team**: The team name (created by testTeam resource)
- **environment**: The environment name (created by testEnvironment resource)
- **permission**: Permission level (read, write, or admin)

The example creates all dependent resources (Team and Environment) to demonstrate a complete working scenario.
