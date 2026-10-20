# yaml-organization-role

End-to-end test of `pulumiservice:OrganizationRole` built with the RBAC helper
functions. It creates the stack and environment it grants access to, and
assigns the role to a team:

- **`pulumiservice:buildRolePermissions`** builds the role's permissions from
  organization-level scopes and entity rules (all, by ID, or by tag).
- **`pulumiservice:getStack`** looks up a stack's unique ID.
- **`pulumiservice:TeamRoleAssignment`** assigns the role to a team.

The `OrganizationRole` docs example lives in
`provider/pkg/resources/docs/organization_role_example`; it covers the same
features against existing entities.

## Prerequisites

- The organization must have the Custom Roles feature enabled.

## Convert to another language

Use `pulumi convert` to translate this program to TypeScript, Python, Go,
C#, or Java:

```bash
pulumi convert --from yaml --language typescript --out ../ts-organization-role
```

See the [`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for more options.
