# DeploymentSchedule Resource Example

This example demonstrates the DeploymentSchedule resource in the Pulumi Service Provider.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Provider built and installed locally (for local testing)

## What This Example Does

This example creates a DeploymentSchedule resource that schedules automatic deployments of a stack. Deployment schedules allow you to automate stack updates on a cron schedule.

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

The DeploymentSchedule resource requires:
- **organization**: The Pulumi organization
- **project**: The project name
- **stack**: The stack name (must match DeploymentSettings stack)
- **scheduleCron**: Cron expression for schedule timing
- **pulumiOperation**: Operation to perform (update, preview, destroy, refresh)

The example creates DeploymentSettings first (required dependency), then creates a schedule that runs on January 1st at midnight annually.
