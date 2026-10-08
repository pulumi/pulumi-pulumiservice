# yaml-rbac-roles

Builds a Pulumi Cloud custom role the way the console does
(**Settings > Access management**):

- **`pulumiservice:RbacPermissionSet`**: a custom permission set, a list of
  scopes for one kind of entity.
- **`pulumiservice:getRbacPermissionSet`**: looks up built-in permission sets
  such as "Stack Read" (`stack-read`) and "Read Only"
  (`org-settings-read-only`).
- **`pulumiservice:RbacRole`**: a role made of organization-level access plus
  entity rules. The rules here cover all stacks, one stack by ID, stacks
  matched by tags (including a `notEquals` condition), and one environment by
  ID.
- **`pulumiservice:TeamRoleAssignment`**: assigns the role to a team.

This program is also the Example Usage shown on the `RbacPermissionSet` and
`RbacRole` API docs.

## Prerequisites

The organization must have the Custom Roles feature enabled.

## Convert to another language

Use `pulumi convert` to translate this program to TypeScript, Python, Go,
C#, or Java:

```bash
pulumi convert --from yaml --language typescript --out ../ts-rbac-roles
```

See the [`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for more options.
