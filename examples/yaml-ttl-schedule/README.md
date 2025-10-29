# TtlSchedule Resource Example

This example demonstrates the TtlSchedule resource in the Pulumi Service Provider.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Provider built and installed locally (for local testing)

## What This Example Does

This example creates a TtlSchedule (Time-To-Live schedule) resource that automatically destroys a stack at a specified time. TTL schedules are useful for temporary environments that should be automatically cleaned up after a certain date.

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

The TtlSchedule resource requires:
- **organization**: The Pulumi organization
- **project**: The project name
- **stack**: The stack name (must match DeploymentSettings stack)
- **timestamp**: ISO 8601 timestamp when the stack should be destroyed
- **deleteAfterDestroy**: Whether to delete the stack after destroying resources (default: false)

The example creates DeploymentSettings first (required dependency), then creates a TTL schedule set to destroy the stack on December 31, 2026.
