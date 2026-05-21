package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_stacks.Stack;
import com.pulumi.pulumiservice.api_stacks.StackArgs;

import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("pulumi-service-stack-example");
            var stackName = config.get("stackName").orElse("dev");
            var stackPurpose = config.get("stackPurpose").orElse("demo");

            var exampleStack = new Stack("exampleStack",
                StackArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .tags(Map.of(
                        "owner", "pulumicloud-api-example",
                        "purpose", stackPurpose))
                    .build());

            ctx.export("stackName", exampleStack.stackName());
        });
    }
}
