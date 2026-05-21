import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
pool_suffix = config.get("poolSuffix") or "dev"
pool_description = config.get("poolDescription") or "api example agent pool"

pool = ps_api.agents.Pool(
    "pool",
    org_name=organization_name,
    name=f"api-agent-pool-{pool_suffix}",
    description=pool_description,
)

pulumi.export("poolName", pool.name)
