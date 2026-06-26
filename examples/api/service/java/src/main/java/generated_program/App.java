package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_services.Service;
import com.pulumi.pulumiservice.api_services.ServiceArgs;
import com.pulumi.pulumiservice.api.inputs.AddServiceItemArgs;
import com.pulumi.pulumiservice.api.inputs.ServicePropertyArgs;

import java.util.List;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var serviceSuffix = config.get("serviceSuffix").orElse("dev");

            new Service("catalogService",
                ServiceArgs.builder()
                    .orgName(organizationName)
                    .name("api-service-" + serviceSuffix)
                    .description("An example api service catalog entry.")
                    .ownerType("team")
                    .ownerName("platform")
                    .items(List.of(
                        AddServiceItemArgs.builder()
                            .type("stack")
                            .name("service-provider-test-org/example-app/dev")
                            .build()))
                    .properties(List.of(
                        ServicePropertyArgs.builder()
                            .key("tier").value("gold").type("string").order(1).build(),
                        ServicePropertyArgs.builder()
                            .key("oncall").value("platform-ops").type("string").order(2).build()))
                    .build());
        });
    }
}
