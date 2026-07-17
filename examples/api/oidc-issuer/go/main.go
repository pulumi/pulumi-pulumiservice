package main

import (
	auth "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/auth"
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
		issuerSuffix := cfg.Get("issuerSuffix")
		if issuerSuffix == "" {
			issuerSuffix = "dev"
		}
		maxExpiration, err := cfg.TryInt("maxExpiration")
		if err != nil {
			maxExpiration = 3600
		}
		// Thumbprints must match the certificate the issuer currently serves, so
		// they have no static default. Compute one with:
		//   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
		pulumiThumbprint := cfg.Require("pulumiThumbprint")
		githubThumbprint := cfg.Require("githubThumbprint")

		pulumiIssuer, err := auth.NewOidcIssuer(ctx, "pulumiIssuer", &auth.OidcIssuerArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("pulumi_issuer_" + issuerSuffix),
			Url:         pulumi.String("https://api.pulumi.com/oidc"),
			Thumbprints: pulumi.StringArray{pulumi.String(pulumiThumbprint)},
		})
		if err != nil {
			return err
		}

		githubIssuer, err := auth.NewOidcIssuer(ctx, "githubIssuer", &auth.OidcIssuerArgs{
			OrgName:       pulumi.String(organizationName),
			Name:          pulumi.String("github_issuer_" + issuerSuffix),
			Url:           pulumi.String("https://token.actions.githubusercontent.com"),
			Thumbprints:   pulumi.StringArray{pulumi.String(githubThumbprint)},
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
