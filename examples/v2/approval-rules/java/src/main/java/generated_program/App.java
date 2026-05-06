package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.PolicyGroup;
import com.pulumi.pulumiservice.v2.PolicyGroupArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");

            var approvers = new PolicyGroup("approvers",
                PolicyGroupArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-approvers")
                    .entityType("stacks")
                    .build());

            ctx.export("policyGroupName", approvers.name());
        });
    }
}
