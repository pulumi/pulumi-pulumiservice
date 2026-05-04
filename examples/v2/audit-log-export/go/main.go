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
		bucketName := cfg.Get("bucketName")
		if bucketName == "" {
			bucketName = "pulumi-audit-log-archive"
		}
		region := cfg.Get("region")
		if region == "" {
			region = "us-west-2"
		}

		exportConfig, err := v2.NewAuditLogExportConfiguration(ctx, "exportConfig", &v2.AuditLogExportConfigurationArgs{
			OrgName:    pulumi.String(serviceOrg),
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
