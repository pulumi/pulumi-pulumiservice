import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
service_suffix = config.get("serviceSuffix") or "dev"

ps_api.services.Service(
    "catalogService",
    org_name=organization_name,
    name=f"api-service-{service_suffix}",
    description="An example api service catalog entry.",
    owner_type="team",
    owner_name="platform",
    items=[
        {"kind": "stack", "ref": "service-provider-test-org/example-app/dev"},
    ],
    properties=[
        {"key": "tier", "value": "gold"},
        {"key": "oncall", "value": "platform-ops"},
    ],
)
