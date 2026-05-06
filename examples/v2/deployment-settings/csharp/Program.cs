using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "my-new-project";
    var stackName = config.Get("stackName") ?? "dev";
    var executorImage = config.Get("executorImage") ?? "pulumi-cli";

    var parentStack = new Ps.V2.Stacks.Stack("parentStack", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        StackName = stackName,
    });

    var settings = new Ps.V2.Deployments.Settings("settings", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        StackName = stackName,
        ExecutorContext = ImmutableDictionary.CreateRange(new[]
        {
            new KeyValuePair<string, object>("executorImage", executorImage),
        }),
        OperationContext = ImmutableDictionary.CreateRange(new[]
        {
            new KeyValuePair<string, object>("preRunCommands", new[] { "yarn" }),
            new KeyValuePair<string, object>("environmentVariables", ImmutableDictionary.CreateRange(new[]
            {
                new KeyValuePair<string, object>("TEST_VAR", "foo"),
            })),
            new KeyValuePair<string, object>("options", ImmutableDictionary.CreateRange(new[]
            {
                new KeyValuePair<string, object>("skipInstallDependencies", true),
            })),
        }),
        SourceContext = ImmutableDictionary.CreateRange(new[]
        {
            new KeyValuePair<string, object>("git", ImmutableDictionary.CreateRange(new[]
            {
                new KeyValuePair<string, object>("repoUrl", "https://github.com/example/example.git"),
                new KeyValuePair<string, object>("branch", "refs/heads/main"),
            })),
        }),
    }, new CustomResourceOptions { DependsOn = { parentStack } });

    return new Dictionary<string, object?>
    {
        ["stackId"] = settings.StackName,
    };
});
