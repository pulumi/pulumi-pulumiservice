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
		bucketName := cfg.Get("bucketName")
		if bucketName == "" {
			bucketName = "pulumi-audit-log-archive"
		}
		region := cfg.Get("region")
		if region == "" {
			region = "us-west-2"
		}

		exportConfig, err := api.NewAuditLogExportConfiguration(ctx, "exportConfig", &api.AuditLogExportConfigurationArgs{
			OrgName:    pulumi.String(organizationName),
			NewEnabled: pulumi.Bool(true),
			NewS3Configuration: pulumi.Map{
				"bucketName": pulumi.String(bucketName),
				"region":     pulumi.String(region),
				"roleArn":    pulumi.String("arn:aws:iam::123456789012:role/PulumiAuditLogExportRole"),
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("exportEnabled", exportConfig.Enabled)
		return nil
	})
}
