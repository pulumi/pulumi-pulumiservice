package main

import (
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
	services "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/services"
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
		serviceSuffix := cfg.Get("serviceSuffix")
		if serviceSuffix == "" {
			serviceSuffix = "dev"
		}

		_, err := services.NewService(ctx, "catalogService", &services.ServiceArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-service-" + serviceSuffix),
			Description: pulumi.String("An example api service catalog entry."),
			OwnerType:   pulumi.String("team"),
			OwnerName:   pulumi.String("platform"),
			Items: api.AddServiceItemArray{
				api.AddServiceItemArgs{
					Type: pulumi.String("stack"),
					Name: pulumi.String("service-provider-test-org/example-app/dev"),
				},
			},
			Properties: api.ServicePropertyArray{
				api.ServicePropertyArgs{
					Key:   pulumi.String("tier"),
					Value: pulumi.String("gold"),
					Type:  pulumi.String("string"),
					Order: pulumi.Int(1),
				},
				api.ServicePropertyArgs{
					Key:   pulumi.String("oncall"),
					Value: pulumi.String("platform-ops"),
					Type:  pulumi.String("string"),
					Order: pulumi.Int(2),
				},
			},
		})
		return err
	})
}
