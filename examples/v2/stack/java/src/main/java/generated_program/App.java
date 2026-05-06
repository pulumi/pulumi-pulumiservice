package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2_stacks.Stack;
import com.pulumi.pulumiservice.v2_stacks.StackArgs;

import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("pulumi-service-stack-example");
            var stackName = config.get("stackName").orElse("dev");
            var stackPurpose = config.get("stackPurpose").orElse("demo");

            var exampleStack = new Stack("exampleStack",
                StackArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .tags(Map.of(
                        "owner", "pulumicloud-v2-example",
                        "purpose", stackPurpose))
                    .build());

            ctx.export("stackName", exampleStack.stackName());
        });
    }
}
