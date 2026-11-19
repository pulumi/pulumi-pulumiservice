{{% examples %}}
## Example Usage

{{% example %}}
### Grant a built-in cloud account permission set

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const accountRead = pulumiservice.getAccountPermissionSetOutput({
    organizationName: "my-org",
    builtIn: pulumiservice.BuiltInAccountPermissionSet.Read,
});
const permissions = pulumiservice.buildRolePermissionsOutput({
    accountRules: [{
        all: true,
        permissionSetScopes: accountRead.scopes,
    }],
});
const accountReaderRole = new pulumiservice.OrganizationRole("accountReaderRole", {
    organizationName: "my-org",
    name: "Account Readers",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

account_read = pulumiservice.get_account_permission_set_output(organization_name="my-org",
    built_in=pulumiservice.BuiltInAccountPermissionSet.READ)
permissions = pulumiservice.build_role_permissions_output(account_rules=[{
    "all": True,
    "permission_set_scopes": account_read.scopes,
}])
account_reader_role = pulumiservice.OrganizationRole("accountReaderRole",
    organization_name="my-org",
    name="Account Readers",
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
		accountRead := pulumiservice.GetAccountPermissionSetOutput(ctx, pulumiservice.GetAccountPermissionSetOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			BuiltIn:          pulumiservice.BuiltInAccountPermissionSetRead,
		}, nil)
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			AccountRules: pulumiservice.RoleAccountRuleArray{
				&pulumiservice.RoleAccountRuleArgs{
					All:                 pulumi.Bool(true),
					PermissionSetScopes: accountRead.Scopes(),
				},
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "accountReaderRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Account Readers"),
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
    var accountRead = PulumiService.GetAccountPermissionSet.Invoke(new()
    {
        OrganizationName = "my-org",
        BuiltIn = PulumiService.BuiltInAccountPermissionSet.Read,
    });

    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        AccountRules = new[]
        {
            new PulumiService.Inputs.RoleAccountRuleInputArgs
            {
                All = true,
                PermissionSetScopes = accountRead.Apply(getAccountPermissionSetResult => getAccountPermissionSetResult.Scopes),
            },
        },
    });

    var accountReaderRole = new PulumiService.OrganizationRole("accountReaderRole", new()
    {
        OrganizationName = "my-org",
        Name = "Account Readers",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.BuiltInAccountPermissionSet;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetAccountPermissionSetArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleAccountRuleArgs;
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
        final var accountRead = PulumiserviceFunctions.getAccountPermissionSet(GetAccountPermissionSetArgs.builder()
            .organizationName("my-org")
            .builtIn(BuiltInAccountPermissionSet.Read)
            .build());

        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .accountRules(RoleAccountRuleArgs.builder()
                .all(true)
                .permissionSetScopes(accountRead.applyValue(_accountRead -> _accountRead.scopes()))
                .build())
            .build());

        var accountReaderRole = new OrganizationRole("accountReaderRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Account Readers")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  accountRead:
    fn::invoke:
      function: pulumiservice:getAccountPermissionSet
      arguments:
        organizationName: my-org
        builtIn: insights-account-read
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        accountRules:
          - all: true
            permissionSetScopes: ${accountRead.scopes}
resources:
  accountReaderRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Account Readers
      permissions: ${permissions.permissions}
```
{{% /example %}}
{{% /examples %}}
