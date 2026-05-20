package main

import (
	v1 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1"
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

		approvers, err := v1.NewPolicyGroup(ctx, "approvers", &v1.PolicyGroupArgs{
			OrgName:    pulumi.String(organizationName),
			Name:       pulumi.String("v1-approvers"),
			EntityType: pulumi.String("stacks"),
		})
		if err != nil {
			return err
		}

		ctx.Export("policyGroupName", approvers.Name)
		return nil
	})
}
