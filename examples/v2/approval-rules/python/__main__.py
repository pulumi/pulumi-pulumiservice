import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"

approvers = ps_v2.PolicyGroup(
    "approvers",
    org_name=service_org,
    name="v2-approvers",
    entity_type="stacks",
)

pulumi.export("policyGroupName", approvers.name)
