# Service Example (YAML)

This example demonstrates how to provision a Service in Pulumi Cloud using the Pulumi Service Provider with YAML.

## Prerequisites

- Pulumi CLI installed
- A Pulumi Cloud account with appropriate permissions
- `PULUMI_ACCESS_TOKEN` environment variable set

## Configuration

The example requires the following configuration:

- `orgName`: The Pulumi organization name
- `ownerName`: The username that will own the service

## Usage

```bash
pulumi config set orgName <your-org-name>
pulumi config set ownerName <your-username>
pulumi up
```

## Converting to Other Languages

You can convert this YAML example to other programming languages using the `pulumi convert` command. See the [Pulumi Convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) for more information.

Example:
```bash
pulumi convert --language typescript --out ./ts-service
```
