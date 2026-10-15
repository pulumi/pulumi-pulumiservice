{{% examples %}}
## Example Usage

{{% example %}}
### Look up a role you defined

```typescript
import * as pulumiservice from "@pulumi/pulumiservice";

// A custom role managed in the Pulumi Cloud console.
// The name must match exactly (case-sensitive).
const ciDeployerRole = pulumiservice.getOrganizationRoleOutput({
    organizationName: "my-org",
    name: "CI Deployer",
});

// Assign it to a team without hardcoding the role ID.
const platformTeamRole = new pulumiservice.TeamRoleAssignment("platform-team-role", {
    organizationName: "my-org",
    teamName: "platform",
    roleId: ciDeployerRole.roleId,
});
```
{{% /example %}}

{{% example %}}
### Look up a built-in role

```typescript
import * as pulumiservice from "@pulumi/pulumiservice";

const adminRole = pulumiservice.getOrganizationRoleOutput({
    organizationName: "my-org",
    name: "Admin",
});

export const adminRoleId = adminRole.roleId;
```
{{% /example %}}
{{% /examples %}}