package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_agents.Pool;
import com.pulumi.pulumiservice.api_agents.PoolArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var poolSuffix = config.get("poolSuffix").orElse("dev");
            var poolDescription = config.get("poolDescription").orElse("api example agent pool");

            var pool = new Pool("pool",
                PoolArgs.builder()
                    .orgName(organizationName)
                    .name("api-agent-pool-" + poolSuffix)
                    .description(poolDescription)
                    .build());

            ctx.export("poolName", pool.name());
        });
    }
}
