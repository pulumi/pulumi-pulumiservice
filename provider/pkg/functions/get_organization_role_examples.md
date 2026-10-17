{{% examples %}}
## Example Usage

{{% example %}}
### Look up a role you defined

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const ciDeployerRole = pulumiservice.getOrganizationRoleOutput({
    organizationName: "my-org",
    name: "CI Deployer",
});
const platformTeamRole = new pulumiservice.TeamRoleAssignment("platformTeamRole", {
    organizationName: "my-org",
    teamName: "platform",
    roleId: ciDeployerRole.roleId,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

ci_deployer_role = pulumiservice.get_organization_role_output(organization_name="my-org",
    name="CI Deployer")
platform_team_role = pulumiservice.TeamRoleAssignment("platformTeamRole",
    organization_name="my-org",
    team_name="platform",
    role_id=ci_deployer_role.role_id)
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		ciDeployerRole := pulumiservice.GetOrganizationRoleOutput(ctx, pulumiservice.GetOrganizationRoleOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("CI Deployer"),
		}, nil)
		_, err := pulumiservice.NewTeamRoleAssignment(ctx, "platformTeamRole", &pulumiservice.TeamRoleAssignmentArgs{
			OrganizationName: pulumi.String("my-org"),
			TeamName:         pulumi.String("platform"),
			RoleId:           ciDeployerRole.RoleId(),
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
    var ciDeployerRole = PulumiService.GetOrganizationRole.Invoke(new()
    {
        OrganizationName = "my-org",
        Name = "CI Deployer",
    });

    var platformTeamRole = new PulumiService.TeamRoleAssignment("platformTeamRole", new()
    {
        OrganizationName = "my-org",
        TeamName = "platform",
        RoleId = ciDeployerRole.Apply(getOrganizationRoleResult => getOrganizationRoleResult.RoleId),
    });

});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetOrganizationRoleArgs;
import com.pulumi.pulumiservice.TeamRoleAssignment;
import com.pulumi.pulumiservice.TeamRoleAssignmentArgs;
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
        final var ciDeployerRole = PulumiserviceFunctions.getOrganizationRole(GetOrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("CI Deployer")
            .build());

        var platformTeamRole = new TeamRoleAssignment("platformTeamRole", TeamRoleAssignmentArgs.builder()
            .organizationName("my-org")
            .teamName("platform")
            .roleId(ciDeployerRole.applyValue(_ciDeployerRole -> _ciDeployerRole.roleId()))
            .build());

    }
}
```

```yaml
variables:
  ciDeployerRole:
    fn::invoke:
      function: pulumiservice:getOrganizationRole
      arguments:
        organizationName: my-org
        name: CI Deployer
resources:
  platformTeamRole:
    type: pulumiservice:TeamRoleAssignment
    properties:
      organizationName: my-org
      teamName: platform
      roleId: ${ciDeployerRole.roleId}
```
{{% /example %}}

{{% example %}}
### Look up a built-in role

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const adminRole = pulumiservice.getOrganizationRoleOutput({
    organizationName: "my-org",
    name: "Admin",
});
export const adminRoleId = adminRole.roleId;
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

admin_role = pulumiservice.get_organization_role_output(organization_name="my-org",
    name="Admin")
pulumi.export("adminRoleId", admin_role.role_id)
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		adminRole := pulumiservice.GetOrganizationRoleOutput(ctx, pulumiservice.GetOrganizationRoleOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Admin"),
		}, nil)
		ctx.Export("adminRoleId", adminRole.RoleId())
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
    var adminRole = PulumiService.GetOrganizationRole.Invoke(new()
    {
        OrganizationName = "my-org",
        Name = "Admin",
    });

    return new Dictionary<string, object?>
    {
        ["adminRoleId"] = adminRole.Apply(getOrganizationRoleResult => getOrganizationRoleResult.RoleId),
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
        final var adminRole = PulumiserviceFunctions.getOrganizationRole(GetOrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Admin")
            .build());

        ctx.export("adminRoleId", adminRole.applyValue(_adminRole -> _adminRole.roleId()));
    }
}
```

```yaml
variables:
  adminRole:
    fn::invoke:
      function: pulumiservice:getOrganizationRole
      arguments:
        organizationName: my-org
        name: Admin
outputs:
  adminRoleId: ${adminRole.roleId}
```
{{% /example %}}
{{% /examples %}}
