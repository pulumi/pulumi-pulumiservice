using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "test-project";
    var envSuffix = config.Get("envSuffix") ?? "dev";

    var environment = new Ps.Api.Esc.Environment("environment", new()
    {
        OrgName = organizationName,
        Project = projectName,
        Name = $"testing-environment-{envSuffix}",
    });

    return new Dictionary<string, object?>
    {
        ["environmentId"] = environment.Id,
    };
});
