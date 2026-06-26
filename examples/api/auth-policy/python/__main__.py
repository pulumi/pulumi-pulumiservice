import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
policy_id = config.get("policyId") or "org"

ps_api.auth.Policy(
    "policy",
    org_name=organization_name,
    policy_id=policy_id,
    policies=[
        {"decision": "allow", "authorizedPermissions": ["read"], "tokenType": "organization", "rules": {}},
        {"decision": "deny", "authorizedPermissions": ["admin"], "tokenType": "organization", "rules": {}},
    ],
)
