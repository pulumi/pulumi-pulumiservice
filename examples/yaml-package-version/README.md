# Publishing a Package Version to Pulumi Registry (YAML)

This example demonstrates how to publish a package version to the Pulumi Registry using the Pulumi Service Provider.

## Prerequisites

- A Pulumi Cloud account and access token
- The package artifacts (schema, index, installation configuration)

## Usage

This example shows how to publish a component package to your organization's private registry.

```yaml
resources:
  example-package:
    type: pulumiservice:index:PackageVersion
    properties:
      source: pulumi
      publisher: my-org
      name: my-component
      version: 1.0.0
      schemaContent: |
        { ... package schema JSON ... }
      indexContent: |
        { ... package index JSON ... }
      installationConfigContent: |
        { ... installation config JSON ... }
      packageStatus: ga
      visibility: private
```

## Converting to Other Languages

To convert this example to other Pulumi-supported languages, use the `pulumi convert` command. For more information, see the [Pulumi Convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).

## Running the Example

1. Set your Pulumi access token:
   ```bash
   export PULUMI_ACCESS_TOKEN=<your-token>
   ```

2. Deploy the stack:
   ```bash
   pulumi up
   ```

## Notes

- The `source` should typically be `"pulumi"` for Pulumi-managed packages
- The `publisher` is usually your organization name
- Package versions are immutable - once published, they cannot be changed
- The schema, index, and installation configuration files should contain valid JSON
