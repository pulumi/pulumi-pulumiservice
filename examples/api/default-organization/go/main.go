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

		def, err := api.NewDefaultOrganization(ctx, "default", &api.DefaultOrganizationArgs{
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
