{{% examples %}}
## Example Usage

{{% example %}}
### A role with organization-level access and entity rules

The program looks up an existing stack, environment, and cloud account, then defines a role that grants scopes at the organization level and on each kind of entity.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") || "my-org";
const prodStack = pulumiservice.getStackOutput({
    organizationName: organizationName,
    projectName: "networking",
    stackName: "prod",
});
const sharedEnvironment = pulumiservice.getEnvironmentOutput({
    organizationName: organizationName,
    projectName: "platform",
    name: "shared",
});
const awsAccount = pulumiservice.getInsightsAccountOutput({
    organizationName: organizationName,
    accountName: "aws-production",
});
const platformPermissions = pulumiservice.buildRolePermissionsOutput({
    organizationScopes: [
        pulumiservice.RbacOrganizationScope.TeamRead,
        pulumiservice.RbacOrganizationScope.StackCreate,
    ],
    stackRules: [
        // Read every stack in the organization
        {
            all: true,
            scopes: [pulumiservice.RbacStackScope.StackRead],
        },
        // Deploy one specific stack
        {
            id: prodStack.stackId,
            scopes: [
                pulumiservice.RbacStackScope.StackWrite,
                pulumiservice.RbacStackScope.StackDeploymentCreate,
            ],
        },
        // Deploy stacks tagged team=platform (the operator defaults to equals)
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
        // Deploy stacks whose env tag is anything but "prod"
        {
            tags: [{
                key: "env",
                value: "prod",
                operator: pulumiservice.TagOperator.NotEquals,
            }],
            scopes: [
                pulumiservice.RbacStackScope.StackWrite,
                pulumiservice.RbacStackScope.StackDeploymentCreate,
            ],
        },
        // Read deployment settings on stacks that have a cost-center tag, whatever its value
        {
            tags: [{
                key: "cost-center",
                operator: pulumiservice.TagOperator.Exists,
            }],
            scopes: [pulumiservice.RbacStackScope.StackDeploymentSettingsRead],
        },
        // Tag stacks that don't have an owner tag yet
        {
            tags: [{
                key: "owner",
                operator: pulumiservice.TagOperator.NotExists,
            }],
            scopes: [pulumiservice.RbacStackScope.StackTagsUpdate],
        },
        // Delete stacks that match every tag condition
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
    environmentRules: [
        // Read every environment
        {
            all: true,
            scopes: [pulumiservice.RbacEnvironmentScope.EnvironmentRead],
        },
        // Open, update, and rotate one specific environment
        {
            id: sharedEnvironment.environmentId,
            scopes: [
                pulumiservice.RbacEnvironmentScope.EnvironmentOpen,
                pulumiservice.RbacEnvironmentScope.EnvironmentWrite,
                pulumiservice.RbacEnvironmentScope.EnvironmentRotate,
            ],
        },
        // Open environments tagged team=platform
        {
            tags: [{
                key: "team",
                value: "platform",
            }],
            scopes: [pulumiservice.RbacEnvironmentScope.EnvironmentOpen],
        },
    ],
    accountRules: [
        // Read every cloud account
        {
            all: true,
            scopes: [pulumiservice.RbacAccountScope.InsightsAccountRead],
        },
        // Scan one specific account
        {
            id: awsAccount.insightsAccountId,
            scopes: [pulumiservice.RbacAccountScope.InsightsAccountScan],
        },
        // Update accounts tagged team=platform
        {
            tags: [{
                key: "team",
                value: "platform",
            }],
            scopes: [pulumiservice.RbacAccountScope.InsightsAccountUpdate],
        },
    ],
});
const platformRole = new pulumiservice.OrganizationRole("platformRole", {
    organizationName: organizationName,
    name: "Platform Engineers",
    description: "Read access across the organization, plus deploy access to platform stacks.",
    permissions: platformPermissions.permissions,
});

```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

config = pulumi.Config()
organization_name = config.get("organizationName")
if organization_name is None:
    organization_name = "my-org"
prod_stack = pulumiservice.get_stack_output(organization_name=organization_name,
    project_name="networking",
    stack_name="prod")
shared_environment = pulumiservice.get_environment_output(organization_name=organization_name,
    project_name="platform",
    name="shared")
aws_account = pulumiservice.get_insights_account_output(organization_name=organization_name,
    account_name="aws-production")
platform_permissions = pulumiservice.build_role_permissions_output(organization_scopes=[
        pulumiservice.RbacOrganizationScope.TEAM_READ,
        pulumiservice.RbacOrganizationScope.STACK_CREATE,
    ],
    stack_rules=[
        # Read every stack in the organization
        {
            "all": True,
            "scopes": [pulumiservice.RbacStackScope.STACK_READ],
        },
        # Deploy one specific stack
        {
            "id": prod_stack.stack_id,
            "scopes": [
                pulumiservice.RbacStackScope.STACK_WRITE,
                pulumiservice.RbacStackScope.STACK_DEPLOYMENT_CREATE,
            ],
        },
        # Deploy stacks tagged team=platform (the operator defaults to equals)
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
        # Deploy stacks whose env tag is anything but "prod"
        {
            "tags": [{
                "key": "env",
                "value": "prod",
                "operator": pulumiservice.TagOperator.NOT_EQUALS,
            }],
            "scopes": [
                pulumiservice.RbacStackScope.STACK_WRITE,
                pulumiservice.RbacStackScope.STACK_DEPLOYMENT_CREATE,
            ],
        },
        # Read deployment settings on stacks that have a cost-center tag, whatever its value
        {
            "tags": [{
                "key": "cost-center",
                "operator": pulumiservice.TagOperator.EXISTS,
            }],
            "scopes": [pulumiservice.RbacStackScope.STACK_DEPLOYMENT_SETTINGS_READ],
        },
        # Tag stacks that don't have an owner tag yet
        {
            "tags": [{
                "key": "owner",
                "operator": pulumiservice.TagOperator.NOT_EXISTS,
            }],
            "scopes": [pulumiservice.RbacStackScope.STACK_TAGS_UPDATE],
        },
        # Delete stacks that match every tag condition
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
    ],
    environment_rules=[
        # Read every environment
        {
            "all": True,
            "scopes": [pulumiservice.RbacEnvironmentScope.ENVIRONMENT_READ],
        },
        # Open, update, and rotate one specific environment
        {
            "id": shared_environment.environment_id,
            "scopes": [
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_OPEN,
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_WRITE,
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_ROTATE,
            ],
        },
        # Open environments tagged team=platform
        {
            "tags": [{
                "key": "team",
                "value": "platform",
            }],
            "scopes": [pulumiservice.RbacEnvironmentScope.ENVIRONMENT_OPEN],
        },
    ],
    account_rules=[
        # Read every cloud account
        {
            "all": True,
            "scopes": [pulumiservice.RbacAccountScope.INSIGHTS_ACCOUNT_READ],
        },
        # Scan one specific account
        {
            "id": aws_account.insights_account_id,
            "scopes": [pulumiservice.RbacAccountScope.INSIGHTS_ACCOUNT_SCAN],
        },
        # Update accounts tagged team=platform
        {
            "tags": [{
                "key": "team",
                "value": "platform",
            }],
            "scopes": [pulumiservice.RbacAccountScope.INSIGHTS_ACCOUNT_UPDATE],
        },
    ])
platform_role = pulumiservice.OrganizationRole("platformRole",
    organization_name=organization_name,
    name="Platform Engineers",
    description="Read access across the organization, plus deploy access to platform stacks.",
    permissions=platform_permissions.permissions)

```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		organizationName := "my-org"
		if param := cfg.Get("organizationName"); param != "" {
			organizationName = param
		}
		prodStack := pulumiservice.GetStackOutput(ctx, pulumiservice.GetStackOutputArgs{
			OrganizationName: pulumi.String(organizationName),
			ProjectName:      pulumi.String("networking"),
			StackName:        pulumi.String("prod"),
		}, nil)
		sharedEnvironment := pulumiservice.GetEnvironmentOutput(ctx, pulumiservice.GetEnvironmentOutputArgs{
			OrganizationName: pulumi.String(organizationName),
			ProjectName:      pulumi.String("platform"),
			Name:             pulumi.String("shared"),
		}, nil)
		awsAccount := pulumiservice.GetInsightsAccountOutput(ctx, pulumiservice.GetInsightsAccountOutputArgs{
			OrganizationName: pulumi.String(organizationName),
			AccountName:      pulumi.String("aws-production"),
		}, nil)
		platformPermissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			OrganizationScopes: pulumiservice.RbacOrganizationScopeArray{
				pulumiservice.RbacOrganizationScopeTeamRead,
				pulumiservice.RbacOrganizationScopeStackCreate,
			},
			StackRules: pulumiservice.RoleStackRuleArray{
				// Read every stack in the organization
				&pulumiservice.RoleStackRuleArgs{
					All: pulumi.Bool(true),
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackRead,
					},
				},
				// Deploy one specific stack
				&pulumiservice.RoleStackRuleArgs{
					Id: prodStack.StackId(),
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackWrite,
						pulumiservice.RbacStackScopeStackDeploymentCreate,
					},
				},
				// Deploy stacks tagged team=platform (the operator defaults to equals)
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
				// Deploy stacks whose env tag is anything but "prod"
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:      pulumi.String("env"),
							Value:    pulumi.String("prod"),
							Operator: pulumiservice.TagOperatorNotEquals,
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackWrite,
						pulumiservice.RbacStackScopeStackDeploymentCreate,
					},
				},
				// Read deployment settings on stacks that have a cost-center tag, whatever its value
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
				// Tag stacks that don't have an owner tag yet
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
				// Delete stacks that match every tag condition
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
			EnvironmentRules: pulumiservice.RoleEnvironmentRuleArray{
				// Read every environment
				&pulumiservice.RoleEnvironmentRuleArgs{
					All: pulumi.Bool(true),
					Scopes: pulumiservice.RbacEnvironmentScopeArray{
						pulumiservice.RbacEnvironmentScopeEnvironmentRead,
					},
				},
				// Open, update, and rotate one specific environment
				&pulumiservice.RoleEnvironmentRuleArgs{
					Id: sharedEnvironment.EnvironmentId(),
					Scopes: pulumiservice.RbacEnvironmentScopeArray{
						pulumiservice.RbacEnvironmentScopeEnvironmentOpen,
						pulumiservice.RbacEnvironmentScopeEnvironmentWrite,
						pulumiservice.RbacEnvironmentScopeEnvironmentRotate,
					},
				},
				// Open environments tagged team=platform
				&pulumiservice.RoleEnvironmentRuleArgs{
					Tags: pulumiservice.TagConditionArray{
						&pulumiservice.TagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
					},
					Scopes: pulumiservice.RbacEnvironmentScopeArray{
						pulumiservice.RbacEnvironmentScopeEnvironmentOpen,
					},
				},
			},
			AccountRules: pulumiservice.RoleAccountRuleArray{
				// Read every cloud account
				&pulumiservice.RoleAccountRuleArgs{
					All: pulumi.Bool(true),
					Scopes: pulumiservice.RbacAccountScopeArray{
						pulumiservice.RbacAccountScopeInsightsAccountRead,
					},
				},
				// Scan one specific account
				&pulumiservice.RoleAccountRuleArgs{
					Id: awsAccount.InsightsAccountId(),
					Scopes: pulumiservice.RbacAccountScopeArray{
						pulumiservice.RbacAccountScopeInsightsAccountScan,
					},
				},
				// Update accounts tagged team=platform
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
		_, err := pulumiservice.NewOrganizationRole(ctx, "platformRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("Platform Engineers"),
			Description:      pulumi.String("Read access across the organization, plus deploy access to platform stacks."),
			Permissions:      platformPermissions.Permissions(),
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
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "my-org";
    var prodStack = PulumiService.GetStack.Invoke(new()
    {
        OrganizationName = organizationName,
        ProjectName = "networking",
        StackName = "prod",
    });

    var sharedEnvironment = PulumiService.GetEnvironment.Invoke(new()
    {
        OrganizationName = organizationName,
        ProjectName = "platform",
        Name = "shared",
    });

    var awsAccount = PulumiService.GetInsightsAccount.Invoke(new()
    {
        OrganizationName = organizationName,
        AccountName = "aws-production",
    });

    var platformPermissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        OrganizationScopes = new[]
        {
            PulumiService.RbacOrganizationScope.TeamRead,
            PulumiService.RbacOrganizationScope.StackCreate,
        },
        StackRules = new[]
        {
            // Read every stack in the organization
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                All = true,
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackRead,
                },
            },
            // Deploy one specific stack
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Id = prodStack.Apply(getStackResult => getStackResult.StackId),
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackWrite,
                    PulumiService.RbacStackScope.StackDeploymentCreate,
                },
            },
            // Deploy stacks tagged team=platform (the operator defaults to equals)
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
            // Deploy stacks whose env tag is anything but "prod"
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
                    PulumiService.RbacStackScope.StackWrite,
                    PulumiService.RbacStackScope.StackDeploymentCreate,
                },
            },
            // Read deployment settings on stacks that have a cost-center tag, whatever its value
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
            // Tag stacks that don't have an owner tag yet
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
            // Delete stacks that match every tag condition
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
        EnvironmentRules = new[]
        {
            // Read every environment
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
            {
                All = true,
                Scopes = new[]
                {
                    PulumiService.RbacEnvironmentScope.EnvironmentRead,
                },
            },
            // Open, update, and rotate one specific environment
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
            // Open environments tagged team=platform
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
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
                    PulumiService.RbacEnvironmentScope.EnvironmentOpen,
                },
            },
        },
        AccountRules = new[]
        {
            // Read every cloud account
            new PulumiService.Inputs.RoleAccountRuleInputArgs
            {
                All = true,
                Scopes = new[]
                {
                    PulumiService.RbacAccountScope.InsightsAccountRead,
                },
            },
            // Scan one specific account
            new PulumiService.Inputs.RoleAccountRuleInputArgs
            {
                Id = awsAccount.Apply(getInsightsAccountResult => getInsightsAccountResult.InsightsAccountId),
                Scopes = new[]
                {
                    PulumiService.RbacAccountScope.InsightsAccountScan,
                },
            },
            // Update accounts tagged team=platform
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

    var platformRole = new PulumiService.OrganizationRole("platformRole", new()
    {
        OrganizationName = organizationName,
        Name = "Platform Engineers",
        Description = "Read access across the organization, plus deploy access to platform stacks.",
        Permissions = platformPermissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

});


```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.enums.RbacAccountScope;
import com.pulumi.pulumiservice.enums.RbacEnvironmentScope;
import com.pulumi.pulumiservice.enums.RbacOrganizationScope;
import com.pulumi.pulumiservice.enums.RbacStackScope;
import com.pulumi.pulumiservice.enums.TagOperator;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetStackArgs;
import com.pulumi.pulumiservice.inputs.GetEnvironmentArgs;
import com.pulumi.pulumiservice.inputs.GetInsightsAccountArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleStackRuleArgs;
import com.pulumi.pulumiservice.inputs.TagConditionArgs;
import com.pulumi.pulumiservice.inputs.RoleEnvironmentRuleArgs;
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
        final var config = ctx.config();
        final var organizationName = config.get("organizationName").orElse("my-org");
        final var prodStack = PulumiserviceFunctions.getStack(GetStackArgs.builder()
            .organizationName(organizationName)
            .projectName("networking")
            .stackName("prod")
            .build());

        final var sharedEnvironment = PulumiserviceFunctions.getEnvironment(GetEnvironmentArgs.builder()
            .organizationName(organizationName)
            .projectName("platform")
            .name("shared")
            .build());

        final var awsAccount = PulumiserviceFunctions.getInsightsAccount(GetInsightsAccountArgs.builder()
            .organizationName(organizationName)
            .accountName("aws-production")
            .build());

        final var platformPermissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .organizationScopes(            
                RbacOrganizationScope.TeamRead,
                RbacOrganizationScope.StackCreate)
            .stackRules(            
                // Read every stack in the organization
                RoleStackRuleArgs.builder()
                    .all(true)
                    .scopes(RbacStackScope.StackRead)
                    .build(),
                // Deploy one specific stack
                RoleStackRuleArgs.builder()
                    .id(prodStack.applyValue(_prodStack -> _prodStack.stackId()))
                    .scopes(                    
                        RbacStackScope.StackWrite,
                        RbacStackScope.StackDeploymentCreate)
                    .build(),
                // Deploy stacks tagged team=platform (the operator defaults to equals)
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(                    
                        RbacStackScope.StackWrite,
                        RbacStackScope.StackDeploymentCreate)
                    .build(),
                // Deploy stacks whose env tag is anything but "prod"
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("env")
                        .value("prod")
                        .operator(TagOperator.NotEquals)
                        .build())
                    .scopes(                    
                        RbacStackScope.StackWrite,
                        RbacStackScope.StackDeploymentCreate)
                    .build(),
                // Read deployment settings on stacks that have a cost-center tag, whatever its value
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("cost-center")
                        .operator(TagOperator.Exists)
                        .build())
                    .scopes(RbacStackScope.StackDeploymentSettingsRead)
                    .build(),
                // Tag stacks that don't have an owner tag yet
                RoleStackRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("owner")
                        .operator(TagOperator.NotExists)
                        .build())
                    .scopes(RbacStackScope.StackTagsUpdate)
                    .build(),
                // Delete stacks that match every tag condition
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
            .environmentRules(            
                // Read every environment
                RoleEnvironmentRuleArgs.builder()
                    .all(true)
                    .scopes(RbacEnvironmentScope.EnvironmentRead)
                    .build(),
                // Open, update, and rotate one specific environment
                RoleEnvironmentRuleArgs.builder()
                    .id(sharedEnvironment.applyValue(_sharedEnvironment -> _sharedEnvironment.environmentId()))
                    .scopes(                    
                        RbacEnvironmentScope.EnvironmentOpen,
                        RbacEnvironmentScope.EnvironmentWrite,
                        RbacEnvironmentScope.EnvironmentRotate)
                    .build(),
                // Open environments tagged team=platform
                RoleEnvironmentRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(RbacEnvironmentScope.EnvironmentOpen)
                    .build())
            .accountRules(            
                // Read every cloud account
                RoleAccountRuleArgs.builder()
                    .all(true)
                    .scopes(RbacAccountScope.InsightsAccountRead)
                    .build(),
                // Scan one specific account
                RoleAccountRuleArgs.builder()
                    .id(awsAccount.applyValue(_awsAccount -> _awsAccount.insightsAccountId()))
                    .scopes(RbacAccountScope.InsightsAccountScan)
                    .build(),
                // Update accounts tagged team=platform
                RoleAccountRuleArgs.builder()
                    .tags(TagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(RbacAccountScope.InsightsAccountUpdate)
                    .build())
            .build());

        var platformRole = new OrganizationRole("platformRole", OrganizationRoleArgs.builder()
            .organizationName(organizationName)
            .name("Platform Engineers")
            .description("Read access across the organization, plus deploy access to platform stacks.")
            .permissions(platformPermissions.applyValue(_platformPermissions -> _platformPermissions.permissions()))
            .build());

    }
}

```

```yaml
config:
  organizationName:
    type: string
    default: my-org

variables:
  # Look up the entities the rules below grant access to. They already exist.
  prodStack:
    fn::invoke:
      function: pulumiservice:getStack
      arguments:
        organizationName: ${organizationName}
        projectName: networking
        stackName: prod
  sharedEnvironment:
    fn::invoke:
      function: pulumiservice:getEnvironment
      arguments:
        organizationName: ${organizationName}
        projectName: platform
        name: shared
  awsAccount:
    fn::invoke:
      function: pulumiservice:getInsightsAccount
      arguments:
        organizationName: ${organizationName}
        accountName: aws-production

  # Organization-level access plus rules for stacks, environments, and cloud
  # accounts. Each rule list accepts only scopes for its own entity type.
  platformPermissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        organizationScopes:
          - team:read
          - stack:create
        stackRules:
          # Read every stack in the organization
          - all: true
            scopes:
              - stack:read
          # Deploy one specific stack
          - id: ${prodStack.stackId}
            scopes:
              - stack:write
              - stack_deployment:create
          # Deploy stacks tagged team=platform (the operator defaults to equals)
          - tags:
              - key: team
                value: platform
            scopes:
              - stack:write
              - stack_deployment:create
          # Deploy stacks whose env tag is anything but "prod"
          - tags:
              - key: env
                value: prod
                operator: notEquals
            scopes:
              - stack:write
              - stack_deployment:create
          # Read deployment settings on stacks that have a cost-center tag, whatever its value
          - tags:
              - key: cost-center
                operator: exists
            scopes:
              - stack_deployment_settings:read
          # Tag stacks that don't have an owner tag yet
          - tags:
              - key: owner
                operator: notExists
            scopes:
              - stack_tags:update
          # Delete stacks that match every tag condition
          - tags:
              - key: team
                value: platform
              - key: lifecycle
                value: ephemeral
            scopes:
              - stack:delete
        environmentRules:
          # Read every environment
          - all: true
            scopes:
              - environment:read
          # Open, update, and rotate one specific environment
          - id: ${sharedEnvironment.environmentId}
            scopes:
              - environment:open
              - environment:write
              - environment:rotate
          # Open environments tagged team=platform
          - tags:
              - key: team
                value: platform
            scopes:
              - environment:open
        accountRules:
          # Read every cloud account
          - all: true
            scopes:
              - insights_account:read
          # Scan one specific account
          - id: ${awsAccount.insightsAccountId}
            scopes:
              - insights_account:scan
          # Update accounts tagged team=platform
          - tags:
              - key: team
                value: platform
            scopes:
              - insights_account:update

resources:
  platformRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: ${organizationName}
      name: Platform Engineers
      description: Read access across the organization, plus deploy access to platform stacks.
      permissions: ${platformPermissions.permissions}
```

{{% /example %}}
{{% /examples %}}
