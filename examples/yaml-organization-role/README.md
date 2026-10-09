# yaml-organization-role

End-to-end test of `pulumiservice:OrganizationRole` built with the RBAC helper
functions. It creates the stack and environment it grants access to, and
assigns both roles to a team:

- **`pulumiservice:buildRolePermissions`** builds a role's permissions from
  organization-level access and entity rules (all, by ID, or by tag).
- **`pulumiservice:getOrganizationPermissionSet`** looks up built-in permission
  sets. Pulumi Cloud accepts them only in a policy, so the second role stores
  them in a `pulumiservice:api:Role` policy and grants it with
  **`pulumiservice:buildComposePermissions`**.
- **`pulumiservice:getStack`** looks up a stack's unique ID.
- **`pulumiservice:buildEnvironmentScopedPermissions`** is folded in through
  `additionalEntries`.
- **`pulumiservice:TeamRoleAssignment`** assigns the roles to a team.

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
