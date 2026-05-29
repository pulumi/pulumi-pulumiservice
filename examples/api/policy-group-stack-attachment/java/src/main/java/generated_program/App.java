package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.PolicyGroup;
import com.pulumi.pulumiservice.api.PolicyGroupArgs;
import com.pulumi.pulumiservice.api.PolicyGroupStackAttachment;
import com.pulumi.pulumiservice.api.PolicyGroupStackAttachmentArgs;
import com.pulumi.pulumiservice.api_stacks.Stack;
import com.pulumi.pulumiservice.api_stacks.StackArgs;
import com.pulumi.resources.CustomResourceOptions;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var groupName = config.get("groupName").orElse("example-attachment-group");
            var projectName = config.get("projectName").orElse("pulumi-service-attachment-example");
            var stackName = config.get("stackName").orElse("dev");

            var exampleStack = new Stack("exampleStack",
                StackArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            var group = new PolicyGroup("group",
                PolicyGroupArgs.builder()
                    .orgName(organizationName)
                    .name(groupName)
                    .entityType("stacks")
                    .build());

            var attachment = new PolicyGroupStackAttachment("attachment",
                PolicyGroupStackAttachmentArgs.builder()
                    .orgName(organizationName)
                    .policyGroup(group.name())
                    .name(exampleStack.stackName())
                    .routingProject(projectName)
                    .build(),
                CustomResourceOptions.builder().dependsOn(group, exampleStack).build());

            ctx.export("attachedStack", attachment.name());
        });
    }
}
