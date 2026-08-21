package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_stacks.Stack;
import com.pulumi.pulumiservice.api_stacks.StackArgs;
import com.pulumi.pulumiservice.api_deployments.Settings;
import com.pulumi.pulumiservice.api_deployments.SettingsArgs;
import com.pulumi.pulumiservice.api_deployments.ScheduledDeployment;
import com.pulumi.pulumiservice.api_deployments.ScheduledDeploymentArgs;
import com.pulumi.pulumiservice.api_deployments.inputs.CreateDeploymentRequestArgs;
import com.pulumi.pulumiservice.api_deployments.inputs.SourceContextGitRequestArgs;
import com.pulumi.pulumiservice.api_deployments.inputs.SourceContextRequestArgs;
import com.pulumi.resources.CustomResourceOptions;

import java.util.List;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("pulumi-service-schedules-example");
            var stackName = config.get("stackName").orElse("dev");
            var scheduleCron = config.get("scheduleCron").orElse("0 7 * * *");

            var parentStack = new Stack("parentStack",
                StackArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            var parentSettings = new Settings("parentSettings",
                SettingsArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .sourceContext(SourceContextRequestArgs.builder()
                        .git(SourceContextGitRequestArgs.builder()
                            .repoUrl("https://github.com/example/example.git")
                            .branch("refs/heads/main")
                            .build())
                        .build())
                    .build(),
                CustomResourceOptions.builder().dependsOn(List.of(parentStack)).build());

            var nightlyDeploy = new ScheduledDeployment("nightlyDeploy",
                ScheduledDeploymentArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .scheduleCron(scheduleCron)
                    .request(CreateDeploymentRequestArgs.builder()
                        .operation("update")
                        .build())
                    .build(),
                CustomResourceOptions.builder().dependsOn(List.of(parentSettings)).build());

            ctx.export("nightlyCron", nightlyDeploy.scheduleCron());
        });
    }
}
