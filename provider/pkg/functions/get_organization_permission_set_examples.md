{{% examples %}}
## Example Usage

{{% example %}}
### Grant a built-in organization-level permission set

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const readOnly = pulumiservice.getOrganizationPermissionSetOutput({
    organizationName: "my-org",
    builtIn: pulumiservice.BuiltInOrganizationPermissionSet.ReadOnly,
});
const permissions = pulumiservice.buildRolePermissionsOutput({
    organizationPermissionSetScopes: readOnly.scopes,
});
const readOnlyRole = new pulumiservice.OrganizationRole("readOnlyRole", {
    organizationName: "my-org",
    name: "Org Readers",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

read_only = pulumiservice.get_organization_permission_set_output(organization_name="my-org",
    built_in=pulumiservice.BuiltInOrganizationPermissionSet.READ_ONLY)
permissions = pulumiservice.build_role_permissions_output(organization_permission_set_scopes=read_only.scopes)
read_only_role = pulumiservice.OrganizationRole("readOnlyRole",
    organization_name="my-org",
    name="Org Readers",
    permissions=permissions.permissions)
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		readOnly := pulumiservice.GetOrganizationPermissionSetOutput(ctx, pulumiservice.GetOrganizationPermissionSetOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			BuiltIn:          pulumiservice.BuiltInOrganizationPermissionSetReadOnly,
		}, nil)
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			OrganizationPermissionSetScopes: readOnly.Scopes(),
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "readOnlyRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Org Readers"),
			Permissions:      permissions.Permissions(),
		})
		if err != nil {
			return err
		}
		return nil
	})
}
```

```csharp
using System.Collections.Generic;
using System.Linq;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() => 
{
    var readOnly = PulumiService.GetOrganizationPermissionSet.Invoke(new()
    {
        OrganizationName = "my-org",
        BuiltIn = PulumiService.BuiltInOrganizationPermissionSet.ReadOnly,
    });

    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        OrganizationPermissionSetScopes = readOnly.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.Scopes),
    });

    var readOnlyRole = new PulumiService.OrganizationRole("readOnlyRole", new()
    {
        OrganizationName = "my-org",
        Name = "Org Readers",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.BuiltInOrganizationPermissionSet;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetOrganizationPermissionSetArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.OrganizationRole;
import com.pulumi.pulumiservice.OrganizationRoleArgs;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Map;
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    public static void stack(Context ctx) {
        final var readOnly = PulumiserviceFunctions.getOrganizationPermissionSet(GetOrganizationPermissionSetArgs.builder()
            .organizationName("my-org")
            .builtIn(BuiltInOrganizationPermissionSet.ReadOnly)
            .build());

        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .organizationPermissionSetScopes(readOnly.applyValue(_readOnly -> _readOnly.scopes()))
            .build());

        var readOnlyRole = new OrganizationRole("readOnlyRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Org Readers")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  readOnly:
    fn::invoke:
      function: pulumiservice:getOrganizationPermissionSet
      arguments:
        organizationName: my-org
        builtIn: org-settings-read-only
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        organizationPermissionSetScopes: ${readOnly.scopes}
resources:
  readOnlyRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Org Readers
      permissions: ${permissions.permissions}
```
{{% /example %}}
{{% /examples %}}
