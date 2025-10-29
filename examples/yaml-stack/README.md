# Stack Resource Example

This example demonstrates the Stack resource in the Pulumi Service Provider.

## Prerequisites

- Pulumi CLI installed
- Pulumi Cloud account and access token
- Provider built and installed locally (for local testing)

## What This Example Does

This example creates a Stack resource in Pulumi Cloud. A Stack is a fundamental resource that represents an instance of a Pulumi program with its own separate state and configuration.

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

The Stack resource creates a new stack with:
- **organizationName**: The Pulumi organization (from provider config)
- **projectName**: The project name (test-project-{digits})
- **stackName**: The stack name (dev-{digits})

The `${digits}` placeholder ensures unique names to avoid conflicts when running tests.
