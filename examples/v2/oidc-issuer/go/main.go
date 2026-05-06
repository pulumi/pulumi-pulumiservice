package main

import (
	auth "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/auth"
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
		issuerSuffix := cfg.Get("issuerSuffix")
		if issuerSuffix == "" {
			issuerSuffix = "dev"
		}
		maxExpiration, err := cfg.TryInt("maxExpiration")
		if err != nil {
			maxExpiration = 3600
		}

		pulumiIssuer, err := auth.NewOidcIssuer(ctx, "pulumiIssuer", &auth.OidcIssuerArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("pulumi_issuer_" + issuerSuffix),
			Url:         pulumi.String("https://api.pulumi.com/oidc"),
			Thumbprints: pulumi.StringArray{pulumi.String("57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da")},
		})
		if err != nil {
			return err
		}

		githubIssuer, err := auth.NewOidcIssuer(ctx, "githubIssuer", &auth.OidcIssuerArgs{
			OrgName:       pulumi.String(serviceOrg),
			Name:          pulumi.String("github_issuer_" + issuerSuffix),
			Url:           pulumi.String("https://token.actions.githubusercontent.com"),
			Thumbprints:   pulumi.StringArray{pulumi.String("caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7")},
			MaxExpiration: pulumi.Int(maxExpiration),
		})
		if err != nil {
			return err
		}

		ctx.Export("pulumiIssuerName", pulumiIssuer.Name)
		ctx.Export("githubIssuerName", githubIssuer.Name)
		return nil
	})
}
