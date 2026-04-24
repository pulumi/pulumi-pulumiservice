# py-rbac

Python example that exercises the RBAC resources and data sources:

- `pulumiservice.get_organization_members` — lists existing org members; the
  program filters this output to choose who to put on the new team.
- `pulumiservice.OrganizationRole` — creates a custom `stack:read` role.
- `pulumiservice.Team` + `pulumiservice.TeamRoleAssignment` — creates a team
  whose members are pulled from the data source, then attaches the custom
  role to it (auto-enables the team custom-roles feature on first use).
- `pulumiservice.OrganizationMember` — adopts an existing member (409
  branch on create) and changes their built-in org role to `admin`. On
  destroy the adopted membership is left in place; only the role change
  needs to be reverted, which the integration test's cleanup hook does.

To convert this example to another language, use
[`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
