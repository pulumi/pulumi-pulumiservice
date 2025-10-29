# EnvironmentRotationSchedule Resource Example

This example demonstrates the EnvironmentRotationSchedule resource in the Pulumi Service Provider.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Provider built and installed locally (for local testing)

## What This Example Does

This example creates an EnvironmentRotationSchedule resource that schedules automatic rotation of environment credentials and secrets. Environment rotation schedules help maintain security by regularly rotating sensitive values.

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

The EnvironmentRotationSchedule resource requires:
- **organization**: The Pulumi organization (from environment)
- **project**: The project name (from environment)
- **environment**: The environment name (created by testEnvironment resource)
- **scheduleCron**: Cron expression for schedule timing

The example creates an environment first, then creates a rotation schedule that runs on January 1st at midnight annually.
