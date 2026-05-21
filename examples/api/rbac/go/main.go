package main

import (
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
			roleDescription = "Read-only access to stacks, created by the api rbac example."
		}

		readOnlyRole, err := api.NewRole(ctx, "readOnlyRole", &api.RoleArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-rbac-read-only-" + nameSuffix),
			Description: pulumi.String(roleDescription),
			UxPurpose:   pulumi.String("role"),
			Details: pulumi.Map{
				"__type":      pulumi.String("PermissionDescriptorAllow"),
				"permissions": pulumi.StringArray{pulumi.String("stack:read")},
			},
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

		ctx.Export("roleName", readOnlyRole.Name)
		ctx.Export("teamName", rbacTeam.Name)
		return nil
	})
}
