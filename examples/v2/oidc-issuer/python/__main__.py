import pulumi
import pulumi_pulumiservice as pulumiservice

config = pulumi.Config()
service_org = config.get("serviceOrg")
if service_org is None:
    service_org = "service-provider-test-org2"
pulumi_issuer = pulumiservice.v2.OidcIssuer("pulumiIssuer",
    org_name=service_org,
    name=pulumi_issuer,
    url=https://api.pulumi.com/oidc,
    thumbprints=[57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da])
github_issuer = pulumiservice.v2.OidcIssuer("githubIssuer",
    org_name=service_org,
    name=github_issuer,
    url=https://token.actions.githubusercontent.com,
    thumbprints=[caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7],
    max_expiration=3600)
pulumi.export("pulumiIssuerUrl", pulumi_issuer["url"])
pulumi.export("githubIssuerUrl", github_issuer["url"])
