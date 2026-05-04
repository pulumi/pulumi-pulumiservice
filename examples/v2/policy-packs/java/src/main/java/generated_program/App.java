package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.PolicyPack;
import com.pulumi.pulumiservice.v2.PolicyPackArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");

            var pack = new PolicyPack("pack",
                PolicyPackArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-example-policy-pack")
                    .displayName("v2 example policy pack")
                    .description("Demo policy pack created via v2 metadata-driven provider.")
                    .policies(List.of(Map.of(
                        "name", "no-public-buckets",
                        "description", "Reject S3 buckets with public ACLs",
                        "enforcementLevel", "advisory")))
                    .build());

            ctx.export("policyPackName", pack.name());
        });
    }
}
