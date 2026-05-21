import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
group_name = config.get("groupName") or "example-policy-group"

group = ps_api.PolicyGroup(
    "group",
    org_name=organization_name,
    name=group_name,
    entity_type="stacks",
)

pulumi.export("policyGroupName", group.name)
