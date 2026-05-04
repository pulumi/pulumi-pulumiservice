import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
service_suffix = config.get("serviceSuffix") or "dev"

ps_v2.Service(
    "catalogService",
    org_name=service_org,
    name=f"v2-service-{service_suffix}",
    description="An example v2 service catalog entry.",
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
