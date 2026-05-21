package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_integrations.CustomVCSIntegration;
import com.pulumi.pulumiservice.api_integrations.CustomVCSIntegrationArgs;
import com.pulumi.pulumiservice.api_integrations.CustomVCSRepository;
import com.pulumi.pulumiservice.api_integrations.CustomVCSRepositoryArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var vcsSuffix = config.get("vcsSuffix").orElse("dev");
            var baseUrl = config.get("baseUrl").orElse("https://git.example.invalid");
            var envRef = config.get("envRef").orElse("organization/vcs-credentials");

            var integration = new CustomVCSIntegration("integration",
                CustomVCSIntegrationArgs.builder()
                    .orgName(organizationName)
                    .name("api-custom-vcs-" + vcsSuffix)
                    .baseUrl(baseUrl)
                    .vcsType("gitea")
                    .environment(envRef)
                    .build());

            var repository = new CustomVCSRepository("repository",
                CustomVCSRepositoryArgs.builder()
                    .orgName(organizationName)
                    .integrationId(integration.integrationId())
                    .name("example-repo-" + vcsSuffix)
                    .displayName("Example Repository")
                    .build());

            ctx.export("integrationId", integration.integrationId());
            ctx.export("repositoryId", repository.id());
        });
    }
}
