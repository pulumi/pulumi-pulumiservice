{{% examples %}}
## Example Usage

{{% example %}}
### List all roles

```typescript
import * as pulumiservice from "@pulumi/pulumiservice";

const allRoles = pulumiservice.getOrganizationRolesOutput({ organizationName: "my-org" });

// Map each role name to its ID.
export const roleIdsByName = allRoles.roles.apply(roles =>
    Object.fromEntries(roles.map(r => [r.name, r.roleId])));
```
{{% /example %}}
{{% /examples %}}