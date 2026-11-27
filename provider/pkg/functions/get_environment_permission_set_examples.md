{{% examples %}}
## Example Usage

{{% example %}}
### Grant a built-in environment permission set

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const environmentOpen = pulumiservice.getEnvironmentPermissionSetOutput({
    organizationName: "my-org",
    builtIn: pulumiservice.BuiltInEnvironmentPermissionSet.Open,
});
const permissions = pulumiservice.buildRolePermissionsOutput({
    environmentRules: [{
        all: true,
        permissionSetScopes: environmentOpen.scopes,
    }],
});
const environmentOpenerRole = new pulumiservice.OrganizationRole("environmentOpenerRole", {
    organizationName: "my-org",
    name: "Environment Openers",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

environment_open = pulumiservice.get_environment_permission_set_output(organization_name="my-org",
    built_in=pulumiservice.BuiltInEnvironmentPermissionSet.OPEN)
permissions = pulumiservice.build_role_permissions_output(environment_rules=[{
    "all": True,
    "permission_set_scopes": environment_open.scopes,
}])
environment_opener_role = pulumiservice.OrganizationRole("environmentOpenerRole",
    organization_name="my-org",
    name="Environment Openers",
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
		environmentOpen := pulumiservice.GetEnvironmentPermissionSetOutput(ctx, pulumiservice.GetEnvironmentPermissionSetOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			BuiltIn:          pulumiservice.BuiltInEnvironmentPermissionSetOpen,
		}, nil)
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			EnvironmentRules: pulumiservice.RoleEnvironmentRuleArray{
				&pulumiservice.RoleEnvironmentRuleArgs{
					All:                 pulumi.Bool(true),
					PermissionSetScopes: environmentOpen.Scopes(),
				},
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "environmentOpenerRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Environment Openers"),
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
    var environmentOpen = PulumiService.GetEnvironmentPermissionSet.Invoke(new()
    {
        OrganizationName = "my-org",
        BuiltIn = PulumiService.BuiltInEnvironmentPermissionSet.Open,
    });

    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        EnvironmentRules = new[]
        {
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
            {
                All = true,
                PermissionSetScopes = environmentOpen.Apply(getEnvironmentPermissionSetResult => getEnvironmentPermissionSetResult.Scopes),
            },
        },
    });

    var environmentOpenerRole = new PulumiService.OrganizationRole("environmentOpenerRole", new()
    {
        OrganizationName = "my-org",
        Name = "Environment Openers",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.BuiltInEnvironmentPermissionSet;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetEnvironmentPermissionSetArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleEnvironmentRuleArgs;
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
        final var environmentOpen = PulumiserviceFunctions.getEnvironmentPermissionSet(GetEnvironmentPermissionSetArgs.builder()
            .organizationName("my-org")
            .builtIn(BuiltInEnvironmentPermissionSet.Open)
            .build());

        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .environmentRules(RoleEnvironmentRuleArgs.builder()
                .all(true)
                .permissionSetScopes(environmentOpen.applyValue(_environmentOpen -> _environmentOpen.scopes()))
                .build())
            .build());

        var environmentOpenerRole = new OrganizationRole("environmentOpenerRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Environment Openers")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  environmentOpen:
    fn::invoke:
      function: pulumiservice:getEnvironmentPermissionSet
      arguments:
        organizationName: my-org
        builtIn: environment-open
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        environmentRules:
          - all: true
            permissionSetScopes: ${environmentOpen.scopes}
resources:
  environmentOpenerRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Environment Openers
      permissions: ${permissions.permissions}
```
{{% /example %}}
{{% /examples %}}
