{{% examples %}}
## Example Usage

{{% example %}}
### Token with an expiry and a custom role

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const ciDeployerRole = pulumiservice.getOrganizationRoleOutput({
    organizationName: "my-org",
    name: "CI Deployer",
});
const ciToken = new pulumiservice.OrgAccessToken("ciToken", {
    organizationName: "my-org",
    name: "ci-token",
    description: "Token for CI pipelines",
    expires: "2027-06-30T17:00:00Z",
    roleId: ciDeployerRole.roleId,
});
export const ciTokenValue = ciToken.value;
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

ci_deployer_role = pulumiservice.get_organization_role_output(organization_name="my-org",
    name="CI Deployer")
ci_token = pulumiservice.OrgAccessToken("ciToken",
    organization_name="my-org",
    name="ci-token",
    description="Token for CI pipelines",
    expires="2027-06-30T17:00:00Z",
    role_id=ci_deployer_role.role_id)
pulumi.export("ciTokenValue", ci_token.value)
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
		ciToken, err := pulumiservice.NewOrgAccessToken(ctx, "ciToken", &pulumiservice.OrgAccessTokenArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("ci-token"),
			Description:      pulumi.String("Token for CI pipelines"),
			Expires:          pulumi.String("2027-06-30T17:00:00Z"),
			RoleId:           ciDeployerRole.RoleId(),
		})
		if err != nil {
			return err
		}
		ctx.Export("ciTokenValue", ciToken.Value)
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

    var ciToken = new PulumiService.OrgAccessToken("ciToken", new()
    {
        OrganizationName = "my-org",
        Name = "ci-token",
        Description = "Token for CI pipelines",
        Expires = "2027-06-30T17:00:00Z",
        RoleId = ciDeployerRole.Apply(getOrganizationRoleResult => getOrganizationRoleResult.RoleId),
    });

    return new Dictionary<string, object?>
    {
        ["ciTokenValue"] = ciToken.Value,
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
import com.pulumi.pulumiservice.OrgAccessToken;
import com.pulumi.pulumiservice.OrgAccessTokenArgs;
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

        var ciToken = new OrgAccessToken("ciToken", OrgAccessTokenArgs.builder()
            .organizationName("my-org")
            .name("ci-token")
            .description("Token for CI pipelines")
            .expires("2027-06-30T17:00:00Z")
            .roleId(ciDeployerRole.applyValue(_ciDeployerRole -> _ciDeployerRole.roleId()))
            .build());

        ctx.export("ciTokenValue", ciToken.value());
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
  ciToken:
    type: pulumiservice:OrgAccessToken
    properties:
      organizationName: my-org
      name: ci-token
      description: Token for CI pipelines
      expires: "2027-06-30T17:00:00Z"
      roleId: ${ciDeployerRole.roleId}
outputs:
  ciTokenValue: ${ciToken.value}
```
{{% /example %}}
{{% /examples %}}
