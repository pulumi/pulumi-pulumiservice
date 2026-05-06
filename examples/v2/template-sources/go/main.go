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
		templateSuffix := cfg.Get("templateSuffix")
		if templateSuffix == "" {
			templateSuffix = "dev"
		}
		sourceURL := cfg.Get("sourceUrl")
		if sourceURL == "" {
			sourceURL = "https://github.com/pulumi/examples"
		}

		source, err := v2.NewOrgTemplateCollection(ctx, "source", &v2.OrgTemplateCollectionArgs{
			OrgName:   pulumi.String(serviceOrg),
			Name:      pulumi.String("v2-templates-" + templateSuffix),
			SourceURL: pulumi.String(sourceURL),
		})
		if err != nil {
			return err
		}

		ctx.Export("collectionName", source.Name)
		return nil
	})
}
