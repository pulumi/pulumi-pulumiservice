// Java variant of canonical/04-deployment-pipeline.
// A git-driven Pulumi Deployments pipeline: source + executor settings,
// drift detection on a cron, TTL on ephemeral stacks. Behavioral twin of
// the sibling YAML program.

import java.util.List;
import java.util.Map;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.stacks_deployments.Settings;
import com.pulumi.pulumiservice.stacks_deployments.SettingsArgs;
import com.pulumi.pulumiservice.stacks_deployments.DriftSchedule;
import com.pulumi.pulumiservice.stacks_deployments.DriftScheduleArgs;
import com.pulumi.pulumiservice.stacks_deployments.TtlSchedule;
import com.pulumi.pulumiservice.stacks_deployments.TtlScheduleArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    private static void stack(Context ctx) {
        var cfg = ctx.config();
        var organizationName = cfg.get("organizationName").orElse("service-provider-test-org");
        var project = cfg.get("project").orElse("infrastructure");
        var stack = cfg.get("stack").orElse("production");

        new Settings("deploymentSettings", SettingsArgs.builder()
            .organization(organizationName)
            .project(project)
            .stack(stack)
            .sourceContext(Map.of("git", Map.of(
                "repoUrl", "https://github.com/acme-corp/infrastructure.git",
                "branch", "refs/heads/main",
                "repoDir", "stacks/production")))
            .github(Map.of(
                "repository", "acme-corp/infrastructure",
                "deployCommits", true,
                "previewPullRequests", true,
                "pullRequestTemplate", true))
            .executorContext(Map.of("executorImage", Map.of("image", "pulumi/pulumi:latest")))
            .operationContext(Map.of(
                "preRunCommands", List.of("npm ci"),
                "environmentVariables", Map.of("NODE_ENV", "production")))
            .build());

        new DriftSchedule("driftCheck", DriftScheduleArgs.builder()
            .organization(organizationName)
            .project(project)
            .stack(stack)
            .scheduleCron("0 */6 * * *")
            .autoRemediate(false)
            .build());

        new TtlSchedule("ephemeralTtl", TtlScheduleArgs.builder()
            .organization(organizationName)
            .project(project)
            .stack(stack)
            .timestamp("2026-12-31T00:00:00Z")
            .deleteAfterDestroy(true)
            .build());
    }
}
