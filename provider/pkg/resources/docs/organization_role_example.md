{{% examples %}}
## Example Usage

{{% example %}}
### A role with organization-level access and entity rules

The program looks up an existing stack, environment, and Insights account, then defines two roles: one granting scopes directly, and one granting permission sets through a policy.

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
const sharedEnvironmentRotate = pulumiservice.buildEnvironmentScopedPermissionsOutput({
    environmentId: sharedEnvironment.environmentId,
    permissions: ["environment:rotate"],
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
                operator: pulumiservice.RoleTagOperator.NotEquals,
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
            }],
            scopes: [pulumiservice.RbacStackScope.StackDeploymentSettingsRead],
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
        // Open and update one specific environment
        {
            id: sharedEnvironment.environmentId,
            scopes: [
                pulumiservice.RbacEnvironmentScope.EnvironmentOpen,
                pulumiservice.RbacEnvironmentScope.EnvironmentWrite,
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
    insightsAccountRules: [
        // Read every Insights account
        {
            all: true,
            scopes: [pulumiservice.RbacInsightsAccountScope.InsightsAccountRead],
        },
        // Scan one specific account
        {
            id: awsAccount.insightsAccountId,
            scopes: [pulumiservice.RbacInsightsAccountScope.InsightsAccountScan],
        },
        // Update accounts tagged team=platform
        {
            tags: [{
                key: "team",
                value: "platform",
            }],
            scopes: [pulumiservice.RbacInsightsAccountScope.InsightsAccountUpdate],
        },
    ],
    additionalEntries: [sharedEnvironmentRotate.permissions],
});
const orgReadOnly = pulumiservice.getOrganizationPermissionSetOutput({
    organizationName: organizationName,
    defaultIdentifier: "org-settings-read-only",
});
const stackWrite = pulumiservice.getOrganizationPermissionSetOutput({
    organizationName: organizationName,
    defaultIdentifier: "stack-write",
});
const environmentRead = pulumiservice.getOrganizationPermissionSetOutput({
    organizationName: organizationName,
    defaultIdentifier: "environment-read",
});
const insightsAccountRead = pulumiservice.getOrganizationPermissionSetOutput({
    organizationName: organizationName,
    defaultIdentifier: "insights-account-read",
});
const operatorPolicyPermissions = pulumiservice.buildRolePermissionsOutput({
    organizationPermissionSets: [{
        permissionSetId: orgReadOnly.permissionSetId,
        resourceType: orgReadOnly.resourceType,
    }],
    // The Stack Write set on stacks tagged team=platform
    stackRules: [{
        tags: [{
            key: "team",
            value: "platform",
        }],
        permissionSets: [{
            permissionSetId: stackWrite.permissionSetId,
            resourceType: stackWrite.resourceType,
        }],
    }],
    // The Environment Read set on every environment
    environmentRules: [{
        all: true,
        permissionSets: [{
            permissionSetId: environmentRead.permissionSetId,
            resourceType: environmentRead.resourceType,
        }],
    }],
    // The Insights Account Read set, plus one extra scope, on one account
    insightsAccountRules: [{
        id: awsAccount.insightsAccountId,
        permissionSets: [{
            permissionSetId: insightsAccountRead.permissionSetId,
            resourceType: insightsAccountRead.resourceType,
        }],
        scopes: [pulumiservice.RbacInsightsAccountScope.InsightsAccountScan],
    }],
});
const operatorPolicy = new pulumiservice.api.Role("operatorPolicy", {
    orgName: organizationName,
    name: "Platform Operators policy",
    uxPurpose: "policy",
    details: operatorPolicyPermissions.permissions,
});
const operatorPermissions = pulumiservice.buildComposePermissionsOutput({
    ids: [operatorPolicy.roleID],
});
const platformRole = new pulumiservice.OrganizationRole("platformRole", {
    organizationName: organizationName,
    name: "Platform Engineers",
    description: "Read access across the organization, plus deploy access to platform stacks.",
    permissions: platformPermissions.permissions,
});
const operatorRole = new pulumiservice.OrganizationRole("operatorRole", {
    organizationName: organizationName,
    name: "Platform Operators",
    description: "Built-in permission sets on platform entities.",
    permissions: operatorPermissions.permissions,
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
shared_environment_rotate = pulumiservice.build_environment_scoped_permissions_output(environment_id=shared_environment.environment_id,
    permissions=["environment:rotate"])
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
                "operator": pulumiservice.RoleTagOperator.NOT_EQUALS,
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
            }],
            "scopes": [pulumiservice.RbacStackScope.STACK_DEPLOYMENT_SETTINGS_READ],
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
        # Open and update one specific environment
        {
            "id": shared_environment.environment_id,
            "scopes": [
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_OPEN,
                pulumiservice.RbacEnvironmentScope.ENVIRONMENT_WRITE,
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
    insights_account_rules=[
        # Read every Insights account
        {
            "all": True,
            "scopes": [pulumiservice.RbacInsightsAccountScope.INSIGHTS_ACCOUNT_READ],
        },
        # Scan one specific account
        {
            "id": aws_account.insights_account_id,
            "scopes": [pulumiservice.RbacInsightsAccountScope.INSIGHTS_ACCOUNT_SCAN],
        },
        # Update accounts tagged team=platform
        {
            "tags": [{
                "key": "team",
                "value": "platform",
            }],
            "scopes": [pulumiservice.RbacInsightsAccountScope.INSIGHTS_ACCOUNT_UPDATE],
        },
    ],
    additional_entries=[shared_environment_rotate.permissions])
org_read_only = pulumiservice.get_organization_permission_set_output(organization_name=organization_name,
    default_identifier="org-settings-read-only")
stack_write = pulumiservice.get_organization_permission_set_output(organization_name=organization_name,
    default_identifier="stack-write")
environment_read = pulumiservice.get_organization_permission_set_output(organization_name=organization_name,
    default_identifier="environment-read")
insights_account_read = pulumiservice.get_organization_permission_set_output(organization_name=organization_name,
    default_identifier="insights-account-read")
operator_policy_permissions = pulumiservice.build_role_permissions_output(organization_permission_sets=[{
        "permission_set_id": org_read_only.permission_set_id,
        "resource_type": org_read_only.resource_type,
    }],
    # The Stack Write set on stacks tagged team=platform
    stack_rules=[{
        "tags": [{
            "key": "team",
            "value": "platform",
        }],
        "permission_sets": [{
            "permission_set_id": stack_write.permission_set_id,
            "resource_type": stack_write.resource_type,
        }],
    }],
    # The Environment Read set on every environment
    environment_rules=[{
        "all": True,
        "permission_sets": [{
            "permission_set_id": environment_read.permission_set_id,
            "resource_type": environment_read.resource_type,
        }],
    }],
    # The Insights Account Read set, plus one extra scope, on one account
    insights_account_rules=[{
        "id": aws_account.insights_account_id,
        "permission_sets": [{
            "permission_set_id": insights_account_read.permission_set_id,
            "resource_type": insights_account_read.resource_type,
        }],
        "scopes": [pulumiservice.RbacInsightsAccountScope.INSIGHTS_ACCOUNT_SCAN],
    }])
operator_policy = pulumiservice.api.Role("operatorPolicy",
    org_name=organization_name,
    name="Platform Operators policy",
    ux_purpose="policy",
    details=operator_policy_permissions.permissions)
operator_permissions = pulumiservice.build_compose_permissions_output(ids=[operator_policy.role_id])
platform_role = pulumiservice.OrganizationRole("platformRole",
    organization_name=organization_name,
    name="Platform Engineers",
    description="Read access across the organization, plus deploy access to platform stacks.",
    permissions=platform_permissions.permissions)
operator_role = pulumiservice.OrganizationRole("operatorRole",
    organization_name=organization_name,
    name="Platform Operators",
    description="Built-in permission sets on platform entities.",
    permissions=operator_permissions.permissions)

```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
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
		sharedEnvironmentRotate := pulumiservice.BuildEnvironmentScopedPermissionsOutput(ctx, pulumiservice.BuildEnvironmentScopedPermissionsOutputArgs{
			EnvironmentId: sharedEnvironment.EnvironmentId(),
			Permissions: pulumi.StringArray{
				pulumi.String("environment:rotate"),
			},
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
					Tags: pulumiservice.RoleTagConditionArray{
						&pulumiservice.RoleTagConditionArgs{
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
					Tags: pulumiservice.RoleTagConditionArray{
						&pulumiservice.RoleTagConditionArgs{
							Key:      pulumi.String("env"),
							Value:    pulumi.String("prod"),
							Operator: pulumiservice.RoleTagOperatorNotEquals,
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackWrite,
						pulumiservice.RbacStackScopeStackDeploymentCreate,
					},
				},
				// Read deployment settings on stacks that have a cost-center tag, whatever its value
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.RoleTagConditionArray{
						&pulumiservice.RoleTagConditionArgs{
							Key: pulumi.String("cost-center"),
						},
					},
					Scopes: pulumiservice.RbacStackScopeArray{
						pulumiservice.RbacStackScopeStackDeploymentSettingsRead,
					},
				},
				// Delete stacks that match every tag condition
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.RoleTagConditionArray{
						&pulumiservice.RoleTagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
						&pulumiservice.RoleTagConditionArgs{
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
				// Open and update one specific environment
				&pulumiservice.RoleEnvironmentRuleArgs{
					Id: sharedEnvironment.EnvironmentId(),
					Scopes: pulumiservice.RbacEnvironmentScopeArray{
						pulumiservice.RbacEnvironmentScopeEnvironmentOpen,
						pulumiservice.RbacEnvironmentScopeEnvironmentWrite,
					},
				},
				// Open environments tagged team=platform
				&pulumiservice.RoleEnvironmentRuleArgs{
					Tags: pulumiservice.RoleTagConditionArray{
						&pulumiservice.RoleTagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
					},
					Scopes: pulumiservice.RbacEnvironmentScopeArray{
						pulumiservice.RbacEnvironmentScopeEnvironmentOpen,
					},
				},
			},
			InsightsAccountRules: pulumiservice.RoleInsightsAccountRuleArray{
				// Read every Insights account
				&pulumiservice.RoleInsightsAccountRuleArgs{
					All: pulumi.Bool(true),
					Scopes: pulumiservice.RbacInsightsAccountScopeArray{
						pulumiservice.RbacInsightsAccountScopeInsightsAccountRead,
					},
				},
				// Scan one specific account
				&pulumiservice.RoleInsightsAccountRuleArgs{
					Id: awsAccount.InsightsAccountId(),
					Scopes: pulumiservice.RbacInsightsAccountScopeArray{
						pulumiservice.RbacInsightsAccountScopeInsightsAccountScan,
					},
				},
				// Update accounts tagged team=platform
				&pulumiservice.RoleInsightsAccountRuleArgs{
					Tags: pulumiservice.RoleTagConditionArray{
						&pulumiservice.RoleTagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
					},
					Scopes: pulumiservice.RbacInsightsAccountScopeArray{
						pulumiservice.RbacInsightsAccountScopeInsightsAccountUpdate,
					},
				},
			},
			AdditionalEntries: pulumi.MapArray{
				sharedEnvironmentRotate.Permissions(),
			},
		}, nil)
		orgReadOnly := pulumiservice.GetOrganizationPermissionSetOutput(ctx, pulumiservice.GetOrganizationPermissionSetOutputArgs{
			OrganizationName:  pulumi.String(organizationName),
			DefaultIdentifier: pulumi.String("org-settings-read-only"),
		}, nil)
		stackWrite := pulumiservice.GetOrganizationPermissionSetOutput(ctx, pulumiservice.GetOrganizationPermissionSetOutputArgs{
			OrganizationName:  pulumi.String(organizationName),
			DefaultIdentifier: pulumi.String("stack-write"),
		}, nil)
		environmentRead := pulumiservice.GetOrganizationPermissionSetOutput(ctx, pulumiservice.GetOrganizationPermissionSetOutputArgs{
			OrganizationName:  pulumi.String(organizationName),
			DefaultIdentifier: pulumi.String("environment-read"),
		}, nil)
		insightsAccountRead := pulumiservice.GetOrganizationPermissionSetOutput(ctx, pulumiservice.GetOrganizationPermissionSetOutputArgs{
			OrganizationName:  pulumi.String(organizationName),
			DefaultIdentifier: pulumi.String("insights-account-read"),
		}, nil)
		operatorPolicyPermissions := pulumiservice.BuildRolePermissionsOutput(ctx, pulumiservice.BuildRolePermissionsOutputArgs{
			OrganizationPermissionSets: pulumiservice.RolePermissionSetRefArray{
				&pulumiservice.RolePermissionSetRefArgs{
					PermissionSetId: orgReadOnly.PermissionSetId(),
					ResourceType:    orgReadOnly.ResourceType(),
				},
			},
			StackRules: pulumiservice.RoleStackRuleArray{
				// The Stack Write set on stacks tagged team=platform
				&pulumiservice.RoleStackRuleArgs{
					Tags: pulumiservice.RoleTagConditionArray{
						&pulumiservice.RoleTagConditionArgs{
							Key:   pulumi.String("team"),
							Value: pulumi.String("platform"),
						},
					},
					PermissionSets: pulumiservice.RolePermissionSetRefArray{
						&pulumiservice.RolePermissionSetRefArgs{
							PermissionSetId: stackWrite.PermissionSetId(),
							ResourceType:    stackWrite.ResourceType(),
						},
					},
				},
			},
			EnvironmentRules: pulumiservice.RoleEnvironmentRuleArray{
				// The Environment Read set on every environment
				&pulumiservice.RoleEnvironmentRuleArgs{
					All: pulumi.Bool(true),
					PermissionSets: pulumiservice.RolePermissionSetRefArray{
						&pulumiservice.RolePermissionSetRefArgs{
							PermissionSetId: environmentRead.PermissionSetId(),
							ResourceType:    environmentRead.ResourceType(),
						},
					},
				},
			},
			InsightsAccountRules: pulumiservice.RoleInsightsAccountRuleArray{
				// The Insights Account Read set, plus one extra scope, on one account
				&pulumiservice.RoleInsightsAccountRuleArgs{
					Id: awsAccount.InsightsAccountId(),
					PermissionSets: pulumiservice.RolePermissionSetRefArray{
						&pulumiservice.RolePermissionSetRefArgs{
							PermissionSetId: insightsAccountRead.PermissionSetId(),
							ResourceType:    insightsAccountRead.ResourceType(),
						},
					},
					Scopes: pulumiservice.RbacInsightsAccountScopeArray{
						pulumiservice.RbacInsightsAccountScopeInsightsAccountScan,
					},
				},
			},
		}, nil)
		operatorPolicy, err := api.NewRole(ctx, "operatorPolicy", &api.RoleArgs{
			OrgName:   pulumi.String(organizationName),
			Name:      pulumi.String("Platform Operators policy"),
			UxPurpose: pulumi.String("policy"),
			Details:   interface{}(operatorPolicyPermissions.Permissions()),
		})
		if err != nil {
			return err
		}
		operatorPermissions := pulumiservice.BuildComposePermissionsOutput(ctx, pulumiservice.BuildComposePermissionsOutputArgs{
			Ids: pulumi.StringArray{
				operatorPolicy.RoleID,
			},
		}, nil)
		_, err = pulumiservice.NewOrganizationRole(ctx, "platformRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("Platform Engineers"),
			Description:      pulumi.String("Read access across the organization, plus deploy access to platform stacks."),
			Permissions:      platformPermissions.Permissions(),
		})
		if err != nil {
			return err
		}
		_, err = pulumiservice.NewOrganizationRole(ctx, "operatorRole", &pulumiservice.OrganizationRoleArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("Platform Operators"),
			Description:      pulumi.String("Built-in permission sets on platform entities."),
			Permissions:      operatorPermissions.Permissions(),
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

    var sharedEnvironmentRotate = PulumiService.BuildEnvironmentScopedPermissions.Invoke(new()
    {
        EnvironmentId = sharedEnvironment.Apply(getEnvironmentResult => getEnvironmentResult.EnvironmentId),
        Permissions = new[]
        {
            "environment:rotate",
        },
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
                    new PulumiService.Inputs.RoleTagConditionInputArgs
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
                    new PulumiService.Inputs.RoleTagConditionInputArgs
                    {
                        Key = "env",
                        Value = "prod",
                        Operator = PulumiService.RoleTagOperator.NotEquals,
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
                    new PulumiService.Inputs.RoleTagConditionInputArgs
                    {
                        Key = "cost-center",
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacStackScope.StackDeploymentSettingsRead,
                },
            },
            // Delete stacks that match every tag condition
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.RoleTagConditionInputArgs
                    {
                        Key = "team",
                        Value = "platform",
                    },
                    new PulumiService.Inputs.RoleTagConditionInputArgs
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
            // Open and update one specific environment
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
            {
                Id = sharedEnvironment.Apply(getEnvironmentResult => getEnvironmentResult.EnvironmentId),
                Scopes = new[]
                {
                    PulumiService.RbacEnvironmentScope.EnvironmentOpen,
                    PulumiService.RbacEnvironmentScope.EnvironmentWrite,
                },
            },
            // Open environments tagged team=platform
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.RoleTagConditionInputArgs
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
        InsightsAccountRules = new[]
        {
            // Read every Insights account
            new PulumiService.Inputs.RoleInsightsAccountRuleInputArgs
            {
                All = true,
                Scopes = new[]
                {
                    PulumiService.RbacInsightsAccountScope.InsightsAccountRead,
                },
            },
            // Scan one specific account
            new PulumiService.Inputs.RoleInsightsAccountRuleInputArgs
            {
                Id = awsAccount.Apply(getInsightsAccountResult => getInsightsAccountResult.InsightsAccountId),
                Scopes = new[]
                {
                    PulumiService.RbacInsightsAccountScope.InsightsAccountScan,
                },
            },
            // Update accounts tagged team=platform
            new PulumiService.Inputs.RoleInsightsAccountRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.RoleTagConditionInputArgs
                    {
                        Key = "team",
                        Value = "platform",
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacInsightsAccountScope.InsightsAccountUpdate,
                },
            },
        },
        AdditionalEntries = new[]
        {
            sharedEnvironmentRotate.Apply(buildEnvironmentScopedPermissionsResult => buildEnvironmentScopedPermissionsResult.Permissions),
        },
    });

    var orgReadOnly = PulumiService.GetOrganizationPermissionSet.Invoke(new()
    {
        OrganizationName = organizationName,
        DefaultIdentifier = "org-settings-read-only",
    });

    var stackWrite = PulumiService.GetOrganizationPermissionSet.Invoke(new()
    {
        OrganizationName = organizationName,
        DefaultIdentifier = "stack-write",
    });

    var environmentRead = PulumiService.GetOrganizationPermissionSet.Invoke(new()
    {
        OrganizationName = organizationName,
        DefaultIdentifier = "environment-read",
    });

    var insightsAccountRead = PulumiService.GetOrganizationPermissionSet.Invoke(new()
    {
        OrganizationName = organizationName,
        DefaultIdentifier = "insights-account-read",
    });

    var operatorPolicyPermissions = PulumiService.BuildRolePermissions.Invoke(new()
    {
        OrganizationPermissionSets = new[]
        {
            new PulumiService.Inputs.RolePermissionSetRefInputArgs
            {
                PermissionSetId = orgReadOnly.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.PermissionSetId),
                ResourceType = orgReadOnly.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.ResourceType),
            },
        },
        StackRules = new[]
        {
            // The Stack Write set on stacks tagged team=platform
            new PulumiService.Inputs.RoleStackRuleInputArgs
            {
                Tags = new[]
                {
                    new PulumiService.Inputs.RoleTagConditionInputArgs
                    {
                        Key = "team",
                        Value = "platform",
                    },
                },
                PermissionSets = new[]
                {
                    new PulumiService.Inputs.RolePermissionSetRefInputArgs
                    {
                        PermissionSetId = stackWrite.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.PermissionSetId),
                        ResourceType = stackWrite.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.ResourceType),
                    },
                },
            },
        },
        EnvironmentRules = new[]
        {
            // The Environment Read set on every environment
            new PulumiService.Inputs.RoleEnvironmentRuleInputArgs
            {
                All = true,
                PermissionSets = new[]
                {
                    new PulumiService.Inputs.RolePermissionSetRefInputArgs
                    {
                        PermissionSetId = environmentRead.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.PermissionSetId),
                        ResourceType = environmentRead.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.ResourceType),
                    },
                },
            },
        },
        InsightsAccountRules = new[]
        {
            // The Insights Account Read set, plus one extra scope, on one account
            new PulumiService.Inputs.RoleInsightsAccountRuleInputArgs
            {
                Id = awsAccount.Apply(getInsightsAccountResult => getInsightsAccountResult.InsightsAccountId),
                PermissionSets = new[]
                {
                    new PulumiService.Inputs.RolePermissionSetRefInputArgs
                    {
                        PermissionSetId = insightsAccountRead.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.PermissionSetId),
                        ResourceType = insightsAccountRead.Apply(getOrganizationPermissionSetResult => getOrganizationPermissionSetResult.ResourceType),
                    },
                },
                Scopes = new[]
                {
                    PulumiService.RbacInsightsAccountScope.InsightsAccountScan,
                },
            },
        },
    });

    var operatorPolicy = new PulumiService.Api.Role("operatorPolicy", new()
    {
        OrgName = organizationName,
        Name = "Platform Operators policy",
        UxPurpose = "policy",
        Details = operatorPolicyPermissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

    var operatorPermissions = PulumiService.BuildComposePermissions.Invoke(new()
    {
        Ids = new[]
        {
            operatorPolicy.RoleID,
        },
    });

    var platformRole = new PulumiService.OrganizationRole("platformRole", new()
    {
        OrganizationName = organizationName,
        Name = "Platform Engineers",
        Description = "Read access across the organization, plus deploy access to platform stacks.",
        Permissions = platformPermissions.Apply(buildRolePermissionsResult => buildRolePermissionsResult.Permissions),
    });

    var operatorRole = new PulumiService.OrganizationRole("operatorRole", new()
    {
        OrganizationName = organizationName,
        Name = "Platform Operators",
        Description = "Built-in permission sets on platform entities.",
        Permissions = operatorPermissions.Apply(buildComposePermissionsResult => buildComposePermissionsResult.Permissions),
    });

});


```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.enums.RbacEnvironmentScope;
import com.pulumi.pulumiservice.enums.RbacInsightsAccountScope;
import com.pulumi.pulumiservice.enums.RbacOrganizationScope;
import com.pulumi.pulumiservice.enums.RbacStackScope;
import com.pulumi.pulumiservice.enums.RoleTagOperator;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetStackArgs;
import com.pulumi.pulumiservice.inputs.GetEnvironmentArgs;
import com.pulumi.pulumiservice.inputs.GetInsightsAccountArgs;
import com.pulumi.pulumiservice.inputs.BuildEnvironmentScopedPermissionsArgs;
import com.pulumi.pulumiservice.inputs.BuildRolePermissionsArgs;
import com.pulumi.pulumiservice.inputs.RoleStackRuleArgs;
import com.pulumi.pulumiservice.inputs.RoleTagConditionArgs;
import com.pulumi.pulumiservice.inputs.RoleEnvironmentRuleArgs;
import com.pulumi.pulumiservice.inputs.RoleInsightsAccountRuleArgs;
import com.pulumi.pulumiservice.inputs.GetOrganizationPermissionSetArgs;
import com.pulumi.pulumiservice.inputs.RolePermissionSetRefArgs;
import com.pulumi.pulumiservice.api.Role;
import com.pulumi.pulumiservice.api.RoleArgs;
import com.pulumi.pulumiservice.inputs.BuildComposePermissionsArgs;
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

        final var sharedEnvironmentRotate = PulumiserviceFunctions.buildEnvironmentScopedPermissions(BuildEnvironmentScopedPermissionsArgs.builder()
            .environmentId(sharedEnvironment.applyValue(_sharedEnvironment -> _sharedEnvironment.environmentId()))
            .permissions("environment:rotate")
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
                    .tags(RoleTagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(                    
                        RbacStackScope.StackWrite,
                        RbacStackScope.StackDeploymentCreate)
                    .build(),
                // Deploy stacks whose env tag is anything but "prod"
                RoleStackRuleArgs.builder()
                    .tags(RoleTagConditionArgs.builder()
                        .key("env")
                        .value("prod")
                        .operator(RoleTagOperator.NotEquals)
                        .build())
                    .scopes(                    
                        RbacStackScope.StackWrite,
                        RbacStackScope.StackDeploymentCreate)
                    .build(),
                // Read deployment settings on stacks that have a cost-center tag, whatever its value
                RoleStackRuleArgs.builder()
                    .tags(RoleTagConditionArgs.builder()
                        .key("cost-center")
                        .build())
                    .scopes(RbacStackScope.StackDeploymentSettingsRead)
                    .build(),
                // Delete stacks that match every tag condition
                RoleStackRuleArgs.builder()
                    .tags(                    
                        RoleTagConditionArgs.builder()
                            .key("team")
                            .value("platform")
                            .build(),
                        RoleTagConditionArgs.builder()
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
                // Open and update one specific environment
                RoleEnvironmentRuleArgs.builder()
                    .id(sharedEnvironment.applyValue(_sharedEnvironment -> _sharedEnvironment.environmentId()))
                    .scopes(                    
                        RbacEnvironmentScope.EnvironmentOpen,
                        RbacEnvironmentScope.EnvironmentWrite)
                    .build(),
                // Open environments tagged team=platform
                RoleEnvironmentRuleArgs.builder()
                    .tags(RoleTagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(RbacEnvironmentScope.EnvironmentOpen)
                    .build())
            .insightsAccountRules(            
                // Read every Insights account
                RoleInsightsAccountRuleArgs.builder()
                    .all(true)
                    .scopes(RbacInsightsAccountScope.InsightsAccountRead)
                    .build(),
                // Scan one specific account
                RoleInsightsAccountRuleArgs.builder()
                    .id(awsAccount.applyValue(_awsAccount -> _awsAccount.insightsAccountId()))
                    .scopes(RbacInsightsAccountScope.InsightsAccountScan)
                    .build(),
                // Update accounts tagged team=platform
                RoleInsightsAccountRuleArgs.builder()
                    .tags(RoleTagConditionArgs.builder()
                        .key("team")
                        .value("platform")
                        .build())
                    .scopes(RbacInsightsAccountScope.InsightsAccountUpdate)
                    .build())
            .additionalEntries(sharedEnvironmentRotate.applyValue(_sharedEnvironmentRotate -> _sharedEnvironmentRotate.permissions()))
            .build());

        final var orgReadOnly = PulumiserviceFunctions.getOrganizationPermissionSet(GetOrganizationPermissionSetArgs.builder()
            .organizationName(organizationName)
            .defaultIdentifier("org-settings-read-only")
            .build());

        final var stackWrite = PulumiserviceFunctions.getOrganizationPermissionSet(GetOrganizationPermissionSetArgs.builder()
            .organizationName(organizationName)
            .defaultIdentifier("stack-write")
            .build());

        final var environmentRead = PulumiserviceFunctions.getOrganizationPermissionSet(GetOrganizationPermissionSetArgs.builder()
            .organizationName(organizationName)
            .defaultIdentifier("environment-read")
            .build());

        final var insightsAccountRead = PulumiserviceFunctions.getOrganizationPermissionSet(GetOrganizationPermissionSetArgs.builder()
            .organizationName(organizationName)
            .defaultIdentifier("insights-account-read")
            .build());

        final var operatorPolicyPermissions = PulumiserviceFunctions.buildRolePermissions(BuildRolePermissionsArgs.builder()
            .organizationPermissionSets(RolePermissionSetRefArgs.builder()
                .permissionSetId(orgReadOnly.applyValue(_orgReadOnly -> _orgReadOnly.permissionSetId()))
                .resourceType(orgReadOnly.applyValue(_orgReadOnly -> _orgReadOnly.resourceType()))
                .build())
            // The Stack Write set on stacks tagged team=platform
            .stackRules(RoleStackRuleArgs.builder()
                .tags(RoleTagConditionArgs.builder()
                    .key("team")
                    .value("platform")
                    .build())
                .permissionSets(RolePermissionSetRefArgs.builder()
                    .permissionSetId(stackWrite.applyValue(_stackWrite -> _stackWrite.permissionSetId()))
                    .resourceType(stackWrite.applyValue(_stackWrite -> _stackWrite.resourceType()))
                    .build())
                .build())
            // The Environment Read set on every environment
            .environmentRules(RoleEnvironmentRuleArgs.builder()
                .all(true)
                .permissionSets(RolePermissionSetRefArgs.builder()
                    .permissionSetId(environmentRead.applyValue(_environmentRead -> _environmentRead.permissionSetId()))
                    .resourceType(environmentRead.applyValue(_environmentRead -> _environmentRead.resourceType()))
                    .build())
                .build())
            // The Insights Account Read set, plus one extra scope, on one account
            .insightsAccountRules(RoleInsightsAccountRuleArgs.builder()
                .id(awsAccount.applyValue(_awsAccount -> _awsAccount.insightsAccountId()))
                .permissionSets(RolePermissionSetRefArgs.builder()
                    .permissionSetId(insightsAccountRead.applyValue(_insightsAccountRead -> _insightsAccountRead.permissionSetId()))
                    .resourceType(insightsAccountRead.applyValue(_insightsAccountRead -> _insightsAccountRead.resourceType()))
                    .build())
                .scopes(RbacInsightsAccountScope.InsightsAccountScan)
                .build())
            .build());

        var operatorPolicy = new Role("operatorPolicy", RoleArgs.builder()
            .orgName(organizationName)
            .name("Platform Operators policy")
            .uxPurpose("policy")
            .details(operatorPolicyPermissions.applyValue(_operatorPolicyPermissions -> _operatorPolicyPermissions.permissions()))
            .build());

        final var operatorPermissions = PulumiserviceFunctions.buildComposePermissions(BuildComposePermissionsArgs.builder()
            .ids(operatorPolicy.roleID())
            .build());

        var platformRole = new OrganizationRole("platformRole", OrganizationRoleArgs.builder()
            .organizationName(organizationName)
            .name("Platform Engineers")
            .description("Read access across the organization, plus deploy access to platform stacks.")
            .permissions(platformPermissions.applyValue(_platformPermissions -> _platformPermissions.permissions()))
            .build());

        var operatorRole = new OrganizationRole("operatorRole", OrganizationRoleArgs.builder()
            .organizationName(organizationName)
            .name("Platform Operators")
            .description("Built-in permission sets on platform entities.")
            .permissions(operatorPermissions.applyValue(_operatorPermissions -> _operatorPermissions.permissions()))
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

  # Any single-purpose helper can be folded into a role through `additionalEntries`.
  sharedEnvironmentRotate:
    fn::invoke:
      function: pulumiservice:buildEnvironmentScopedPermissions
      arguments:
        environmentId: ${sharedEnvironment.environmentId}
        permissions:
          - environment:rotate

  # A role built from scopes: organization-level access plus rules for stacks,
  # environments, and Insights accounts. Each rule list accepts only scopes for
  # its own entity type.
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
            scopes:
              - stack_deployment_settings:read
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
          # Open and update one specific environment
          - id: ${sharedEnvironment.environmentId}
            scopes:
              - environment:open
              - environment:write
          # Open environments tagged team=platform
          - tags:
              - key: team
                value: platform
            scopes:
              - environment:open
        insightsAccountRules:
          # Read every Insights account
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
        additionalEntries:
          - ${sharedEnvironmentRotate.permissions}

  # Permission sets bundle scopes and are shared by reference, so a role that
  # uses them follows changes to the set. Look up the built-in sets to grant,
  # and pass each set's ID with its entity type, which the rule lists check.
  orgReadOnly:
    fn::invoke:
      function: pulumiservice:getOrganizationPermissionSet
      arguments:
        organizationName: ${organizationName}
        defaultIdentifier: org-settings-read-only
  stackWrite:
    fn::invoke:
      function: pulumiservice:getOrganizationPermissionSet
      arguments:
        organizationName: ${organizationName}
        defaultIdentifier: stack-write
  environmentRead:
    fn::invoke:
      function: pulumiservice:getOrganizationPermissionSet
      arguments:
        organizationName: ${organizationName}
        defaultIdentifier: environment-read
  insightsAccountRead:
    fn::invoke:
      function: pulumiservice:getOrganizationPermissionSet
      arguments:
        organizationName: ${organizationName}
        defaultIdentifier: insights-account-read

  # Pulumi Cloud accepts permission sets only in a policy, so these
  # permissions are stored in the policy below.
  operatorPolicyPermissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        organizationPermissionSets:
          - permissionSetId: ${orgReadOnly.permissionSetId}
            resourceType: ${orgReadOnly.resourceType}
        stackRules:
          # The Stack Write set on stacks tagged team=platform
          - tags:
              - key: team
                value: platform
            permissionSets:
              - permissionSetId: ${stackWrite.permissionSetId}
                resourceType: ${stackWrite.resourceType}
        environmentRules:
          # The Environment Read set on every environment
          - all: true
            permissionSets:
              - permissionSetId: ${environmentRead.permissionSetId}
                resourceType: ${environmentRead.resourceType}
        insightsAccountRules:
          # The Insights Account Read set, plus one extra scope, on one account
          - id: ${awsAccount.insightsAccountId}
            permissionSets:
              - permissionSetId: ${insightsAccountRead.permissionSetId}
                resourceType: ${insightsAccountRead.resourceType}
            scopes:
              - insights_account:scan

  # The role grants the policy.
  operatorPermissions:
    fn::invoke:
      function: pulumiservice:buildComposePermissions
      arguments:
        ids:
          - ${operatorPolicy.roleID}

resources:
  platformRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: ${organizationName}
      name: Platform Engineers
      description: Read access across the organization, plus deploy access to platform stacks.
      permissions: ${platformPermissions.permissions}

  operatorPolicy:
    type: pulumiservice:api:Role
    properties:
      orgName: ${organizationName}
      name: Platform Operators policy
      uxPurpose: policy
      details: ${operatorPolicyPermissions.permissions}

  operatorRole:
    type: pulumiservice:OrganizationRole
    properties:
      organizationName: ${organizationName}
      name: Platform Operators
      description: Built-in permission sets on platform entities.
      permissions: ${operatorPermissions.permissions}
```

{{% /example %}}
{{% /examples %}}
