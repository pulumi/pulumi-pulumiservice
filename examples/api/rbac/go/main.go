package main

import (
	ps "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
	teams "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/teams"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		organizationName := cfg.Get("organizationName")
		if organizationName == "" {
			organizationName = "service-provider-test-org"
		}
		nameSuffix := cfg.Get("nameSuffix")
		if nameSuffix == "" {
			nameSuffix = "manual"
		}
		roleDescription := cfg.Get("roleDescription")
		if roleDescription == "" {
			roleDescription = "Stack read everywhere, plus write on stacks tagged team=platform."
		}

		// Pulumi Cloud models a role in three layers: a set grants scopes, a policy
		// composes sets (optionally gated on a condition), and a role composes policies.
		stackReadSet, err := api.NewRole(ctx, "stackReadSet", &api.RoleArgs{
			OrgName:      pulumi.String(organizationName),
			Name:         pulumi.String("api-rbac-stack-read-" + nameSuffix),
			UxPurpose:    pulumi.String("set"),
			ResourceType: pulumi.String("stack"),
			Details: ps.BuildAllowPermissionsOutput(ctx, ps.BuildAllowPermissionsOutputArgs{
				Permissions: pulumi.StringArray{pulumi.String("stack:read")},
			}).Permissions(),
		})
		if err != nil {
			return err
		}

		stackWriteSet, err := api.NewRole(ctx, "stackWriteSet", &api.RoleArgs{
			OrgName:      pulumi.String(organizationName),
			Name:         pulumi.String("api-rbac-stack-write-" + nameSuffix),
			UxPurpose:    pulumi.String("set"),
			ResourceType: pulumi.String("stack"),
			Details: ps.BuildAllowPermissionsOutput(ctx, ps.BuildAllowPermissionsOutputArgs{
				Permissions: pulumi.StringArray{pulumi.String("stack:read"), pulumi.String("stack:write")},
			}).Permissions(),
		})
		if err != nil {
			return err
		}

		platformPolicy, err := api.NewRole(ctx, "platformPolicy", &api.RoleArgs{
			OrgName:   pulumi.String(organizationName),
			Name:      pulumi.String("api-rbac-policy-" + nameSuffix),
			UxPurpose: pulumi.String("policy"),
			Details: ps.BuildGroupPermissionsOutput(ctx, ps.BuildGroupPermissionsOutputArgs{
				Entries: pulumi.MapArray{
					ps.BuildComposePermissionsOutput(ctx, ps.BuildComposePermissionsOutputArgs{
						PermissionDescriptorIds: pulumi.StringArray{stackReadSet.RoleID},
					}).Permissions(),
					ps.BuildTagConditionalPermissionsOutput(ctx, ps.BuildTagConditionalPermissionsOutputArgs{
						EntityType: ps.RbacResourceTypeStack,
						TagKey:     pulumi.String("team"),
						TagValue:   pulumi.String("platform"),
						SetIds:     pulumi.StringArray{stackWriteSet.RoleID},
					}).Permissions(),
				},
			}).Permissions(),
		})
		if err != nil {
			return err
		}

		platformRole, err := api.NewRole(ctx, "platformRole", &api.RoleArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-rbac-role-" + nameSuffix),
			Description: pulumi.String(roleDescription),
			UxPurpose:   pulumi.String("role"),
			Details: ps.BuildComposePermissionsOutput(ctx, ps.BuildComposePermissionsOutputArgs{
				PermissionDescriptorIds: pulumi.StringArray{platformPolicy.RoleID},
			}).Permissions(),
		})
		if err != nil {
			return err
		}

		rbacTeam, err := teams.NewTeam(ctx, "rbacTeam", &teams.TeamArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-rbac-team-" + nameSuffix),
			DisplayName: pulumi.String("api RBAC Team " + nameSuffix),
			Description: pulumi.String("Team scaffold used by the api rbac example."),
		})
		if err != nil {
			return err
		}

		ctx.Export("roleName", platformRole.Name)
		ctx.Export("teamName", rbacTeam.Name)
		return nil
	})
}
