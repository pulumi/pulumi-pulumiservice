package main

import (
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
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
			Policies: api.AuthPolicyDefinitionArray{
				api.AuthPolicyDefinitionArgs{
					Decision:              pulumi.String("allow"),
					AuthorizedPermissions: pulumi.ToStringArray([]string{"read"}),
					TokenType:             pulumi.String("organization"),
					Rules:                 pulumi.Map{},
				},
				api.AuthPolicyDefinitionArgs{
					Decision:              pulumi.String("deny"),
					AuthorizedPermissions: pulumi.ToStringArray([]string{"admin"}),
					TokenType:             pulumi.String("organization"),
					Rules:                 pulumi.Map{},
				},
			},
		})
		return err
	})
}
