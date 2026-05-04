package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.Service;
import com.pulumi.pulumiservice.v2.ServiceArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var serviceSuffix = config.get("serviceSuffix").orElse("dev");

            new Service("catalogService",
                ServiceArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-service-" + serviceSuffix)
                    .description("An example v2 service catalog entry.")
                    .ownerType("team")
                    .ownerName("platform")
                    .items(List.of(
                        Map.of("kind", "stack", "ref", "service-provider-test-org/example-app/dev")))
                    .properties(List.of(
                        Map.of("key", "tier", "value", "gold"),
                        Map.of("key", "oncall", "value", "platform-ops")))
                    .build());
        });
    }
}
