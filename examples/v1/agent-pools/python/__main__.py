import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
pool_suffix = config.get("poolSuffix") or "dev"
pool_description = config.get("poolDescription") or "v1 example agent pool"

pool = ps_v1.agents.Pool(
    "pool",
    org_name=organization_name,
    name=f"v1-agent-pool-{pool_suffix}",
    description=pool_description,
)

pulumi.export("poolName", pool.name)
