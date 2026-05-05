import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
pool_suffix = config.get("poolSuffix") or "dev"
pool_description = config.get("poolDescription") or "v2 example agent pool"

pool = ps_v2.agents.Pool(
    "pool",
    org_name=service_org,
    name=f"v2-agent-pool-{pool_suffix}",
    description=pool_description,
)

pulumi.export("poolName", pool.name)
