import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
issuer_suffix = config.get("issuerSuffix") or "dev"
max_expiration = config.get_int("maxExpiration") or 3600

pulumi_issuer = ps_api.auth.OidcIssuer(
    "pulumiIssuer",
    org_name=organization_name,
    name=f"pulumi_issuer_{issuer_suffix}",
    url="https://api.pulumi.com/oidc",
    thumbprints=["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
)

github_issuer = ps_api.auth.OidcIssuer(
    "githubIssuer",
    org_name=organization_name,
    name=f"github_issuer_{issuer_suffix}",
    url="https://token.actions.githubusercontent.com",
    thumbprints=["39517789ff0132a9212bafea4dc37401eae58b1bfac9756109d14301c90a6ab5"],
    max_expiration=max_expiration,
)

pulumi.export("pulumiIssuerName", pulumi_issuer.name)
pulumi.export("githubIssuerName", github_issuer.name)
