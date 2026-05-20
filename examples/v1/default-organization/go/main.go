package main

import (
	v1 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1"
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

		def, err := v1.NewDefaultOrganization(ctx, "default", &v1.DefaultOrganizationArgs{
			OrgName: pulumi.String(organizationName),
		})
		if err != nil {
			return err
		}

		ctx.Export("defaultOrg", pulumi.String(organizationName))
		ctx.Export("defaultOrgGitHubLogin", def.GitHubLogin)
		return nil
	})
}
