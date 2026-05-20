package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v1_agents.Pool;
import com.pulumi.pulumiservice.v1_agents.PoolArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var poolSuffix = config.get("poolSuffix").orElse("dev");
            var poolDescription = config.get("poolDescription").orElse("v1 example agent pool");

            var pool = new Pool("pool",
                PoolArgs.builder()
                    .orgName(organizationName)
                    .name("v1-agent-pool-" + poolSuffix)
                    .description(poolDescription)
                    .build());

            ctx.export("poolName", pool.name());
        });
    }
}
