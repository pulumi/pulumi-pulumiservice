package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Raw maps: the typed variant structs are pending an upstream fix for
// double-underscore field names in Go codegen.
func envMatch(identity string) pulumi.Map {
	return pulumi.Map{
		"type__": pulumi.String("PermissionExpressionEqual"),
		"left":   pulumi.Map{"type__": pulumi.String("PermissionExpressionEnvironment")},
		"right": pulumi.Map{
			"type__":   pulumi.String("PermissionLiteralExpressionEnvironment"),
			"identity": pulumi.String(identity),
		},
	}
}

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
			roleDescription = "Environment-scoped read access, created by the api rbac-scoped example."
		}
		allowedEnvironmentID := cfg.Get("allowedEnvironmentId")
		if allowedEnvironmentID == "" {
			allowedEnvironmentID = "c5549aa1-87db-4d67-a195-455b56772900"
		}
		deniedEnvironmentID := cfg.Get("deniedEnvironmentId")
		if deniedEnvironmentID == "" {
			deniedEnvironmentID = "3cb9b7ad-0848-4e0d-aeff-8e9f093fd2d9"
		}

		scopedRole, err := api.NewRole(ctx, "scopedRole", &api.RoleArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-rbac-scoped-" + nameSuffix),
			Description: pulumi.String(roleDescription),
			UxPurpose:   pulumi.String("role"),
			Details: pulumi.Map{
				"type__": pulumi.String("PermissionDescriptorGroup"),
				"entries": pulumi.Array{
					pulumi.Map{
						"type__": pulumi.String("PermissionDescriptorCondition"),
						"condition": pulumi.Map{
							"type__": pulumi.String("PermissionExpressionAnd"),
							"left":   envMatch(allowedEnvironmentID),
							"right": pulumi.Map{
								"type__": pulumi.String("PermissionExpressionNot"),
								"node":   envMatch(deniedEnvironmentID),
							},
						},
						"subNode": pulumi.Map{
							"type__": pulumi.String("PermissionDescriptorAllow"),
							"permissions": pulumi.ToStringArray([]string{
								"environment:read", "environment:open",
							}),
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}
		ctx.Export("roleName", scopedRole.Name)
		return nil
	})
}
