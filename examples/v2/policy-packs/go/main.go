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

		pack, err := v2.NewPolicyPack(ctx, "pack", &v2.PolicyPackArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-example-policy-pack"),
			DisplayName: pulumi.String("v2 example policy pack"),
			Description: pulumi.String("Demo policy pack created via v2 metadata-driven provider."),
			Policies: pulumi.Array{
				pulumi.Map{
					"name":             pulumi.String("no-public-buckets"),
					"description":      pulumi.String("Reject S3 buckets with public ACLs"),
					"enforcementLevel": pulumi.String("advisory"),
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("policyPackName", pack.Name)
		return nil
	})
}
