package main

import (
	auth "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/auth"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const idpDescriptor = `<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
                     entityID="https://example-idp.invalid/metadata">
  <md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"/>
</md:EntityDescriptor>`

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		organizationName := cfg.Get("organizationName")
		if organizationName == "" {
			organizationName = "service-provider-test-org"
		}

		_, err := auth.NewSAML(ctx, "saml", &auth.SAMLArgs{
			OrgName:             pulumi.String(organizationName),
			NewIdpSsoDescriptor: pulumi.String(idpDescriptor),
		})
		return err
	})
}
