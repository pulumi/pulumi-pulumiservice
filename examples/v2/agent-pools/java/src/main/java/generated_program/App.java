package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.AgentPool;
import com.pulumi.pulumiservice.v2.AgentPoolArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var poolSuffix = config.get("poolSuffix").orElse("dev");
            var poolDescription = config.get("poolDescription").orElse("v2 example agent pool");

            var pool = new AgentPool("pool",
                AgentPoolArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-agent-pool-" + poolSuffix)
                    .description(poolDescription)
                    .build());

            ctx.export("poolName", pool.name());
        });
    }
}
