# Neo Task Example

This example demonstrates how to create a Neo task using the Pulumi Service provider with YAML.

## Prerequisites

- Pulumi CLI
- Pulumi Cloud account with access to Neo agent system
- Valid Pulumi access token

## Running the Example

1. Set your Pulumi organization name:
   ```bash
   pulumi config set pulumiservice:organizationName <your-org-name>
   ```

2. Deploy the stack:
   ```bash
   pulumi up
   ```

## Converting to Other Languages

You can convert this YAML example to other programming languages using the `pulumi convert` command. See the [Pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) for more details.

For example, to convert to TypeScript:
```bash
pulumi convert --language typescript --out ../ts-tasks
```
