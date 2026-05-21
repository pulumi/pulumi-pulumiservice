package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_esc.EnvironmentDraft;
import com.pulumi.pulumiservice.api_esc.EnvironmentDraftArgs;
import com.pulumi.pulumiservice.api_esc.EnvironmentSettings;
import com.pulumi.pulumiservice.api_esc.EnvironmentSettingsArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("api-envcfg-example");
            var envName = config.get("envName").orElse("api-envcfg-env");

            var draft = new EnvironmentDraft("draft",
                EnvironmentDraftArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .envName(envName)
                    .build());

            var settings = new EnvironmentSettings("settings",
                EnvironmentSettingsArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .envName(envName)
                    .deletionProtected(true)
                    .build());

            ctx.export("draftId", draft.changeRequestId());
            ctx.export("protected", settings.deletionProtected());
        });
    }
}
