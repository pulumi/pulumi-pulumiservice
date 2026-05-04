package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.Stack;
import com.pulumi.pulumiservice.v2.StackArgs;
import com.pulumi.pulumiservice.v2.StackConfig;
import com.pulumi.pulumiservice.v2.StackConfigArgs;
import com.pulumi.pulumiservice.v2.StackWebhook;
import com.pulumi.pulumiservice.v2.StackWebhookArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("v2-stack-config-example");
            var stackName = config.get("stackName").orElse("dev");
            var hookUrl = config.get("hookUrl").orElse("https://example.invalid/hooks/example");
            var envRef = config.get("envRef").orElse("organization/credentials");

            var parentStack = new Stack("parentStack",
                StackArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            new StackConfig("config",
                StackConfigArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(parentStack.projectName())
                    .stackName(parentStack.stackName())
                    .environment(envRef)
                    .build());

            new StackWebhook("hook",
                StackWebhookArgs.builder()
                    .organizationName(serviceOrg)
                    .projectName(parentStack.projectName())
                    .stackName(parentStack.stackName())
                    .name("v2-stackhook")
                    .displayName("Stack hook example")
                    .payloadUrl(hookUrl)
                    .active(true)
                    .format("pulumi")
                    .build());

            ctx.export("stack", parentStack.id());
        });
    }
}
