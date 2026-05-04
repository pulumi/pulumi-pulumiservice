package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.PolicyGroup;
import com.pulumi.pulumiservice.v2.PolicyGroupArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var groupName = config.get("groupName").orElse("example-policy-group");

            var group = new PolicyGroup("group",
                PolicyGroupArgs.builder()
                    .orgName(serviceOrg)
                    .name(groupName)
                    .entityType("stacks")
                    .build());

            ctx.export("policyGroupName", group.name());
        });
    }
}
