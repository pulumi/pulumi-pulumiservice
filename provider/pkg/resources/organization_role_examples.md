{{% examples %}}
## Example Usage

{{% example %}}
### Organization-level scopes

Organization scopes apply across the whole organization rather than to a single stack, environment, or cloud account.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const permissions = pulumiservice.buildRolePermissionsOutput({
    organizationScopes: [
        pulumiservice.RbacOrganizationScope.TeamRead,
        pulumiservice.RbacOrganizationScope.StackCreate,
    ],
});
const teamReaderRole = new pulumiservice.OrganizationRole("teamReaderRole", {
    organizationName: "my-org",
    name: "Team Readers",
    description: "Read teams and create stacks.",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

permissions = pulumiservice.build_role_permissions_output(organization_scopes=[
    pulumiservice.RbacOrganizationScope.TEAM_READ,
    pulumiservice.RbacOrganizationScope.STACK_CREATE,
])
team_reader_role = pulumiservice.OrganizationRole("teamReaderRole",
    organization_name="my-org",
    name="Team Readers",
    description="Read teams and create stacks.",
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
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			OrganizationScopes: pulumiservice.RbacOrganizationScopeArray{
				pulumiservice.RbacOrganizationScopeTeamRead,
				pulumiservice.RbacOrganizationScopeStackCreate,
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "teamReaderRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Team Readers"),
			Description:      pulumi.String("Read teams and create stacks."),
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
    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        OrganizationScopes = new[]
        {
            PulumiService.RbacOrganizationScope.TeamRead,
            PulumiService.RbacOrganizationScope.StackCreate,
        },
    });

    var teamReaderRole = new PulumiService.OrganizationRole("teamReaderRole", new()
    {
        OrganizationName = "my-org",
        Name = "Team Readers",
        Description = "Read teams and create stacks.",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.RbacOrganizationScope;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
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
        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .organizationScopes(            
                RbacOrganizationScope.TeamRead,
                RbacOrganizationScope.StackCreate)
            .build());

        var teamReaderRole = new OrganizationRole("teamReaderRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Team Readers")
            .description("Read teams and create stacks.")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        organizationScopes:
          - team:read
          - stack:create
resources:
  teamReaderRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Team Readers
      description: Read teams and create stacks.
      permissions: ${permissions.permissions}
```
{{% /example %}}

{{% example %}}
### Access to every stack, or to one stack

A stack rule with `all: true` applies to every stack in the organization. A rule with `id` applies to a single stack, identified by the `stackId` that `getStack` returns.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const prodStack = pulumiservice.getStackOutput({
    organizationName: "my-org",
    projectName: "networking",
    stackName: "prod",
});
const permissions = pulumiservice.buildRolePermissionsOutput({
    stackRules: [
        {
            all: true,
            scopes: [pulumiservice.RbacStackScope.StackRead],
        },
        {
            id: prodStack.stackId,
            scopes: [
                pulumiservice.RbacStackScope.StackWrite,
                pulumiservice.RbacStackScope.StackDeploymentCreate,
            ],
        },
    ],
});
const prodDeployerRole = new pulumiservice.OrganizationRole("prodDeployerRole", {
    organizationName: "my-org",
    name: "Prod Deployers",
    description: "Read every stack and deploy networking/prod.",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

prod_stack = pulumiservice.get_stack_output(organization_name="my-org",
    project_name="networking",
    stack_name="prod")
permissions = pulumiservice.build_role_permissions_output(stack_rules=[
    {
        "all": True,
        "scopes": [pulumiservice.RbacStackScope.STACK_READ],
    },
    {
        "id": prod_stack.stack_id,
        "scopes": [
            pulumiservice.RbacStackScope.STACK_WRITE,
            pulumiservice.RbacStackScope.STACK_DEPLOYMENT_CREATE,
        ],
    },
])
prod_deployer_role = pulumiservice.OrganizationRole("prodDeployerRole",
    organization_name="my-org",
    name="Prod Deployers",
    description="Read every stack and deploy networking/prod.",
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
		prodStack := pulumiservice.GetStackOutput(ctx, pulumiservice.GetStackOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			ProjectName:      pulumi.String("networking"),
			StackName:        pulumi.String("prod"),
		}, nil)
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			StackRules: pulumiservice.RoleStackRuleArray{
				&pulumiservice.RoleStackRuleArgs{
					All: pulumi.Bool(true),
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackRead,
					},
				},
				&pulumiservice.RoleStackRuleArgs{
					Id: prodStack.StackId(),
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackWrite,
						pulumiservice.RbacStackScopeStackDeploymentCreate,
					},
				},
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "prodDeployerRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Prod Deployers"),
			Description:      pulumi.String("Read every stack and deploy networking/prod."),
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
    var prodStack = PulumiService.GetStack.Invoke(new()
    {
        OrganizationName = "my-org",
        ProjectName = "networking",
        StackName = "prod",
    });

    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        StackRules = new[]
        {
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                All = true,
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackRead,
                },
            },
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Id = prodStack.Apply(getStackResult => getStackResult.StackId),
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackWrite,
                    PulumiService.RbacStackScope.StackDeploymentCreate,
                },
            },
        },
    });

    var prodDeployerRole = new PulumiService.OrganizationRole("prodDeployerRole", new()
    {
        OrganizationName = "my-org",
        Name = "Prod Deployers",
        Description = "Read every stack and deploy networking/prod.",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.RbacStackScope;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetStackArgs;
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
        final var prodStack = PulumiserviceFunctions.getStack(GetStackArgs.builder()
            .organizationName("my-org")
            .projectName("networking")
            .stackName("prod")
            .build());

        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .stackRules(            
                RoleStackRuleArgs.builder()
                    .all(true)
                    .scopes(RbacStackScope.StackRead)
                    .build(),
                RoleStackRuleArgs.builder()
                    .id(prodStack.applyValue(_prodStack -> _prodStack.stackId()))
                    .scopes(                    
                        RbacStackScope.StackWrite,
                        RbacStackScope.StackDeploymentCreate)
                    .build())
            .build());

        var prodDeployerRole = new OrganizationRole("prodDeployerRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Prod Deployers")
            .description("Read every stack and deploy networking/prod.")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  prodStack:
    fn::invoke:
      function: pulumiservice:getStack
      arguments:
        organizationName: my-org
        projectName: networking
        stackName: prod
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        stackRules:
          - all: true
            scopes:
              - stack:read
          - id: ${prodStack.stackId}
            scopes:
              - stack:write
              - stack_deployment:create
resources:
  prodDeployerRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Prod Deployers
      description: Read every stack and deploy networking/prod.
      permissions: ${permissions.permissions}
```
{{% /example %}}

{{% example %}}
### Access to stacks by tag

A rule with `tags` applies to stacks whose tags match every condition in the list. A condition without an `operator` matches stacks where the tag equals `value`. `notEquals` also matches stacks that don't have the tag. `exists` and `notExists` check only whether the tag is set.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const permissions = pulumiservice.buildRolePermissionsOutput({
    stackRules: [
        {
            tags: [{
                key: "team",
                value: "platform",
            }],
            scopes: [
                pulumiservice.RbacStackScope.StackWrite,
                pulumiservice.RbacStackScope.StackDeploymentCreate,
            ],
        },
        {
            tags: [{
                key: "env",
                value: "prod",
                operator: pulumiservice.TagOperator.NotEquals,
            }],
            scopes: [pulumiservice.RbacStackScope.StackDeploymentCreate],
        },
        {
            tags: [{
                key: "cost-center",
                operator: pulumiservice.TagOperator.Exists,
            }],
            scopes: [pulumiservice.RbacStackScope.StackDeploymentSettingsRead],
        },
        {
            tags: [{
                key: "owner",
                operator: pulumiservice.TagOperator.NotExists,
            }],
            scopes: [pulumiservice.RbacStackScope.StackTagsUpdate],
        },
        {
            tags: [
                {
                    key: "team",
                    value: "platform",
                },
                {
                    key: "lifecycle",
                    value: "ephemeral",
                },
            ],
            scopes: [pulumiservice.RbacStackScope.StackDelete],
        },
    ],
});
const platformRole = new pulumiservice.OrganizationRole("platformRole", {
    organizationName: "my-org",
    name: "Platform Engineers",
    description: "Deploy platform stacks and clean up ephemeral ones.",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

permissions = pulumiservice.build_role_permissions_output(stack_rules=[
    {
        "tags": [{
            "key": "team",
            "value": "platform",
        }],
        "scopes": [
            pulumiservice.RbacStackScope.STACK_WRITE,
            pulumiservice.RbacStackScope.STACK_DEPLOYMENT_CREATE,
        ],
    },
    {
        "tags": [{
            "key": "env",
            "value": "prod",
            "operator": pulumiservice.TagOperator.NOT_EQUALS,
        }],
        "scopes": [pulumiservice.RbacStackScope.STACK_DEPLOYMENT_CREATE],
    },
    {
        "tags": [{
            "key": "cost-center",
            "operator": pulumiservice.TagOperator.EXISTS,
        }],
        "scopes": [pulumiservice.RbacStackScope.STACK_DEPLOYMENT_SETTINGS_READ],
    },
    {
        "tags": [{
            "key": "owner",
            "operator": pulumiservice.TagOperator.NOT_EXISTS,
        }],
        "scopes": [pulumiservice.RbacStackScope.STACK_TAGS_UPDATE],
    },
    {
        "tags": [
            {
                "key": "team",
                "value": "platform",
            },
            {
                "key": "lifecycle",
                "value": "ephemeral",
            },
        ],
        "scopes": [pulumiservice.RbacStackScope.STACK_DELETE],
    },
])
platform_role = pulumiservice.OrganizationRole("platformRole",
    organization_name="my-org",
    name="Platform Engineers",
    description="Deploy platform stacks and clean up ephemeral ones.",
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
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			StackRules: pulumiservice.RoleStackRuleArray{
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackWrite,
						pulumiservice.RbacStackScopeStackDeploymentCreate,
					},
				},
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:      pulumi.String("env"),
							Value:    pulumi.String("prod"),
							Operator: pulumiservice.TagOperatorNotEquals,
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackDeploymentCreate,
					},
				},
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:      pulumi.String("cost-center"),
							Operator: pulumiservice.TagOperatorExists,
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackDeploymentSettingsRead,
					},
				},
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:      pulumi.String("owner"),
							Operator: pulumiservice.TagOperatorNotExists,
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackTagsUpdate,
					},
				},
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
						&pulumiservice.TagConditionArgs{
							Key:   pulumi.String("lifecycle"),
							Value: pulumi.String("ephemeral"),
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackDelete,
					},
				},
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "platformRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Platform Engineers"),
			Description:      pulumi.String("Deploy platform stacks and clean up ephemeral ones."),
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
    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        StackRules = new[]
        {
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.TagConditionInputArgs
                    {
                        Key = "team",
                        Value = "platform",
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackWrite,
                    PulumiService.RbacStackScope.StackDeploymentCreate,
                },
            },
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.TagConditionInputArgs
                    {
                        Key = "env",
                        Value = "prod",
                        Operator = PulumiService.TagOperator.NotEquals,
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackDeploymentCreate,
                },
            },
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.TagConditionInputArgs
                    {
                        Key = "cost-center",
                        Operator = PulumiService.TagOperator.Exists,
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackDeploymentSettingsRead,
                },
            },
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.TagConditionInputArgs
                    {
                        Key = "owner",
                        Operator = PulumiService.TagOperator.NotExists,
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackTagsUpdate,
                },
            },
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.TagConditionInputArgs
                    {
                        Key = "team",
                        Value = "platform",
                    },
                    new PulumiService.Inputs.TagConditionInputArgs
                    {
                        Key = "lifecycle",
                        Value = "ephemeral",
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackDelete,
                },
            },
        },
    });

    var platformRole = new PulumiService.OrganizationRole("platformRole", new()
    {
        OrganizationName = "my-org",
        Name = "Platform Engineers",
        Description = "Deploy platform stacks and clean up ephemeral ones.",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.RbacStackScope;
import com.pulumi.pulumiservice.enums.TagOperator;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleStackRuleArgs;
import com.pulumi.pulumiservice.inputs.TagConditionArgs;
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
        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .stackRules(            
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(                    
                        RbacStackScope.StackWrite,
                        RbacStackScope.StackDeploymentCreate)
                    .build(),
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("env")
                        .value("prod")
                        .operator(TagOperator.NotEquals)
                        .build())
                    .scopes(RbacStackScope.StackDeploymentCreate)
                    .build(),
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("cost-center")
                        .operator(TagOperator.Exists)
                        .build())
                    .scopes(RbacStackScope.StackDeploymentSettingsRead)
                    .build(),
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("owner")
                        .operator(TagOperator.NotExists)
                        .build())
                    .scopes(RbacStackScope.StackTagsUpdate)
                    .build(),
                RoleStackRuleArgs.builder()
                    .tags(                    
                        TagConditionArgs.builder()
                            .key("team")
                            .value("platform")
                            .build(),
                        TagConditionArgs.builder()
                            .key("lifecycle")
                            .value("ephemeral")
                            .build())
                    .scopes(RbacStackScope.StackDelete)
                    .build())
            .build());

        var platformRole = new OrganizationRole("platformRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Platform Engineers")
            .description("Deploy platform stacks and clean up ephemeral ones.")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        stackRules:
          - tags:
              - key: team
                value: platform
            scopes:
              - stack:write
              - stack_deployment:create
          - tags:
              - key: env
                value: prod
                operator: notEquals
            scopes:
              - stack_deployment:create
          - tags:
              - key: cost-center
                operator: exists
            scopes:
              - stack_deployment_settings:read
          - tags:
              - key: owner
                operator: notExists
            scopes:
              - stack_tags:update
          - tags:
              - key: team
                value: platform
              - key: lifecycle
                value: ephemeral
            scopes:
              - stack:delete
resources:
  platformRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Platform Engineers
      description: Deploy platform stacks and clean up ephemeral ones.
      permissions: ${permissions.permissions}
```
{{% /example %}}

{{% example %}}
### Access to environments and cloud accounts

Environment and cloud account rules take the same `all`, `id`, and `tags` selectors as stack rules. Each rule list accepts only the scopes for its own entity type.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const sharedEnvironment = pulumiservice.getEnvironmentOutput({
    organizationName: "my-org",
    projectName: "platform",
    name: "shared",
});
const awsAccount = pulumiservice.getInsightsAccountOutput({
    organizationName: "my-org",
    accountName: "aws-production",
});
const permissions = pulumiservice.buildRolePermissionsOutput({
    environmentRules: [
        {
            all: true,
            scopes: [pulumiservice.RbacEnvironmentScope.EnvironmentRead],
        },
        {
            id: sharedEnvironment.environmentId,
            scopes: [
                pulumiservice.RbacEnvironmentScope.EnvironmentOpen,
                pulumiservice.RbacEnvironmentScope.EnvironmentWrite,
                pulumiservice.RbacEnvironmentScope.EnvironmentRotate,
            ],
        },
    ],
    accountRules: [
        {
            all: true,
            scopes: [pulumiservice.RbacAccountScope.InsightsAccountRead],
        },
        {
            id: awsAccount.insightsAccountId,
            scopes: [pulumiservice.RbacAccountScope.InsightsAccountScan],
        },
        {
            tags: [{
                key: "team",
                value: "platform",
            }],
            scopes: [pulumiservice.RbacAccountScope.InsightsAccountUpdate],
        },
    ],
});
const infraRole = new pulumiservice.OrganizationRole("infraRole", {
    organizationName: "my-org",
    name: "Infrastructure Operators",
    description: "Read environments and cloud accounts, and operate the shared ones.",
    permissions: permissions.permissions,
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

shared_environment = pulumiservice.get_environment_output(organization_name="my-org",
    project_name="platform",
    name="shared")
aws_account = pulumiservice.get_insights_account_output(organization_name="my-org",
    account_name="aws-production")
permissions = pulumiservice.build_role_permissions_output(environment_rules=[
        {
            "all": True,
            "scopes": [pulumiservice.RbacEnvironmentScope.ENVIRONMENT_READ],
        },
        {
            "id": shared_environment.environment_id,
            "scopes": [
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_OPEN,
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_WRITE,
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_ROTATE,
            ],
        },
    ],
    account_rules=[
        {
            "all": True,
            "scopes": [pulumiservice.RbacAccountScope.INSIGHTS_ACCOUNT_READ],
        },
        {
            "id": aws_account.insights_account_id,
            "scopes": [pulumiservice.RbacAccountScope.INSIGHTS_ACCOUNT_SCAN],
        },
        {
            "tags": [{
                "key": "team",
                "value": "platform",
            }],
            "scopes": [pulumiservice.RbacAccountScope.INSIGHTS_ACCOUNT_UPDATE],
        },
    ])
infra_role = pulumiservice.OrganizationRole("infraRole",
    organization_name="my-org",
    name="Infrastructure Operators",
    description="Read environments and cloud accounts, and operate the shared ones.",
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
		sharedEnvironment := pulumiservice.GetEnvironmentOutput(ctx, pulumiservice.GetEnvironmentOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			ProjectName:      pulumi.String("platform"),
			Name:             pulumi.String("shared"),
		}, nil)
		awsAccount := pulumiservice.GetInsightsAccountOutput(ctx, pulumiservice.GetInsightsAccountOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			AccountName:      pulumi.String("aws-production"),
		}, nil)
		permissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			EnvironmentRules: pulumiservice.RoleEnvironmentRuleArray{
				&pulumiservice.RoleEnvironmentRuleArgs{
					All: pulumi.Bool(true),
					Scopes: pulumiservice.RbacEnvironmentScopeArray{
						pulumiservice.RbacEnvironmentScopeEnvironmentRead,
					},
				},
				&pulumiservice.RoleEnvironmentRuleArgs{
					Id: sharedEnvironment.EnvironmentId(),
					Scopes: pulumiservice.RbacEnvironmentScopeArray{
						pulumiservice.RbacEnvironmentScopeEnvironmentOpen,
						pulumiservice.RbacEnvironmentScopeEnvironmentWrite,
						pulumiservice.RbacEnvironmentScopeEnvironmentRotate,
					},
				},
			},
			AccountRules: pulumiservice.RoleAccountRuleArray{
				&pulumiservice.RoleAccountRuleArgs{
					All: pulumi.Bool(true),
					Scopes: pulumiservice.RbacAccountScopeArray{
						pulumiservice.RbacAccountScopeInsightsAccountRead,
					},
				},
				&pulumiservice.RoleAccountRuleArgs{
					Id: awsAccount.InsightsAccountId(),
					Scopes: pulumiservice.RbacAccountScopeArray{
						pulumiservice.RbacAccountScopeInsightsAccountScan,
					},
				},
				&pulumiservice.RoleAccountRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
					},
					Scopes: pulumiservice.RbacAccountScopeArray{
						pulumiservice.RbacAccountScopeInsightsAccountUpdate,
					},
				},
			},
		}, nil)
		_, err := pulumiservice.NewOrganizationRole(ctx, "infraRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("Infrastructure Operators"),
			Description:      pulumi.String("Read environments and cloud accounts, and operate the shared ones."),
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
    var sharedEnvironment = PulumiService.GetEnvironment.Invoke(new()
    {
        OrganizationName = "my-org",
        ProjectName = "platform",
        Name = "shared",
    });

    var awsAccount = PulumiService.GetInsightsAccount.Invoke(new()
    {
        OrganizationName = "my-org",
        AccountName = "aws-production",
    });

    var permissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        EnvironmentRules = new[]
        {
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
            {
                All = true,
                Scopes = new[]
                {
                    PulumiService.RbacEnvironmentScope.EnvironmentRead,
                },
            },
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
            {
                Id = sharedEnvironment.Apply(getEnvironmentResult => getEnvironmentResult.EnvironmentId),
                Scopes = new[]
                {
                    PulumiService.RbacEnvironmentScope.EnvironmentOpen,
                    PulumiService.RbacEnvironmentScope.EnvironmentWrite,
                    PulumiService.RbacEnvironmentScope.EnvironmentRotate,
                },
            },
        },
        AccountRules = new[]
        {
            new PulumiService.Inputs.RoleAccountRuleInputArgs
            {
                All = true,
                Scopes = new[]
                {
                    PulumiService.RbacAccountScope.InsightsAccountRead,
                },
            },
            new PulumiService.Inputs.RoleAccountRuleInputArgs
            {
                Id = awsAccount.Apply(getInsightsAccountResult => getInsightsAccountResult.InsightsAccountId),
                Scopes = new[]
                {
                    PulumiService.RbacAccountScope.InsightsAccountScan,
                },
            },
            new PulumiService.Inputs.RoleAccountRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.TagConditionInputArgs
                    {
                        Key = "team",
                        Value = "platform",
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacAccountScope.InsightsAccountUpdate,
                },
            },
        },
    });

    var infraRole = new PulumiService.OrganizationRole("infraRole", new()
    {
        OrganizationName = "my-org",
        Name = "Infrastructure Operators",
        Description = "Read environments and cloud accounts, and operate the shared ones.",
        Permissions = permissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});
```

```java
package generated_program;

import com.pulumi.pulumiservice.enums.RbacAccountScope;
import com.pulumi.pulumiservice.enums.RbacEnvironmentScope;
import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetEnvironmentArgs;
import com.pulumi.pulumiservice.inputs.GetInsightsAccountArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleEnvironmentRuleArgs;
import com.pulumi.pulumiservice.inputs.RoleAccountRuleArgs;
import com.pulumi.pulumiservice.inputs.TagConditionArgs;
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
        final var sharedEnvironment = PulumiserviceFunctions.getEnvironment(GetEnvironmentArgs.builder()
            .organizationName("my-org")
            .projectName("platform")
            .name("shared")
            .build());

        final var awsAccount = PulumiserviceFunctions.getInsightsAccount(GetInsightsAccountArgs.builder()
            .organizationName("my-org")
            .accountName("aws-production")
            .build());

        final var permissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .environmentRules(            
                RoleEnvironmentRuleArgs.builder()
                    .all(true)
                    .scopes(RbacEnvironmentScope.EnvironmentRead)
                    .build(),
                RoleEnvironmentRuleArgs.builder()
                    .id(sharedEnvironment.applyValue(_sharedEnvironment -> _sharedEnvironment.environmentId()))
                    .scopes(                    
                        RbacEnvironmentScope.EnvironmentOpen,
                        RbacEnvironmentScope.EnvironmentWrite,
                        RbacEnvironmentScope.EnvironmentRotate)
                    .build())
            .accountRules(            
                RoleAccountRuleArgs.builder()
                    .all(true)
                    .scopes(RbacAccountScope.InsightsAccountRead)
                    .build(),
                RoleAccountRuleArgs.builder()
                    .id(awsAccount.applyValue(_awsAccount -> _awsAccount.insightsAccountId()))
                    .scopes(RbacAccountScope.InsightsAccountScan)
                    .build(),
                RoleAccountRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(RbacAccountScope.InsightsAccountUpdate)
                    .build())
            .build());

        var infraRole = new OrganizationRole("infraRole", OrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("Infrastructure Operators")
            .description("Read environments and cloud accounts, and operate the shared ones.")
            .permissions(permissions.applyValue(_permissions -> _permissions.permissions()))
            .build());

    }
}
```

```yaml
variables:
  sharedEnvironment:
    fn::invoke:
      function: pulumiservice:getEnvironment
      arguments:
        organizationName: my-org
        projectName: platform
        name: shared
  awsAccount:
    fn::invoke:
      function: pulumiservice:getInsightsAccount
      arguments:
        organizationName: my-org
        accountName: aws-production
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        environmentRules:
          - all: true
            scopes:
              - environment:read
          - id: ${sharedEnvironment.environmentId}
            scopes:
              - environment:open
              - environment:write
              - environment:rotate
        accountRules:
          - all: true
            scopes:
              - insights_account:read
          - id: ${awsAccount.insightsAccountId}
            scopes:
              - insights_account:scan
          - tags:
              - key: team
                value: platform
            scopes:
              - insights_account:update
resources:
  infraRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: my-org
      name: Infrastructure Operators
      description: Read environments and cloud accounts, and operate the shared ones.
      permissions: ${permissions.permissions}
```
{{% /example %}}
{{% /examples %}}
