import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"

approvers = ps_v2.PolicyGroup(
    "approvers",
    org_name=organization_name,
    name="v2-approvers",
    entity_type="stacks",
)

pulumi.export("policyGroupName", approvers.name)
