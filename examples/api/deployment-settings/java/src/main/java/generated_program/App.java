package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_stacks.Stack;
import com.pulumi.pulumiservice.api_stacks.StackArgs;
import com.pulumi.pulumiservice.api_deployments.Settings;
import com.pulumi.pulumiservice.api_deployments.SettingsArgs;
import com.pulumi.pulumiservice.api.inputs.DockerImageRequestArgs;
import com.pulumi.pulumiservice.api.inputs.ExecutorSettingsRequestArgs;
import com.pulumi.pulumiservice.api.inputs.OperationContextOptionsRequestArgs;
import com.pulumi.pulumiservice.api.inputs.OperationContextRequestArgs;
import com.pulumi.pulumiservice.api.inputs.SourceContextGitRequestArgs;
import com.pulumi.pulumiservice.api.inputs.SourceContextRequestArgs;
import com.pulumi.resources.CustomResourceOptions;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("my-new-project");
            var stackName = config.get("stackName").orElse("dev");
            var executorImage = config.get("executorImage").orElse("pulumi-cli");

            var parentStack = new Stack("parentStack",
                StackArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .build());

            var settings = new Settings("settings",
                SettingsArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .stackName(stackName)
                    .executorContext(ExecutorSettingsRequestArgs.builder()
                        .executorImage(DockerImageRequestArgs.builder()
                            .reference(executorImage)
                            .build())
                        .build())
                    .operationContext(OperationContextRequestArgs.builder()
                        .preRunCommands(List.of("yarn"))
                        .environmentVariables(Map.of("TEST_VAR", "foo"))
                        .options(OperationContextOptionsRequestArgs.builder()
                            .skipInstallDependencies(true)
                            .build())
                        .build())
                    .sourceContext(SourceContextRequestArgs.builder()
                        .git(SourceContextGitRequestArgs.builder()
                            .repoUrl("https://github.com/example/example.git")
                            .branch("refs/heads/main")
                            .build())
                        .build())
                    .build(),
                CustomResourceOptions.builder().dependsOn(List.of(parentStack)).build());

            ctx.export("stackId", com.pulumi.core.Output.of(organizationName + "/" + projectName + "/" + stackName));
        });
    }
}
