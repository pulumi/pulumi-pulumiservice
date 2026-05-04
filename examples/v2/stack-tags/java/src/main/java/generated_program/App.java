package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.Stack;
import com.pulumi.pulumiservice.v2.StackArgs;
import com.pulumi.pulumiservice.v2.StackTag;
import com.pulumi.pulumiservice.v2.StackTagArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("v2-stack-tags-example");
            var stackName = config.get("stackName").orElse("dev");
            var tagValue = config.get("tagValue").orElse("v2-tag-value");

            var parentStack = new Stack("parentStack",
                StackArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            new StackTag("ownerTag",
                StackTagArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(parentStack.projectName())
                    .stackName(parentStack.stackName())
                    .name("owner")
                    .value("pulumicloud-v2-example")
                    .build());

            new StackTag("customTag",
                StackTagArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(parentStack.projectName())
                    .stackName(parentStack.stackName())
                    .name("purpose")
                    .value(tagValue)
                    .build());

            ctx.export("parent", parentStack.id());
        });
    }
}
