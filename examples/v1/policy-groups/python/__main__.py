import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
group_name = config.get("groupName") or "example-policy-group"

group = ps_v1.PolicyGroup(
    "group",
    org_name=organization_name,
    name=group_name,
    entity_type="stacks",
)

pulumi.export("policyGroupName", group.name)
