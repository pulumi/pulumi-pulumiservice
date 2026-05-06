package main

import (
	v2 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		serviceOrg := cfg.Get("serviceOrg")
		if serviceOrg == "" {
			serviceOrg = "service-provider-test-org"
		}

		approvers, err := v2.NewPolicyGroup(ctx, "approvers", &v2.PolicyGroupArgs{
			OrgName:    pulumi.String(serviceOrg),
			Name:       pulumi.String("v2-approvers"),
			EntityType: pulumi.String("stacks"),
		})
		if err != nil {
			return err
		}

		ctx.Export("policyGroupName", approvers.Name)
		return nil
	})
}
