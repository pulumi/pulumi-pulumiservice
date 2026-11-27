{{% examples %}}
## Example Usage

{{% example %}}
### Grant a built-in stack permission set

Built-in permission sets, such as Stack Write, are the access levels every organization has. Pass the set's `scopes` to a stack rule's `permissionSetScopes`.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const stackWrite = pulumiservice.getStackPermissionSetOutput({
    organizationName: "my-org",
    builtIn: pulumiservice.BuiltInStackPermissionSet.Write,
});
const permissions = pulumiservice.buildRolePermissionsOutput({
    stackRules: [{
        all: true,
        permissionSetScopes: stackWrite.scopes,
    }],
});
const stackWriterRole = new pulumiservice.OrganizationRole("stackWriterRole", {
    organizationName: "my-org",
    name: "Stack Writers",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

stack_write = pulumiservice.get_stack_permission_set_output(organization_name="my-org",
    built_in=pulumiservice.BuiltInStackPermissionSet.WRITE)
permissions = pulumiservice.build_role_permissions_output(stack_rules=[{
    "all": True,
    "permission_set_scopes": stack_write.scopes,
}])
stack_writer_role = pulumiservice.OrganizationRole("stackWriterRole",
    organization_name="my-org",
    name="Stack Writers",
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
		stackWrite := pulumiservice.GetStackPermissionSetOutput(ctx, pulumiservice.GetStackPermissionSetOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			BuiltIn:          pulumiservice.BuiltInStackPermissionSetWrite,
		}, nil)
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			StackRules: pulumiservice.RoleStackRuleArray{
				&pulumiservice.RoleStackRuleArgs{
					All:                 pulumi.Bool(true),
					PermissionSetScopes: stackWrite.Scopes(),
				},
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "stackWriterRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Stack Writers"),
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
    var stackWrite = PulumiService.GetStackPermissionSet.Invoke(new()
    {
        OrganizationName = "my-org",
        BuiltIn = PulumiService.BuiltInStackPermissionSet.Write,
    });

    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        StackRules = new[]
        {
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                All = true,
                PermissionSetScopes = stackWrite.Apply(getStackPermissionSetResult => getStackPermissionSetResult.Scopes),
            },
        },
    });

    var stackWriterRole = new PulumiService.OrganizationRole("stackWriterRole", new()
    {
        OrganizationName = "my-org",
        Name = "Stack Writers",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.BuiltInStackPermissionSet;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetStackPermissionSetArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleStackRuleArgs;
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
        final var stackWrite = PulumiserviceFunctions.getStackPermissionSet(GetStackPermissionSetArgs.builder()
            .organizationName("my-org")
            .builtIn(BuiltInStackPermissionSet.Write)
            .build());

        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .stackRules(RoleStackRuleArgs.builder()
                .all(true)
                .permissionSetScopes(stackWrite.applyValue(_stackWrite -> _stackWrite.scopes()))
                .build())
            .build());

        var stackWriterRole = new OrganizationRole("stackWriterRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Stack Writers")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  stackWrite:
    fn::invoke:
      function: pulumiservice:getStackPermissionSet
      arguments:
        organizationName: my-org
        builtIn: stack-write
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        stackRules:
          - all: true
            permissionSetScopes: ${stackWrite.scopes}
resources:
  stackWriterRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Stack Writers
      permissions: ${permissions.permissions}
```
{{% /example %}}

{{% example %}}
### Grant a custom stack permission set

A custom permission set is looked up by the name shown in the Pulumi Cloud console.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const deployers = pulumiservice.getStackPermissionSetOutput({
    organizationName: "my-org",
    name: "Deployers",
});
const permissions = pulumiservice.buildRolePermissionsOutput({
    stackRules: [{
        all: true,
        permissionSetScopes: deployers.scopes,
    }],
});
const deployerRole = new pulumiservice.OrganizationRole("deployerRole", {
    organizationName: "my-org",
    name: "Deployers",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

deployers = pulumiservice.get_stack_permission_set_output(organization_name="my-org",
    name="Deployers")
permissions = pulumiservice.build_role_permissions_output(stack_rules=[{
    "all": True,
    "permission_set_scopes": deployers.scopes,
}])
deployer_role = pulumiservice.OrganizationRole("deployerRole",
    organization_name="my-org",
    name="Deployers",
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
		deployers := pulumiservice.GetStackPermissionSetOutput(ctx, pulumiservice.GetStackPermissionSetOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Deployers"),
		}, nil)
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			StackRules: pulumiservice.RoleStackRuleArray{
				&pulumiservice.RoleStackRuleArgs{
					All:                 pulumi.Bool(true),
					PermissionSetScopes: deployers.Scopes(),
				},
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "deployerRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Deployers"),
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
    var deployers = PulumiService.GetStackPermissionSet.Invoke(new()
    {
        OrganizationName = "my-org",
        Name = "Deployers",
    });

    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        StackRules = new[]
        {
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                All = true,
                PermissionSetScopes = deployers.Apply(getStackPermissionSetResult => getStackPermissionSetResult.Scopes),
            },
        },
    });

    var deployerRole = new PulumiService.OrganizationRole("deployerRole", new()
    {
        OrganizationName = "my-org",
        Name = "Deployers",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetStackPermissionSetArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleStackRuleArgs;
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
        final var deployers = PulumiserviceFunctions.getStackPermissionSet(GetStackPermissionSetArgs.builder()
            .organizationName("my-org")
            .name("Deployers")
            .build());

        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .stackRules(RoleStackRuleArgs.builder()
                .all(true)
                .permissionSetScopes(deployers.applyValue(_deployers -> _deployers.scopes()))
                .build())
            .build());

        var deployerRole = new OrganizationRole("deployerRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Deployers")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  deployers:
    fn::invoke:
      function: pulumiservice:getStackPermissionSet
      arguments:
        organizationName: my-org
        name: Deployers
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        stackRules:
          - all: true
            permissionSetScopes: ${deployers.scopes}
resources:
  deployerRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Deployers
      permissions: ${permissions.permissions}
```
{{% /example %}}
{{% /examples %}}
