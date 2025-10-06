# YAML Role Example

This example demonstrates how to create and manage custom RBAC roles in Pulumi Cloud using the Pulumi YAML language.

## Prerequisites

- A Pulumi Cloud account with an organization that has the `CustomRoles` feature enabled
- A Pulumi access token set as `PULUMI_ACCESS_TOKEN` environment variable

## Running the Example

```bash
pulumi up
```

## Converting to Other Languages

This YAML example can be converted to other Pulumi-supported languages using the `pulumi convert` command. See the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) for more information.

For example, to convert to TypeScript:

```bash
pulumi convert --from yaml --to typescript --out typescript-role
```

## Clean Up

```bash
pulumi destroy
```
