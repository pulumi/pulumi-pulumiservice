# EnvironmentVersionTag Resource Example

This example demonstrates the EnvironmentVersionTag resource in the Pulumi Service Provider.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Provider built and installed locally (for local testing)

## What This Example Does

This example creates an EnvironmentVersionTag resource that tags a specific revision of an environment. Environment version tags allow you to mark specific versions of environment configurations for easy reference and rollback.

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

The EnvironmentVersionTag resource requires:
- **organization**: The Pulumi organization
- **environment**: The environment name (created by testEnvironment resource)
- **tagName**: The tag name (e.g., "stable", "production")
- **revision**: The environment revision number to tag

The example creates the environment first, then tags revision 1 with the "stable" tag.
