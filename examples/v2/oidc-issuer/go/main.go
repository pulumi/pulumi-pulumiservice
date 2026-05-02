package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		serviceOrg := "service-provider-test-org2"
		if param := cfg.Get("serviceOrg"); param != "" {
			serviceOrg = param
		}
		pulumiIssuer, err := v2.NewOidcIssuer(ctx, "pulumiIssuer", &v2.OidcIssuerArgs{
			OrgName: serviceOrg,
			Name:    "pulumi_issuer",
			Url:     "https://api.pulumi.com/oidc",
			Thumbprints: []string{
				"57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da",
			},
		})
		if err != nil {
			return err
		}
		githubIssuer, err := v2.NewOidcIssuer(ctx, "githubIssuer", &v2.OidcIssuerArgs{
			OrgName: serviceOrg,
			Name:    "github_issuer",
			Url:     "https://token.actions.githubusercontent.com",
			Thumbprints: []string{
				"caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7",
			},
			MaxExpiration: 3600,
		})
		if err != nil {
			return err
		}
		ctx.Export("pulumiIssuerUrl", pulumi.Any(pulumiIssuer.Url))
		ctx.Export("githubIssuerUrl", pulumi.Any(githubIssuer.Url))
		return nil
	})
}
