# yaml-organization-role-permission-sets

End-to-end test of granting built-in permission sets in a
`pulumiservice:OrganizationRole`:

- **`pulumiservice:getStackPermissionSet`**, **`getEnvironmentPermissionSet`**,
  **`getAccountPermissionSet`** and **`getOrganizationPermissionSet`** look up
  a built-in permission set's scopes.
- **`pulumiservice:buildRolePermissions`** grants them through
  `permissionSetScopes` and `organizationPermissionSetScopes`.

`yaml-organization-role` covers the rest of `buildRolePermissions`.

## Prerequisites

- The organization must have the Custom Roles feature enabled.

## Convert to another language

Use `pulumi convert` to translate this program to TypeScript, Python, Go,
C#, or Java:

```bash
pulumi convert --from yaml --language typescript --out ../ts-organization-role-permission-sets
```

See the [`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for more options.
