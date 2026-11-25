{{% examples %}}
## Example Usage

{{% example %}}
### A service principal with a role and entity access

The service principal holds the built-in Reader role across the organization, and can also open one environment. Its `login` output identifies it to other resources, such as `OrganizationMemberEntityAccess` and `DeploymentSettings`.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const readerRole = pulumiservice.getOrganizationRoleOutput({
    organizationName: "my-org",
    name: "Reader",
});
const deployer = new pulumiservice.ServicePrincipal("deployer", {
    organizationName: "my-org",
    name: "deployer",
    roleIds: [readerRole.roleId],
});
const deployerEnvironmentAccess = new pulumiservice.OrganizationMemberEntityAccess("deployerEnvironmentAccess", {
    organizationName: "my-org",
    username: deployer.login,
    entityType: "environment",
    entityName: "my-project/deploy-config",
    permission: "environment-open",
});
export const deployerLogin = deployer.login;
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

reader_role = pulumiservice.get_organization_role_output(organization_name="my-org",
    name="Reader")
deployer = pulumiservice.ServicePrincipal("deployer",
    organization_name="my-org",
    name="deployer",
    role_ids=[reader_role.role_id])
deployer_environment_access = pulumiservice.OrganizationMemberEntityAccess("deployerEnvironmentAccess",
    organization_name="my-org",
    username=deployer.login,
    entity_type="environment",
    entity_name="my-project/deploy-config",
    permission="environment-open")
pulumi.export("deployerLogin", deployer.login)
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		readerRole := pulumiservice.GetOrganizationRoleOutput(ctx, pulumiservice.GetOrganizationRoleOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Reader"),
		}, nil)
		deployer, err := pulumiservice.NewServicePrincipal(ctx, "deployer", &pulumiservice.ServicePrincipalArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("deployer"),
			RoleIds: pulumi.StringArray{
				readerRole.RoleId(),
			},
		})
		if err != nil {
			return err
		}
		_, err = pulumiservice.NewOrganizationMemberEntityAccess(ctx, "deployerEnvironmentAccess", &pulumiservice.OrganizationMemberEntityAccessArgs{
			OrganizationName: pulumi.String("my-org"),
			Username:         deployer.Login,
			EntityType:       pulumi.String("environment"),
			EntityName:       pulumi.String("my-project/deploy-config"),
			Permission:       pulumi.String("environment-open"),
		})
		if err != nil {
			return err
		}
		ctx.Export("deployerLogin", deployer.Login)
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
    var readerRole = PulumiService.GetOrganizationRole.Invoke(new()
    {
        OrganizationName = "my-org",
        Name = "Reader",
    });

    var deployer = new PulumiService.ServicePrincipal("deployer", new()
    {
        OrganizationName = "my-org",
        Name = "deployer",
        RoleIds = new[]
        {
            readerRole.Apply(getOrganizationRoleResult => getOrganizationRoleResult.RoleId),
        },
    });

    var deployerEnvironmentAccess = new PulumiService.OrganizationMemberEntityAccess("deployerEnvironmentAccess", new()
    {
        OrganizationName = "my-org",
        Username = deployer.Login,
        EntityType = "environment",
        EntityName = "my-project/deploy-config",
        Permission = "environment-open",
    });

    return new Dictionary<string, object?>
    {
        ["deployerLogin"] = deployer.Login,
    };
});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetOrganizationRoleArgs;
import com.pulumi.pulumiservice.ServicePrincipal;
import com.pulumi.pulumiservice.ServicePrincipalArgs;
import com.pulumi.pulumiservice.OrganizationMemberEntityAccess;
import com.pulumi.pulumiservice.OrganizationMemberEntityAccessArgs;
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
        final var readerRole = PulumiserviceFunctions.getOrganizationRole(GetOrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Reader")
            .build());

        var deployer = new ServicePrincipal("deployer", ServicePrincipalArgs.builder()
            .organizationName("my-org")
            .name("deployer")
            .roleIds(readerRole.applyValue(_readerRole -> _readerRole.roleId()))
            .build());

        var deployerEnvironmentAccess = new OrganizationMemberEntityAccess("deployerEnvironmentAccess", OrganizationMemberEntityAccessArgs.builder()
            .organizationName("my-org")
            .username(deployer.login())
            .entityType("environment")
            .entityName("my-project/deploy-config")
            .permission("environment-open")
            .build());

        ctx.export("deployerLogin", deployer.login());
    }
}
```

```yaml
variables:
  readerRole:
    fn::invoke:
      function: pulumiservice:getOrganizationRole
      arguments:
        organizationName: my-org
        name: Reader
resources:
  deployer:
    type: pulumiservice:ServicePrincipal
    properties:
      organizationName: my-org
      name: deployer
      roleIds:
        - ${readerRole.roleId}
  deployerEnvironmentAccess:
    type: pulumiservice:OrganizationMemberEntityAccess
    properties:
      organizationName: my-org
      username: ${deployer.login}
      entityType: environment
      entityName: my-project/deploy-config
      permission: environment-open
outputs:
  deployerLogin: ${deployer.login}
```
{{% /example %}}
{{% /examples %}}
