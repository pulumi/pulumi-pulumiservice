package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.PolicyGroup;
import com.pulumi.pulumiservice.api.PolicyGroupArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var groupName = config.get("groupName").orElse("example-policy-group");

            var group = new PolicyGroup("group",
                PolicyGroupArgs.builder()
                    .orgName(organizationName)
                    .name(groupName)
                    .entityType("stacks")
                    .build());

            ctx.export("policyGroupName", group.name());
        });
    }
}
