package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.Stack;
import com.pulumi.pulumiservice.v2.StackArgs;
import com.pulumi.pulumiservice.v2.DeploymentSettings;
import com.pulumi.pulumiservice.v2.DeploymentSettingsArgs;
import com.pulumi.resources.CustomResourceOptions;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("my-new-project");
            var stackName = config.get("stackName").orElse("dev");
            var executorImage = config.get("executorImage").orElse("pulumi-cli");

            var parentStack = new Stack("parentStack",
                StackArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            var settings = new DeploymentSettings("settings",
                DeploymentSettingsArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .executorContext(Map.of("executorImage", executorImage))
                    .operationContext(Map.of(
                        "preRunCommands", List.of("yarn"),
                        "environmentVariables", Map.of("TEST_VAR", "foo"),
                        "options", Map.of("skipInstallDependencies", true)))
                    .sourceContext(Map.of(
                        "git", Map.of(
                            "repoUrl", "https://github.com/example/example.git",
                            "branch", "refs/heads/main")))
                    .build(),
                CustomResourceOptions.builder().dependsOn(List.of(parentStack)).build());

            ctx.export("stackId", settings.stackName());
        });
    }
}
