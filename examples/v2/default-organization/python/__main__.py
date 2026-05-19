import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"

default = ps_v2.DefaultOrganization(
    "default",
    org_name=organization_name,
)

pulumi.export("defaultOrg", organization_name)
pulumi.export("defaultOrgGitHubLogin", default.git_hub_login)
