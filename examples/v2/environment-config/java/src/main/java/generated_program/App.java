package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2_esc.EnvironmentDraft;
import com.pulumi.pulumiservice.v2_esc.EnvironmentDraftArgs;
import com.pulumi.pulumiservice.v2_esc.EnvironmentSettings;
import com.pulumi.pulumiservice.v2_esc.EnvironmentSettingsArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("v2-envcfg-example");
            var envName = config.get("envName").orElse("v2-envcfg-env");

            var draft = new EnvironmentDraft("draft",
                EnvironmentDraftArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .envName(envName)
                    .build());

            var settings = new EnvironmentSettings("settings",
                EnvironmentSettingsArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .envName(envName)
                    .deletionProtected(true)
                    .build());

            ctx.export("draftId", draft.changeRequestID());
            ctx.export("protected", settings.deletionProtected());
        });
    }
}
