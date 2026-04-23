# yaml-rbac

Demonstrates the Pulumi Cloud RBAC resources:

- **`pulumiservice:OrganizationRole`** – define a custom fine-grained role.
- **`pulumiservice:OrganizationMember`** – manage org membership and assign a
  role to a user.
- **`pulumiservice:TeamRoleAssignment`** – assign a custom role to a team.
- **`pulumiservice:getOrganizationMembers`** – data source listing every
  organization member and their assigned role.
- **`pulumiservice:getOrganizationRoleScopes`** – data source listing the
  permission scopes available when defining a custom role.

## Prerequisites

- The organization must have the Custom Roles feature enabled.
- `memberUsername` must already have a Pulumi Cloud account.

## Convert to another language

Use `pulumi convert` to translate this program to TypeScript, Python, Go,
C#, or Java:

```bash
pulumi convert --from yaml --language typescript --out ../ts-rbac
```

See the [`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for more options.
