# Policy Groups Example (YAML)

This example demonstrates:
- Creating a `PolicyGroup` resource
- Using the `getPolicyPacks` data source to list all policy packs in an organization
- Using the `getPolicyPack` data source to get details about a specific policy pack (if available)

## Prerequisites

- Pulumi CLI installed
- Pulumi Service access token set in `PULUMI_ACCESS_TOKEN` environment variable
- An organization in Pulumi Cloud (defaults to `service-provider-test-org`)

## Running the Example

```bash
pulumi config set digits $(date +%s)
pulumi up
```

## Converting to Other Languages

You can convert this YAML example to other programming languages using the Pulumi conversion tool:

```bash
pulumi convert --language typescript --out ../ts-policy-groups-converted
pulumi convert --language python --out ../py-policy-groups-converted
pulumi convert --language go --out ../go-policy-groups-converted
```

For more information, see the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
