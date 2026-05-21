import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"

default = ps_api.DefaultOrganization(
    "default",
    org_name=organization_name,
)

pulumi.export("defaultOrg", organization_name)
pulumi.export("defaultOrgGitHubLogin", default.git_hub_login)
