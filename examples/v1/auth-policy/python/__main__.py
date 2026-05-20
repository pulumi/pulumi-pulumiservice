import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
policy_id = config.get("policyId") or "org"

ps_v1.auth.Policy(
    "policy",
    org_name=organization_name,
    policy_id=policy_id,
    policies=[
        {"decision": "allow", "permission": "read", "tokenType": "organization"},
        {"decision": "deny", "permission": "admin", "tokenType": "organization"},
    ],
)
