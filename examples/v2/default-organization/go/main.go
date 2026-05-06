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

		def, err := v2.NewDefaultOrganization(ctx, "default", &v2.DefaultOrganizationArgs{
			OrgName: pulumi.String(serviceOrg),
		})
		if err != nil {
			return err
		}

		ctx.Export("defaultOrg", def.OrgName)
		return nil
	})
}
