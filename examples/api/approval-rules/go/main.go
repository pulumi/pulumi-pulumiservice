package main

import (
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
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
		groupName := cfg.Get("groupName")
		if groupName == "" {
			groupName = "api-approvers"
		}

		approvers, err := api.NewPolicyGroup(ctx, "approvers", &api.PolicyGroupArgs{
			OrgName:    pulumi.String(organizationName),
			Name:       pulumi.String(groupName),
			EntityType: pulumi.String("stacks"),
		})
		if err != nil {
			return err
		}

		ctx.Export("policyGroupName", approvers.Name)
		return nil
	})
}
