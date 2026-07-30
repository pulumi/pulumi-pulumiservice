package main

import (
	auth "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/auth"
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
			Policies: auth.AuthPolicyDefinitionArray{
				auth.AuthPolicyDefinitionArgs{
					Decision:              pulumi.String("allow"),
					TokenType:             pulumi.String("organization"),
					AuthorizedPermissions: pulumi.ToStringArray([]string{"standard"}),
					Rules:                 pulumi.MapMap{},
				},
				auth.AuthPolicyDefinitionArgs{
					Decision:              pulumi.String("deny"),
					TokenType:             pulumi.String("organization"),
					AuthorizedPermissions: pulumi.ToStringArray([]string{"admin"}),
					Rules:                 pulumi.MapMap{},
				},
			},
		})
		return err
	})
}
