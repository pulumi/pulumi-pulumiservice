package main

import (
	v1 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1"
	teams "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/teams"
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
			roleDescription = "Read-only access to stacks, created by the v1 rbac example."
		}

		readOnlyRole, err := v1.NewRole(ctx, "readOnlyRole", &v1.RoleArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("v1-rbac-read-only-" + nameSuffix),
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
			Name:        pulumi.String("v1-rbac-team-" + nameSuffix),
			DisplayName: pulumi.String("v1 RBAC Team " + nameSuffix),
			Description: pulumi.String("Team scaffold used by the v1 rbac example."),
		})
		if err != nil {
			return err
		}

		ctx.Export("roleName", readOnlyRole.Name)
		ctx.Export("teamName", rbacTeam.Name)
		return nil
	})
}
