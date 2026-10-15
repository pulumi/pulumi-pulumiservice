{{% examples %}}
## Example Usage

{{% example %}}
### List all roles

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const allRoles = pulumiservice.getOrganizationRolesOutput({
    organizationName: "my-org",
});
export const roles = allRoles.apply(allRoles => allRoles.roles);
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

all_roles = pulumiservice.get_organization_roles_output(organization_name="my-org")
pulumi.export("roles", all_roles.roles)
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		allRoles := pulumiservice.GetOrganizationRolesOutput(ctx, pulumiservice.GetOrganizationRolesOutputArgs{
			OrganizationName: pulumi.String("my-org"),
		}, nil)
		ctx.Export("roles", allRoles.ApplyT(func(allRoles pulumiservice.GetOrganizationRolesResult) ([]pulumiservice.OrganizationRoleInfo, error) {
			return allRoles.Roles.([]pulumiservice.OrganizationRoleInfo), nil
		}).(pulumi.ArrayOutput))
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
    var allRoles = PulumiService.GetOrganizationRoles.Invoke(new()
    {
        OrganizationName = "my-org",
    });

    return new Dictionary<string, object?>
    {
        ["roles"] = allRoles.Apply(getOrganizationRolesResult => getOrganizationRolesResult.Roles),
    };
});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetOrganizationRolesArgs;
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
        final var allRoles = PulumiserviceFunctions.getOrganizationRoles(GetOrganizationRolesArgs.builder()
            .organizationName("my-org")
            .build());

        ctx.export("roles", allRoles.applyValue(_allRoles -> _allRoles.roles()));
    }
}
```

```yaml
variables:
  allRoles:
    fn::invoke:
      function: pulumiservice:getOrganizationRoles
      arguments:
        organizationName: my-org
outputs:
  roles: ${allRoles.roles}
```
{{% /example %}}
{{% /examples %}}
