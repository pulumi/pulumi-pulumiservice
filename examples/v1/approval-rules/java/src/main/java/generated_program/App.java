package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v1.PolicyGroup;
import com.pulumi.pulumiservice.v1.PolicyGroupArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");

            var approvers = new PolicyGroup("approvers",
                PolicyGroupArgs.builder()
                    .orgName(organizationName)
                    .name("v1-approvers")
                    .entityType("stacks")
                    .build());

            ctx.export("policyGroupName", approvers.name());
        });
    }
}
