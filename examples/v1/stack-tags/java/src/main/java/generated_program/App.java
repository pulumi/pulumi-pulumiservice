package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v1_stacks.Stack;
import com.pulumi.pulumiservice.v1_stacks.StackArgs;
import com.pulumi.pulumiservice.v1_stacks.Tag;
import com.pulumi.pulumiservice.v1_stacks.TagArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("v1-stack-tags-example");
            var stackName = config.get("stackName").orElse("dev");
            var tagValue = config.get("tagValue").orElse("v1-tag-value");

            var parentStack = new Stack("parentStack",
                StackArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            new Tag("ownerTag",
                TagArgs.builder()
                    .orgName(organizationName)
                    .projectName(parentStack.projectName())
                    .stackName(parentStack.stackName())
                    .name("owner")
                    .value("pulumicloud-v1-example")
                    .build());

            new Tag("customTag",
                TagArgs.builder()
                    .orgName(organizationName)
                    .projectName(parentStack.projectName())
                    .stackName(parentStack.stackName())
                    .name("purpose")
                    .value(tagValue)
                    .build());

            ctx.export("parent", parentStack.id());
        });
    }
}
