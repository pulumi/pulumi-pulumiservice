import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"

pack = ps_v2.PolicyPack(
    "pack",
    org_name=organization_name,
    name="v2-example-policy-pack",
    display_name="v2 example policy pack",
    description="Demo policy pack created via v2 metadata-driven provider.",
    policies=[{
        "name": "no-public-buckets",
        "description": "Reject S3 buckets with public ACLs",
        "enforcementLevel": "advisory",
    }],
)

pulumi.export("policyPackName", pack.name)
pulumi.export("policyPackVersion", pack.version)
