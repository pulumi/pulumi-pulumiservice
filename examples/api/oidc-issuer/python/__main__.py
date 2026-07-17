import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
issuer_suffix = config.get("issuerSuffix") or "dev"
max_expiration = config.get_int("maxExpiration") or 3600
# Thumbprints must match the certificate the issuer currently serves, so they
# have no static default. Compute one with:
#   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
pulumi_thumbprint = config.require("pulumiThumbprint")
github_thumbprint = config.require("githubThumbprint")

pulumi_issuer = ps_api.auth.OidcIssuer(
    "pulumiIssuer",
    org_name=organization_name,
    name=f"pulumi_issuer_{issuer_suffix}",
    url="https://api.pulumi.com/oidc",
    thumbprints=[pulumi_thumbprint],
)

github_issuer = ps_api.auth.OidcIssuer(
    "githubIssuer",
    org_name=organization_name,
    name=f"github_issuer_{issuer_suffix}",
    url="https://token.actions.githubusercontent.com",
    thumbprints=[github_thumbprint],
    max_expiration=max_expiration,
)

pulumi.export("pulumiIssuerName", pulumi_issuer.name)
pulumi.export("githubIssuerName", github_issuer.name)
