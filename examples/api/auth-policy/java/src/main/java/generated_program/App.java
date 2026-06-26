package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_auth.Policy;
import com.pulumi.pulumiservice.api_auth.PolicyArgs;
import com.pulumi.pulumiservice.api.inputs.AuthPolicyDefinitionArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var policyId = config.get("policyId").orElse("org");

            new Policy("policy",
                PolicyArgs.builder()
                    .orgName(organizationName)
                    .policyId(policyId)
                    .policies(List.of(
                        AuthPolicyDefinitionArgs.builder()
                            .decision("allow")
                            .authorizedPermissions(List.of("read"))
                            .tokenType("organization")
                            .rules(Map.of())
                            .build(),
                        AuthPolicyDefinitionArgs.builder()
                            .decision("deny")
                            .authorizedPermissions(List.of("admin"))
                            .tokenType("organization")
                            .rules(Map.of())
                            .build()))
                    .build());
        });
    }
}
