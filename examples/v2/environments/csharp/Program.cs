using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "test-project";
    var envSuffix = config.Get("envSuffix") ?? "dev";

    var environment = new Ps.V2.Esc.Environment("environment", new()
    {
        OrgName = serviceOrg,
        Project = projectName,
        Name = $"testing-environment-{envSuffix}",
    });

    return new Dictionary<string, object?>
    {
        ["envName"] = environment.Name,
    };
});
