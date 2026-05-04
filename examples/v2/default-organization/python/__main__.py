import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"

default = ps_v2.DefaultOrganization(
    "default",
    org_name=service_org,
)

pulumi.export("defaultOrg", default.org_name)
