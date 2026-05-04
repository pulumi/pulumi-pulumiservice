import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
policy_id = config.get("policyId") or "org"

ps_v2.AuthPolicy(
    "policy",
    org_name=service_org,
    policy_id=policy_id,
    policies=[
        {"decision": "allow", "permission": "read", "tokenType": "organization"},
        {"decision": "deny", "permission": "admin", "tokenType": "organization"},
    ],
)
