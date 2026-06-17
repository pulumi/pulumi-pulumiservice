using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "my-new-project";
    var stackName = config.Get("stackName") ?? "dev";
    var executorImage = config.Get("executorImage") ?? "pulumi-cli";

    var parentStack = new Ps.Api.Stacks.Stack("parentStack", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
    });

    var settings = new Ps.Api.Deployments.Settings("settings", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
        ExecutorContext = new Ps.Api.Inputs.ExecutorSettingsRequestArgs
        {
            ExecutorImage = new Ps.Api.Inputs.DockerImageRequestArgs { Reference = executorImage },
        },
        OperationContext = new Ps.Api.Inputs.OperationContextRequestArgs
        {
            PreRunCommands = { "yarn" },
            EnvironmentVariables = { { "TEST_VAR", "foo" } },
            Options = new Ps.Api.Inputs.OperationContextOptionsRequestArgs { SkipInstallDependencies = true },
        },
        SourceContext = new Ps.Api.Inputs.SourceContextRequestArgs
        {
            Git = new Ps.Api.Inputs.SourceContextGitRequestArgs
            {
                RepoUrl = "https://github.com/example/example.git",
                Branch = "refs/heads/main",
            },
        },
    }, new CustomResourceOptions { DependsOn = { parentStack } });

    _ = settings;
    return new Dictionary<string, object?>
    {
        ["stackId"] = $"{organizationName}/{projectName}/{stackName}",
    };
});
