package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.DefaultOrganization;
import com.pulumi.pulumiservice.api.DefaultOrganizationArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");

            var def = new DefaultOrganization("default",
                DefaultOrganizationArgs.builder()
                    .orgName(organizationName)
                    .build());

            ctx.export("defaultOrg", com.pulumi.core.Output.of(organizationName));
            ctx.export("defaultOrgGitHubLogin", def.GitHubLogin());
        });
    }
}
