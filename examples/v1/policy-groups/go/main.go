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
		groupName := cfg.Get("groupName")
		if groupName == "" {
			groupName = "example-policy-group"
		}

		group, err := v1.NewPolicyGroup(ctx, "group", &v1.PolicyGroupArgs{
			OrgName:    pulumi.String(organizationName),
			Name:       pulumi.String(groupName),
			EntityType: pulumi.String("stacks"),
		})
		if err != nil {
			return err
		}

		ctx.Export("policyGroupName", group.Name)
		return nil
	})
}
