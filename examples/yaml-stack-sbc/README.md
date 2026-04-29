# Stack with Service-Backed Configuration (SBC) Example

This example demonstrates linking a `pulumiservice:Stack` to an ESC environment
that holds the stack's configuration and secrets, via the `configEnvironment`
input.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Service-Backed Configuration must be enabled on the test organization
  (see the `39623-service-backed-config` LaunchDarkly flag)

## What This Example Does

The example creates two stacks that demonstrate the two SBC modes:

1. **Reference an existing env** (`refStack`): the program also creates a
   sibling `pulumiservice:Environment` (`sharedCfg`) and links the stack to it
   via `configEnvironment.project` + `configEnvironment.environment`. PSP does
   not own the env's lifecycle — on stack delete the linked env is preserved.
2. **Stack-managed env** (`autoStack`): with `configEnvironment.auto: true`,
   the server creates a dedicated env named `<projectName>/<stackName>` at
   stack creation time and deletes it when the stack is destroyed.

`refStack` also pins to a specific revision via `configEnvironment.version:
${sharedCfg.revision}`, so editing the env doesn't silently shift the stack's
config until the program is re-applied.

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
