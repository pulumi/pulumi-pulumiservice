package main

import (
	"os"

	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		org := os.Getenv("POC_ORG")
		if org == "" {
			org = "poc-org"
		}

		// The generated PermissionDescriptor*Args structs render __type as an
		// UNEXPORTED field (leading underscores survive into the Go identifier),
		// so the typed path cannot set the tag. Raw maps are the only option.
		role, err := api.NewRole(ctx, "poc-union-role", &api.RoleArgs{
			OrgName:      pulumi.String(org),
			Name:         pulumi.String("poc-union-role"),
			UxPurpose:    pulumi.String("set"),
			ResourceType: pulumi.String("global"),
			Description:  pulumi.String("POC: discriminated union permission tree"),
			Details: pulumi.Map{
				"__type": pulumi.String("PermissionDescriptorGroup"),
				"entries": pulumi.Array{
					pulumi.Map{
						"__type":      pulumi.String("PermissionDescriptorAllow"),
						"permissions": pulumi.ToStringArray([]string{"organization:read_usage"}),
					},
				},
			},
		})
		if err != nil {
			return err
		}
		ctx.Export("roleId", role.ID())
		ctx.Export("detailsOut", role.Details)
		return nil
	})
}
