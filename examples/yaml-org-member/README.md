# Organization Member Example (YAML)

This example demonstrates managing Pulumi Cloud organization members using the Pulumi Service Provider.

## Overview

The example creates an organization member with the "member" role. You can change the role to "admin" if needed.

## Prerequisites

- A Pulumi Cloud organization
- A valid Pulumi user account (the user must already exist in Pulumi Cloud)
- `PULUMI_ACCESS_TOKEN` environment variable set

## Usage

```bash
# Preview the changes
pulumi preview

# Apply the changes
pulumi up

# Remove the member
pulumi destroy
```

## Converting to Other Languages

To convert this example to other programming languages (TypeScript, Python, Go, C#, Java), use:

https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/

Example:
```bash
pulumi convert --language typescript --out ts-org-member
```

## Properties

- `organizationName`: The name of your Pulumi organization
- `userName`: The username of the member to add (must be an existing Pulumi user)
- `role`: Either "admin" or "member"

## Outputs

- `memberId`: The resource ID (format: `organization/userName`)
- `memberRole`: The assigned role
- `knownToPulumi`: Whether the user is known to Pulumi Cloud
