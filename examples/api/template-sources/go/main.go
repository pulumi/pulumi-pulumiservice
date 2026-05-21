package main

import (
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
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
		templateSuffix := cfg.Get("templateSuffix")
		if templateSuffix == "" {
			templateSuffix = "dev"
		}
		sourceURL := cfg.Get("sourceUrl")
		if sourceURL == "" {
			sourceURL = "https://github.com/pulumi/examples"
		}

		source, err := api.NewOrgTemplateCollection(ctx, "source", &api.OrgTemplateCollectionArgs{
			OrgName:   pulumi.String(organizationName),
			Name:      pulumi.String("api-templates-" + templateSuffix),
			SourceURL: pulumi.String(sourceURL),
		})
		if err != nil {
			return err
		}

		ctx.Export("collectionName", source.Name)
		return nil
	})
}
