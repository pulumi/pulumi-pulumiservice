# Policy Groups Example (YAML)

This example demonstrates:
- Creating a `PolicyGroup` resource
- Using the `getPolicyPacks` data source to list all policy packs in an organization
- Using the `getPolicyPack` data source to get details about a specific policy pack (commented out - requires existing policy pack)

## Policy Pack Provenance

Both data sources return the pack's Pulumi Registry provenance:

- **source**: `pulumi` for packs published by Pulumi (for example `cis-aws`), `private` for packs published by your own organization
- **publisher**: `pulumi` for Pulumi-published packs, otherwise the publishing organization's name

Use these to filter packs by who published them — "apply Pulumi's compliance
packs, leave ours alone" becomes `publisher == "pulumi"` instead of a name-matching
heuristic against Pulumi's `<framework>-<cloud>` convention, which goes stale
whenever Pulumi publishes a pack that doesn't fit the pattern.

Both fields are optional and are omitted when the provider could not determine
registry metadata for a pack, so treat an absent value as unknown rather than
as "not published by Pulumi".

## New Policy Fields (v0.32.0+)

The `getPolicyPack` function now returns enhanced policy metadata including:
- **severity**: Policy severity level (low, medium, high, critical)
- **framework**: Compliance framework details (name, version, reference, specification)
- **tags**: Array of tags associated with the policy
- **remediationSteps**: Description of remediation steps
- **url**: URL to more information about the policy

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
