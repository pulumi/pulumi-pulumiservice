package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2_stacks.Stack;
import com.pulumi.pulumiservice.v2_stacks.StackArgs;
import com.pulumi.pulumiservice.v2_stacks.Config;
import com.pulumi.pulumiservice.v2_stacks.ConfigArgs;
import com.pulumi.pulumiservice.v2_stacks.Webhook;
import com.pulumi.pulumiservice.v2_stacks.WebhookArgs;

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

            new Config("config",
                ConfigArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(parentStack.projectName())
                    .stackName(parentStack.stackName())
                    .environment(envRef)
                    .build());

            new Webhook("hook",
                WebhookArgs.builder()
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
