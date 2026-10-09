{{% examples %}}
## Example Usage

{{% example %}}
### A role with organization-level access and entity rules

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") || "my-org";
const currentUser = pulumiservice.getCurrentUserOutput({});
const orgReadOnly = pulumiservice.getRbacPermissionSetOutput({
    organizationName: organizationName,
    defaultIdentifier: "org-settings-read-only",
});
const stackRead = pulumiservice.getRbacPermissionSetOutput({
    organizationName: organizationName,
    defaultIdentifier: "stack-read",
});
const envRead = pulumiservice.getRbacPermissionSetOutput({
    organizationName: organizationName,
    defaultIdentifier: "environment-read",
});
const deployer = new pulumiservice.RbacPermissionSet("deployer", {
    organizationName: organizationName,
    name: "Stack Deployer",
    description: "Read, update, and deploy stacks.",
    resourceType: pulumiservice.RbacResourceType.Stack,
    permissions: [
        pulumiservice.RbacScope.StackRead,
        pulumiservice.RbacScope.StackWrite,
        pulumiservice.RbacScope.StackDeploymentCreate,
    ],
});
const prodStack = new pulumiservice.Stack("prodStack", {
    organizationName: organizationName,
    projectName: "rbac-example",
    stackName: "prod",
});
const sharedEnv = new pulumiservice.Environment("sharedEnv", {
    organization: organizationName,
    project: "rbac-example",
    name: "shared",
    yaml: new pulumi.asset.StringAsset(`values:
  region: us-west-2`),
});
const platformRole = new pulumiservice.RbacRole("platformRole", {
    organizationName: organizationName,
    name: "Platform Engineers",
    description: "Deploys platform stacks and reads shared configuration.",
    organizationPermissionSetIds: [orgReadOnly.permissionSetId],
    entityRules: [
        {
            permissionSetIds: [stackRead.permissionSetId],
            stack: {
                all: true,
            },
        },
        {
            permissionSetIds: [deployer.permissionSetId],
            stack: {
                id: prodStack.stackId,
            },
        },
        {
            permissionSetIds: [deployer.permissionSetId],
            stack: {
                tags: [
                    {
                        key: "team",
                        value: "platform",
                    },
                    {
                        key: "env",
                        value: "prod",
                        operator: pulumiservice.RbacTagOperator.NotEquals,
                    },
                ],
            },
        },
        {
            permissionSetIds: [envRead.permissionSetId],
            environment: {
                id: sharedEnv.environmentId,
            },
        },
    ],
});
const platformTeam = new pulumiservice.Team("platformTeam", {
    organizationName: organizationName,
    name: "platform",
    teamType: "pulumi",
    displayName: "Platform",
    members: [currentUser.username],
});
const platformTeamRole = new pulumiservice.TeamRoleAssignment("platformTeamRole", {
    organizationName: organizationName,
    teamName: platformTeam.name,
    roleId: platformRole.roleId,
});
export const roleId = platformRole.roleId;
export const policyId = platformRole.policyId;

```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

config = pulumi.Config()
organization_name = config.get("organizationName")
if organization_name is None:
    organization_name = "my-org"
current_user = pulumiservice.get_current_user_output()
org_read_only = pulumiservice.get_rbac_permission_set_output(organization_name=organization_name,
    default_identifier="org-settings-read-only")
stack_read = pulumiservice.get_rbac_permission_set_output(organization_name=organization_name,
    default_identifier="stack-read")
env_read = pulumiservice.get_rbac_permission_set_output(organization_name=organization_name,
    default_identifier="environment-read")
deployer = pulumiservice.RbacPermissionSet("deployer",
    organization_name=organization_name,
    name="Stack Deployer",
    description="Read, update, and deploy stacks.",
    resource_type=pulumiservice.RbacResourceType.STACK,
    permissions=[
        pulumiservice.RbacScope.STACK_READ,
        pulumiservice.RbacScope.STACK_WRITE,
        pulumiservice.RbacScope.STACK_DEPLOYMENT_CREATE,
    ])
prod_stack = pulumiservice.Stack("prodStack",
    organization_name=organization_name,
    project_name="rbac-example",
    stack_name="prod")
shared_env = pulumiservice.Environment("sharedEnv",
    organization=organization_name,
    project="rbac-example",
    name="shared",
    yaml=pulumi.StringAsset("""values:
  region: us-west-2"""))
platform_role = pulumiservice.RbacRole("platformRole",
    organization_name=organization_name,
    name="Platform Engineers",
    description="Deploys platform stacks and reads shared configuration.",
    organization_permission_set_ids=[org_read_only.permission_set_id],
    entity_rules=[
        {
            "permission_set_ids": [stack_read.permission_set_id],
            "stack": {
                "all": True,
            },
        },
        {
            "permission_set_ids": [deployer.permission_set_id],
            "stack": {
                "id": prod_stack.stack_id,
            },
        },
        {
            "permission_set_ids": [deployer.permission_set_id],
            "stack": {
                "tags": [
                    {
                        "key": "team",
                        "value": "platform",
                    },
                    {
                        "key": "env",
                        "value": "prod",
                        "operator": pulumiservice.RbacTagOperator.NOT_EQUALS,
                    },
                ],
            },
        },
        {
            "permission_set_ids": [env_read.permission_set_id],
            "environment": {
                "id": shared_env.environment_id,
            },
        },
    ])
platform_team = pulumiservice.Team("platformTeam",
    organization_name=organization_name,
    name="platform",
    team_type="pulumi",
    display_name="Platform",
    members=[current_user.username])
platform_team_role = pulumiservice.TeamRoleAssignment("platformTeamRole",
    organization_name=organization_name,
    team_name=platform_team.name,
    role_id=platform_role.role_id)
pulumi.export("roleId", platform_role.role_id)
pulumi.export("policyId", platform_role.policy_id)

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
		currentUser := pulumiservice.GetCurrentUserOutput(ctx, pulumiservice.GetCurrentUserOutputArgs{}, nil)
		orgReadOnly := pulumiservice.GetRbacPermissionSetOutput(ctx, pulumiservice.GetRbacPermissionSetOutputArgs{
			OrganizationName:  pulumi.String(organizationName),
			DefaultIdentifier: pulumi.String("org-settings-read-only"),
		}, nil)
		stackRead := pulumiservice.GetRbacPermissionSetOutput(ctx, pulumiservice.GetRbacPermissionSetOutputArgs{
			OrganizationName:  pulumi.String(organizationName),
			DefaultIdentifier: pulumi.String("stack-read"),
		}, nil)
		envRead := pulumiservice.GetRbacPermissionSetOutput(ctx, pulumiservice.GetRbacPermissionSetOutputArgs{
			OrganizationName:  pulumi.String(organizationName),
			DefaultIdentifier: pulumi.String("environment-read"),
		}, nil)
		deployer, err := pulumiservice.NewRbacPermissionSet(ctx, "deployer", &pulumiservice.RbacPermissionSetArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("Stack Deployer"),
			Description:      pulumi.String("Read, update, and deploy stacks."),
			ResourceType:     pulumiservice.RbacResourceTypeStack,
			Permissions: pulumiservice.RbacScopeArray{
				pulumiservice.RbacScopeStackRead,
				pulumiservice.RbacScopeStackWrite,
				pulumiservice.RbacScopeStackDeploymentCreate,
			},
		})
		if err != nil {
			return err
		}
		prodStack, err := pulumiservice.NewStack(ctx, "prodStack", &pulumiservice.StackArgs{
			OrganizationName: pulumi.String(organizationName),
			ProjectName:      pulumi.String("rbac-example"),
			StackName:        pulumi.String("prod"),
		})
		if err != nil {
			return err
		}
		sharedEnv, err := pulumiservice.NewEnvironment(ctx, "sharedEnv", &pulumiservice.EnvironmentArgs{
			Organization: pulumi.String(organizationName),
			Project:      pulumi.String("rbac-example"),
			Name:         pulumi.String("shared"),
			Yaml:         pulumi.NewStringAsset("values:\n  region: us-west-2"),
		})
		if err != nil {
			return err
		}
		platformRole, err := pulumiservice.NewRbacRole(ctx, "platformRole", &pulumiservice.RbacRoleArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("Platform Engineers"),
			Description:      pulumi.String("Deploys platform stacks and reads shared configuration."),
			OrganizationPermissionSetIds: pulumi.StringArray{
				orgReadOnly.PermissionSetId(),
			},
			EntityRules: pulumiservice.RbacEntityRuleArray{
				&pulumiservice.RbacEntityRuleArgs{
					PermissionSetIds: pulumi.StringArray{
						stackRead.PermissionSetId(),
					},
					Stack: &pulumiservice.RbacEntitySelectorArgs{
						All: pulumi.Bool(true),
					},
				},
				&pulumiservice.RbacEntityRuleArgs{
					PermissionSetIds: pulumi.StringArray{
						deployer.PermissionSetId,
					},
					Stack: &pulumiservice.RbacEntitySelectorArgs{
						Id: prodStack.StackId,
					},
				},
				&pulumiservice.RbacEntityRuleArgs{
					PermissionSetIds: pulumi.StringArray{
						deployer.PermissionSetId,
					},
					Stack: &pulumiservice.RbacEntitySelectorArgs{
						Tags: pulumiservice.RbacTagConditionArray{
							&pulumiservice.RbacTagConditionArgs{
								Key:   pulumi.String("team"),
								Value: pulumi.String("platform"),
							},
							&pulumiservice.RbacTagConditionArgs{
								Key:      pulumi.String("env"),
								Value:    pulumi.String("prod"),
								Operator: pulumiservice.RbacTagOperatorNotEquals,
							},
						},
					},
				},
				&pulumiservice.RbacEntityRuleArgs{
					PermissionSetIds: pulumi.StringArray{
						envRead.PermissionSetId(),
					},
					Environment: &pulumiservice.RbacEntitySelectorArgs{
						Id: sharedEnv.EnvironmentId,
					},
				},
			},
		})
		if err != nil {
			return err
		}
		platformTeam, err := pulumiservice.NewTeam(ctx, "platformTeam", &pulumiservice.TeamArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("platform"),
			TeamType:         pulumi.String("pulumi"),
			DisplayName:      pulumi.String("Platform"),
			Members: pulumi.StringArray{
				currentUser.Username(),
			},
		})
		if err != nil {
			return err
		}
		_, err = pulumiservice.NewTeamRoleAssignment(ctx, "platformTeamRole", &pulumiservice.TeamRoleAssignmentArgs{
			OrganizationName: pulumi.String(organizationName),
			TeamName:         platformTeam.Name,
			RoleId:           platformRole.RoleId,
		})
		if err != nil {
			return err
		}
		ctx.Export("roleId", platformRole.RoleId)
		ctx.Export("policyId", platformRole.PolicyId)
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
    var currentUser = PulumiService.GetCurrentUser.Invoke();

    var orgReadOnly = PulumiService.GetRbacPermissionSet.Invoke(new()
    {
        OrganizationName = organizationName,
        DefaultIdentifier = "org-settings-read-only",
    });

    var stackRead = PulumiService.GetRbacPermissionSet.Invoke(new()
    {
        OrganizationName = organizationName,
        DefaultIdentifier = "stack-read",
    });

    var envRead = PulumiService.GetRbacPermissionSet.Invoke(new()
    {
        OrganizationName = organizationName,
        DefaultIdentifier = "environment-read",
    });

    var deployer = new PulumiService.RbacPermissionSet("deployer", new()
    {
        OrganizationName = organizationName,
        Name = "Stack Deployer",
        Description = "Read, update, and deploy stacks.",
        ResourceType = PulumiService.RbacResourceType.Stack,
        Permissions = new[]
        {
            PulumiService.RbacScope.StackRead,
            PulumiService.RbacScope.StackWrite,
            PulumiService.RbacScope.StackDeploymentCreate,
        },
    });

    var prodStack = new PulumiService.Stack("prodStack", new()
    {
        OrganizationName = organizationName,
        ProjectName = "rbac-example",
        StackName = "prod",
    });

    var sharedEnv = new PulumiService.Environment("sharedEnv", new()
    {
        Organization = organizationName,
        Project = "rbac-example",
        Name = "shared",
        Yaml = new StringAsset(@"values:
  region: us-west-2"),
    });

    var platformRole = new PulumiService.RbacRole("platformRole", new()
    {
        OrganizationName = organizationName,
        Name = "Platform Engineers",
        Description = "Deploys platform stacks and reads shared configuration.",
        OrganizationPermissionSetIds = new[]
        {
            orgReadOnly.Apply(getRbacPermissionSetResult => getRbacPermissionSetResult.PermissionSetId),
        },
        EntityRules = new[]
        {
            new PulumiService.Inputs.RbacEntityRuleArgs
            {
                PermissionSetIds = new[]
                {
                    stackRead.Apply(getRbacPermissionSetResult => getRbacPermissionSetResult.PermissionSetId),
                },
                Stack = new PulumiService.Inputs.RbacEntitySelectorArgs
                {
                    All = true,
                },
            },
            new PulumiService.Inputs.RbacEntityRuleArgs
            {
                PermissionSetIds = new[]
                {
                    deployer.PermissionSetId,
                },
                Stack = new PulumiService.Inputs.RbacEntitySelectorArgs
                {
                    Id = prodStack.StackId,
                },
            },
            new PulumiService.Inputs.RbacEntityRuleArgs
            {
                PermissionSetIds = new[]
                {
                    deployer.PermissionSetId,
                },
                Stack = new PulumiService.Inputs.RbacEntitySelectorArgs
                {
                    Tags = new[]
                    {
                        new PulumiService.Inputs.RbacTagConditionArgs
                        {
                            Key = "team",
                            Value = "platform",
                        },
                        new PulumiService.Inputs.RbacTagConditionArgs
                        {
                            Key = "env",
                            Value = "prod",
                            Operator = PulumiService.RbacTagOperator.NotEquals,
                        },
                    },
                },
            },
            new PulumiService.Inputs.RbacEntityRuleArgs
            {
                PermissionSetIds = new[]
                {
                    envRead.Apply(getRbacPermissionSetResult => getRbacPermissionSetResult.PermissionSetId),
                },
                Environment = new PulumiService.Inputs.RbacEntitySelectorArgs
                {
                    Id = sharedEnv.EnvironmentId,
                },
            },
        },
    });

    var platformTeam = new PulumiService.Team("platformTeam", new()
    {
        OrganizationName = organizationName,
        Name = "platform",
        TeamType = "pulumi",
        DisplayName = "Platform",
        Members = new[]
        {
            currentUser.Apply(getCurrentUserResult => getCurrentUserResult.Username),
        },
    });

    var platformTeamRole = new PulumiService.TeamRoleAssignment("platformTeamRole", new()
    {
        OrganizationName = organizationName,
        TeamName = platformTeam.Name,
        RoleId = platformRole.RoleId,
    });

    return new Dictionary<string, object?>
    {
        ["roleId"] = platformRole.RoleId,
        ["policyId"] = platformRole.PolicyId,
    };
});


```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetCurrentUserArgs;
import com.pulumi.pulumiservice.inputs.GetRbacPermissionSetArgs;
import com.pulumi.pulumiservice.RbacPermissionSet;
import com.pulumi.pulumiservice.RbacPermissionSetArgs;
import com.pulumi.pulumiservice.Stack;
import com.pulumi.pulumiservice.StackArgs;
import com.pulumi.pulumiservice.Environment;
import com.pulumi.pulumiservice.EnvironmentArgs;
import com.pulumi.pulumiservice.RbacRole;
import com.pulumi.pulumiservice.RbacRoleArgs;
import com.pulumi.pulumiservice.inputs.RbacEntityRuleArgs;
import com.pulumi.pulumiservice.inputs.RbacEntitySelectorArgs;
import com.pulumi.pulumiservice.inputs.RbacTagConditionArgs;
import com.pulumi.pulumiservice.Team;
import com.pulumi.pulumiservice.TeamArgs;
import com.pulumi.pulumiservice.TeamRoleAssignment;
import com.pulumi.pulumiservice.TeamRoleAssignmentArgs;
import com.pulumi.asset.StringAsset;
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
        final var currentUser = PulumiserviceFunctions.getCurrentUser(GetCurrentUserArgs.builder()
            .build());

        final var orgReadOnly = PulumiserviceFunctions.getRbacPermissionSet(GetRbacPermissionSetArgs.builder()
            .organizationName(organizationName)
            .defaultIdentifier("org-settings-read-only")
            .build());

        final var stackRead = PulumiserviceFunctions.getRbacPermissionSet(GetRbacPermissionSetArgs.builder()
            .organizationName(organizationName)
            .defaultIdentifier("stack-read")
            .build());

        final var envRead = PulumiserviceFunctions.getRbacPermissionSet(GetRbacPermissionSetArgs.builder()
            .organizationName(organizationName)
            .defaultIdentifier("environment-read")
            .build());

        var deployer = new RbacPermissionSet("deployer", RbacPermissionSetArgs.builder()
            .organizationName(organizationName)
            .name("Stack Deployer")
            .description("Read, update, and deploy stacks.")
            .resourceType("stack")
            .permissions(            
                "stack:read",
                "stack:write",
                "stack_deployment:create")
            .build());

        var prodStack = new Stack("prodStack", StackArgs.builder()
            .organizationName(organizationName)
            .projectName("rbac-example")
            .stackName("prod")
            .build());

        var sharedEnv = new Environment("sharedEnv", EnvironmentArgs.builder()
            .organization(organizationName)
            .project("rbac-example")
            .name("shared")
            .yaml(new StringAsset("""
values:
  region: us-west-2            """))
            .build());

        var platformRole = new RbacRole("platformRole", RbacRoleArgs.builder()
            .organizationName(organizationName)
            .name("Platform Engineers")
            .description("Deploys platform stacks and reads shared configuration.")
            .organizationPermissionSetIds(orgReadOnly.applyValue(_orgReadOnly -> _orgReadOnly.permissionSetId()))
            .entityRules(            
                RbacEntityRuleArgs.builder()
                    .permissionSetIds(stackRead.applyValue(_stackRead -> _stackRead.permissionSetId()))
                    .stack(RbacEntitySelectorArgs.builder()
                        .all(true)
                        .build())
                    .build(),
                RbacEntityRuleArgs.builder()
                    .permissionSetIds(deployer.permissionSetId())
                    .stack(RbacEntitySelectorArgs.builder()
                        .id(prodStack.stackId())
                        .build())
                    .build(),
                RbacEntityRuleArgs.builder()
                    .permissionSetIds(deployer.permissionSetId())
                    .stack(RbacEntitySelectorArgs.builder()
                        .tags(                        
                            RbacTagConditionArgs.builder()
                                .key("team")
                                .value("platform")
                                .build(),
                            RbacTagConditionArgs.builder()
                                .key("env")
                                .value("prod")
                                .operator("notEquals")
                                .build())
                        .build())
                    .build(),
                RbacEntityRuleArgs.builder()
                    .permissionSetIds(envRead.applyValue(_envRead -> _envRead.permissionSetId()))
                    .environment(RbacEntitySelectorArgs.builder()
                        .id(sharedEnv.environmentId())
                        .build())
                    .build())
            .build());

        var platformTeam = new Team("platformTeam", TeamArgs.builder()
            .organizationName(organizationName)
            .name("platform")
            .teamType("pulumi")
            .displayName("Platform")
            .members(currentUser.applyValue(_currentUser -> _currentUser.username()))
            .build());

        var platformTeamRole = new TeamRoleAssignment("platformTeamRole", TeamRoleAssignmentArgs.builder()
            .organizationName(organizationName)
            .teamName(platformTeam.name())
            .roleId(platformRole.roleId())
            .build());

        ctx.export("roleId", platformRole.roleId());
        ctx.export("policyId", platformRole.policyId());
    }
}

```

```yaml
config:
  organizationName:
    type: string
    default: my-org

resources:
  # A custom permission set for deploying stacks.
  deployer:
    type: pulumiservice:RbacPermissionSet
    properties:
      organizationName: ${organizationName}
      name: Stack Deployer
      description: Read, update, and deploy stacks.
      resourceType: stack
      permissions:
        - stack:read
        - stack:write
        - stack_deployment:create

  prodStack:
    type: pulumiservice:Stack
    properties:
      organizationName: ${organizationName}
      projectName: rbac-example
      stackName: prod

  sharedEnv:
    type: pulumiservice:Environment
    properties:
      organization: ${organizationName}
      project: rbac-example
      name: shared
      yaml:
        fn::stringAsset: |-
          values:
            region: us-west-2

  # A role laid out like the console: organization-level access plus entity rules.
  platformRole:
    type: pulumiservice:RbacRole
    properties:
      organizationName: ${organizationName}
      name: Platform Engineers
      description: Deploys platform stacks and reads shared configuration.
      organizationPermissionSetIds:
        - ${orgReadOnly.permissionSetId}
      entityRules:
        # Read every stack in the organization.
        - permissionSetIds: ["${stackRead.permissionSetId}"]
          stack:
            all: true
        # Deploy one specific stack.
        - permissionSetIds: ["${deployer.permissionSetId}"]
          stack:
            id: ${prodStack.stackId}
        # Deploy stacks tagged team=platform, except those tagged env=prod.
        - permissionSetIds: ["${deployer.permissionSetId}"]
          stack:
            tags:
              - key: team
                value: platform
              - key: env
                value: prod
                operator: notEquals
        # Read one environment.
        - permissionSetIds: ["${envRead.permissionSetId}"]
          environment:
            id: ${sharedEnv.environmentId}

  platformTeam:
    type: pulumiservice:Team
    properties:
      organizationName: ${organizationName}
      name: platform
      teamType: pulumi
      displayName: Platform
      members:
        - ${currentUser.username}

  platformTeamRole:
    type: pulumiservice:TeamRoleAssignment
    properties:
      organizationName: ${organizationName}
      teamName: ${platformTeam.name}
      roleId: ${platformRole.roleId}

variables:
  currentUser:
    fn::invoke:
      function: pulumiservice:getCurrentUser
  # Built-in permission sets are looked up, not recreated.
  orgReadOnly:
    fn::invoke:
      function: pulumiservice:getRbacPermissionSet
      arguments:
        organizationName: ${organizationName}
        defaultIdentifier: org-settings-read-only
  stackRead:
    fn::invoke:
      function: pulumiservice:getRbacPermissionSet
      arguments:
        organizationName: ${organizationName}
        defaultIdentifier: stack-read
  envRead:
    fn::invoke:
      function: pulumiservice:getRbacPermissionSet
      arguments:
        organizationName: ${organizationName}
        defaultIdentifier: environment-read

outputs:
  roleId: ${platformRole.roleId}
  policyId: ${platformRole.policyId}
```

{{% /example %}}
{{% /examples %}}
