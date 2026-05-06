package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.PolicyIssue;
import com.pulumi.pulumiservice.v2.PolicyIssueArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var issueId = config.get("issueId").orElse("example-issue-id");

            new PolicyIssue("issue",
                PolicyIssueArgs.builder()
                    .orgName(serviceOrg)
                    .issueId(issueId)
                    .priority("high")
                    .status("in_progress")
                    .assignedTo("pulumi-bot")
                    .build());
        });
    }
}
