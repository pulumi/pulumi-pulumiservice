import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const poolSuffix = config.get("poolSuffix") ?? "dev";
const poolDescription = config.get("poolDescription") ?? "api example agent pool";

const pool = new ps.api.agents.Pool("pool", {
    orgName: organizationName,
    name: `api-agent-pool-${poolSuffix}`,
    description: poolDescription,
});

export const poolName = pool.name;
