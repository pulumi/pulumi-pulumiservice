package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.Stack;
import com.pulumi.pulumiservice.v2.StackArgs;
import com.pulumi.pulumiservice.v2.DeploymentSettings;
import com.pulumi.pulumiservice.v2.DeploymentSettingsArgs;
import com.pulumi.pulumiservice.v2.ScheduledDeployment;
import com.pulumi.pulumiservice.v2.ScheduledDeploymentArgs;
import com.pulumi.resources.CustomResourceOptions;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("pulumi-service-schedules-example");
            var stackName = config.get("stackName").orElse("dev");
            var scheduleCron = config.get("scheduleCron").orElse("0 7 * * *");

            var parentStack = new Stack("parentStack",
                StackArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            var parentSettings = new DeploymentSettings("parentSettings",
                DeploymentSettingsArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .sourceContext(Map.of("git", Map.of(
                        "repoUrl", "https://github.com/example/example.git",
                        "branch", "refs/heads/main")))
                    .build(),
                CustomResourceOptions.builder().dependsOn(List.of(parentStack)).build());

            var nightlyDeploy = new ScheduledDeployment("nightlyDeploy",
                ScheduledDeploymentArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(projectName)
                    .stackName(stackName)
                    .scheduleCron(scheduleCron)
                    .request(Map.of("operation", "update"))
                    .build(),
                CustomResourceOptions.builder().dependsOn(List.of(parentSettings)).build());

            ctx.export("nightlyCron", nightlyDeploy.scheduleCron());
        });
    }
}
