# py-rbac

Python example that exercises the RBAC resources and data sources:

- `pulumiservice.get_organization_members` — lists existing org members; the
  program filters this output to choose who to put on the new team.
- `pulumiservice.OrganizationRole` — creates a custom `stack:read` role.
- `pulumiservice.Team` + `pulumiservice.TeamRoleAssignment` — creates a team
  whose members are pulled from the data source, then attaches the custom
  role to it. Requires the custom-roles feature to already be enabled on
  the organization.

To convert this example to another language, use
[`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
