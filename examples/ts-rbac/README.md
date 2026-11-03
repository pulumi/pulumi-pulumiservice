# ts-rbac

TypeScript example that exercises the RBAC resources and data sources added in
the member-updates PR:

- `pulumiservice.OrganizationRole` — creates a custom `stack:read` role.
- `pulumiservice.Team` + `pulumiservice.TeamRoleAssignment` — creates a team
  and assigns the custom role to it. Requires the custom-roles feature to
  already be enabled on the organization.
- `pulumiservice.getOrganizationMembers` — lists org members.
- `pulumiservice.getOrganizationMember` — single-member lookup by username
  (demonstrates the #41668 follow-up).
- `pulumiservice.getOrganizationRoleScopes` — lists the scope catalogue you
  can reference in `OrganizationRole.permissions`.

## Prerequisites

- `PULUMI_ACCESS_TOKEN` set for an org admin on your target organization.
- The Custom Roles feature enabled on the organization.
- A local build of this provider branch. From the repo root:
  ```
  make build
  make install_nodejs_sdk   # yarn-links @pulumi/pulumiservice locally
  ```

## Run

```
cd examples/ts-rbac
yarn install
yarn link @pulumi/pulumiservice

# Point Pulumi at the locally-built provider binary. Without this, Pulumi
# resolves the published plugin from the registry, which does not know about
# the RBAC resources added in this PR.
export PATH=$(git rev-parse --show-toplevel)/bin:$PATH

pulumi stack init dev
pulumi config set organizationName <your-org>
pulumi config set targetUsername <existing-pulumi-cloud-username>
# Optional: override the suffix used in role/team names (default: "manual").
# pulumi config set nameSuffix $(date +%s)

pulumi up
```

After `pulumi up` you should see outputs including:

- `lookedUpByUsernameRole` with `targetUsername`'s org role.
- `firstMember` and `firstScope` from the two data sources.

## Teardown

```
pulumi destroy
pulumi stack rm dev
```

`destroy` revokes the custom-role assignment from the team and deletes the
role and the team.
