// C# variant of canonical/04-deployment-pipeline.
// A git-driven Pulumi Deployments pipeline: source + executor settings,
// drift detection on a cron, TTL on ephemeral stacks. Behavioral twin of
// the sibling YAML program.

using System.Collections.Generic;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Pulumi.Deployment.RunAsync(() =>
{
    var cfg = new Config();
    var organizationName = cfg.Get("organizationName") ?? "service-provider-test-org";
    var project = cfg.Get("project") ?? "infrastructure";
    var stack = cfg.Get("stack") ?? "production";

    _ = new PulumiService.Stacks.Deployments.Settings("deploymentSettings", new()
    {
        Organization = organizationName,
        Project = project,
        Stack = stack,
        SourceContext = new Dictionary<string, object?>
        {
            ["git"] = new Dictionary<string, object?>
            {
                ["repoUrl"] = "https://github.com/acme-corp/infrastructure.git",
                ["branch"] = "refs/heads/main",
                ["repoDir"] = "stacks/production",
            },
        },
        Github = new Dictionary<string, object?>
        {
            ["repository"] = "acme-corp/infrastructure",
            ["deployCommits"] = true,
            ["previewPullRequests"] = true,
            ["pullRequestTemplate"] = true,
        },
        ExecutorContext = new Dictionary<string, object?>
        {
            ["executorImage"] = new Dictionary<string, object?> { ["image"] = "pulumi/pulumi:latest" },
        },
        OperationContext = new Dictionary<string, object?>
        {
            ["preRunCommands"] = new[] { "npm ci" },
            ["environmentVariables"] = new Dictionary<string, object?> { ["NODE_ENV"] = "production" },
        },
    });

    _ = new PulumiService.Stacks.Deployments.DriftSchedule("driftCheck", new()
    {
        Organization = organizationName,
        Project = project,
        Stack = stack,
        ScheduleCron = "0 */6 * * *",
        AutoRemediate = false,
    });

    _ = new PulumiService.Stacks.Deployments.TtlSchedule("ephemeralTtl", new()
    {
        Organization = organizationName,
        Project = project,
        Stack = stack,
        Timestamp = "2026-12-31T00:00:00Z",
        DeleteAfterDestroy = true,
    });

    return new Dictionary<string, object?>();
});
