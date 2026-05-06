import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
group_name = config.get("groupName") or "example-policy-group"

group = ps_v2.PolicyGroup(
    "group",
    org_name=service_org,
    name=group_name,
    entity_type="stacks",
)

pulumi.export("policyGroupName", group.name)
