import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
group_name = config.get("groupName") or "api-approvers"

approvers = ps_api.PolicyGroup(
    "approvers",
    org_name=organization_name,
    name=group_name,
    entity_type="stacks",
)

pulumi.export("policyGroupName", approvers.name)
