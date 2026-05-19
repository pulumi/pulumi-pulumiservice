package main

import (
	auth "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/auth"
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
		policyId := cfg.Get("policyId")
		if policyId == "" {
			policyId = "org"
		}

		_, err := auth.NewPolicy(ctx, "policy", &auth.PolicyArgs{
			OrgName:  pulumi.String(organizationName),
			PolicyId: pulumi.String(policyId),
			Policies: pulumi.Array{
				pulumi.Map{
					"decision":   pulumi.String("allow"),
					"permission": pulumi.String("read"),
					"tokenType":  pulumi.String("organization"),
				},
				pulumi.Map{
					"decision":   pulumi.String("deny"),
					"permission": pulumi.String("admin"),
					"tokenType":  pulumi.String("organization"),
				},
			},
		})
		return err
	})
}
