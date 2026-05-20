import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"

approvers = ps_v1.PolicyGroup(
    "approvers",
    org_name=organization_name,
    name="v1-approvers",
    entity_type="stacks",
)

pulumi.export("policyGroupName", approvers.name)
