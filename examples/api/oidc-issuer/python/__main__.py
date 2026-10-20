import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
issuer_suffix = config.get("issuerSuffix") or "dev"
max_expiration = config.get_int("maxExpiration") or 3600
# Tests override the issuer URLs so that concurrent runs don't conflict;
# see oidcIssuerTestConfig in examples/examples_utils_test.go.
pulumi_issuer_url = config.get("pulumiIssuerUrl") or "https://api.pulumi.com/oidc"
github_issuer_url = config.get("githubIssuerUrl") or "https://token.actions.githubusercontent.com"
# Thumbprints must match the certificate the issuer currently serves, so they
# have no static default. Compute one with:
#   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
pulumi_thumbprint = config.require("pulumiThumbprint")
github_thumbprint = config.require("githubThumbprint")

pulumi_issuer = ps_api.auth.OidcIssuer(
    "pulumiIssuer",
    org_name=organization_name,
    name=f"pulumi_issuer_{issuer_suffix}",
    url=pulumi_issuer_url,
    thumbprints=[pulumi_thumbprint],
)

github_issuer = ps_api.auth.OidcIssuer(
    "githubIssuer",
    org_name=organization_name,
    name=f"github_issuer_{issuer_suffix}",
    url=github_issuer_url,
    thumbprints=[github_thumbprint],
    max_expiration=max_expiration,
)

pulumi.export("pulumiIssuerName", pulumi_issuer.name)
pulumi.export("githubIssuerName", github_issuer.name)
