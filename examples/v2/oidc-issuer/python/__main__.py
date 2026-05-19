import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
issuer_suffix = config.get("issuerSuffix") or "dev"
max_expiration = config.get_int("maxExpiration") or 3600

pulumi_issuer = ps_v2.auth.OidcIssuer(
    "pulumiIssuer",
    org_name=service_org,
    name=f"pulumi_issuer_{issuer_suffix}",
    url="https://api.pulumi.com/oidc",
    thumbprints=["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
)

github_issuer = ps_v2.auth.OidcIssuer(
    "githubIssuer",
    org_name=service_org,
    name=f"github_issuer_{issuer_suffix}",
    url="https://token.actions.githubusercontent.com",
    thumbprints=["b41ae0832808ebc94951437bf7e92b93ccb6479364daf894d46d6001bee7a486"],
    max_expiration=max_expiration,
)

pulumi.export("pulumiIssuerName", pulumi_issuer.name)
pulumi.export("githubIssuerName", github_issuer.name)
