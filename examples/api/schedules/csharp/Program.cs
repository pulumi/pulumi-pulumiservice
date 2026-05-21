using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "pulumi-service-schedules-example";
    var stackName = config.Get("stackName") ?? "dev";
    var scheduleCron = config.Get("scheduleCron") ?? "0 7 * * *";

    var parentStack = new Ps.Api.Stacks.Stack("parentStack", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
    });

    var parentSettings = new Ps.Api.Deployments.Settings("parentSettings", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
        SourceContext = ImmutableDictionary.CreateRange(new[]
        {
            new KeyValuePair<string, object>("git", ImmutableDictionary.CreateRange(new[]
            {
                new KeyValuePair<string, object>("repoUrl", "https://github.com/example/example.git"),
                new KeyValuePair<string, object>("branch", "refs/heads/main"),
            })),
        }),
    }, new CustomResourceOptions { DependsOn = { parentStack } });

    var nightlyDeploy = new Ps.Api.Deployments.ScheduledDeployment("nightlyDeploy", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
        ScheduleCron = scheduleCron,
        Request = ImmutableDictionary.CreateRange(new[]
        {
            new KeyValuePair<string, object>("operation", "update"),
        }),
    }, new CustomResourceOptions { DependsOn = { parentSettings } });

    return new Dictionary<string, object?>
    {
        ["nightlyCron"] = nightlyDeploy.ScheduleCron,
    };
});
