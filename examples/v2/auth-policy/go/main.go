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
		policyId := cfg.Get("policyId")
		if policyId == "" {
			policyId = "org"
		}

		_, err := v2.NewAuthPolicy(ctx, "policy", &v2.AuthPolicyArgs{
			OrgName:  pulumi.String(serviceOrg),
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
