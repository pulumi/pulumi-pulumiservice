using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "test-project";
    var envSuffix = config.Get("envSuffix") ?? "dev";
    var tagValue = config.Get("tagValue") ?? "env-tag-initial";

    var environment = new Ps.Api.Esc.Environment("environment", new()
    {
        OrgName = organizationName,
        Project = projectName,
        Name = $"testing-environment-{envSuffix}",
    });

    var environmentTag = new Ps.Api.Esc.EnvironmentTag("environmentTag", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        EnvName = $"testing-environment-{envSuffix}",
        Name = "purpose",
        Value = tagValue,
    }, new CustomResourceOptions { DependsOn = { environment } });

    return new Dictionary<string, object?>
    {
        ["environmentId"] = environment.Id,
        ["environmentTagValue"] = environmentTag.Value,
    };
});
