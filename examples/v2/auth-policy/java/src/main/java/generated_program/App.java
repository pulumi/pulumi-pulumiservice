package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2_auth.Policy;
import com.pulumi.pulumiservice.v2_auth.PolicyArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var policyId = config.get("policyId").orElse("org");

            new Policy("policy",
                PolicyArgs.builder()
                    .orgName(serviceOrg)
                    .policyId(policyId)
                    .policies(List.of(
                        Map.of(
                            "decision", "allow",
                            "permission", "read",
                            "tokenType", "organization"),
                        Map.of(
                            "decision", "deny",
                            "permission", "admin",
                            "tokenType", "organization")))
                    .build());
        });
    }
}
