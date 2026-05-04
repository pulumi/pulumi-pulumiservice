using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "pulumi-service-schedules-example";
    var stackName = config.Get("stackName") ?? "dev";
    var scheduleCron = config.Get("scheduleCron") ?? "0 7 * * *";

    var parentStack = new Ps.V2.Stack("parentStack", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        StackName = stackName,
    });

    var parentSettings = new Ps.V2.DeploymentSettings("parentSettings", new()
    {
        OrgName = serviceOrg,
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

    var nightlyDeploy = new Ps.V2.ScheduledDeployment("nightlyDeploy", new()
    {
        OrgName = serviceOrg,
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
