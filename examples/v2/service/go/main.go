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
		serviceSuffix := cfg.Get("serviceSuffix")
		if serviceSuffix == "" {
			serviceSuffix = "dev"
		}

		_, err := v2.NewService(ctx, "catalogService", &v2.ServiceArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-service-" + serviceSuffix),
			Description: pulumi.String("An example v2 service catalog entry."),
			OwnerType:   pulumi.String("team"),
			OwnerName:   pulumi.String("platform"),
			Items: pulumi.Array{
				pulumi.Map{
					"kind": pulumi.String("stack"),
					"ref":  pulumi.String("service-provider-test-org/example-app/dev"),
				},
			},
			Properties: pulumi.Array{
				pulumi.Map{"key": pulumi.String("tier"), "value": pulumi.String("gold")},
				pulumi.Map{"key": pulumi.String("oncall"), "value": pulumi.String("platform-ops")},
			},
		})
		return err
	})
}
