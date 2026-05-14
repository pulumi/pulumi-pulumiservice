# Stack with Service-Backed Configuration (SBC) Example

This example demonstrates linking a `pulumiservice:Stack` to a stack-managed
ESC environment that holds the stack's configuration and secrets, via the
`configEnvironment` input.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Service-Backed Configuration must be enabled on the test organization
  (see the `39623-service-backed-config` LaunchDarkly flag)

## What This Example Does

With `configEnvironment.managed: true`, the server creates a dedicated ESC
environment named `<projectName>/<stackName>` at stack creation time and
deletes it when the stack is destroyed.

## Running the Example

The `digits` parameter is automatically set by the test framework to ensure
unique resource names.

If running manually:

```bash
pulumi config set digits 12345
pulumi up
```

## Converting to Other Languages

See the
[`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for converting this YAML example to TypeScript, Python, Go, C#, or Java.
