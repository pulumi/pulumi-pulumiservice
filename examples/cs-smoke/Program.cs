using System.Collections.Generic;
using Pulumi;
using Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Pulumi.Config();
    var org = config.Require("organizationName");
    var name = "test-smoke-cs-" + config.Require("digits");

    var environment = new Pulumi.PulumiService.Environment("environment", new EnvironmentArgs
    {
        Organization = org,
        Project = "test-smoke",
        Name = name,
        Yaml = new StringAsset("values:\n  greeting: hello\n"),
    });

    var token = new OrgAccessToken("token", new OrgAccessTokenArgs
    {
        OrganizationName = org,
        Name = name,
        Description = "Created by the cs-smoke example.",
        Admin = false,
    });

    var lookedUp = GetEnvironment.Invoke(new GetEnvironmentInvokeArgs
    {
        OrganizationName = environment.Organization,
        ProjectName = environment.Project,
        Name = environment.Name,
    });

    return new Dictionary<string, object?>
    {
        ["environmentName"] = lookedUp.Apply(result => result.Name),
        ["tokenValue"] = token.Value,
    };
});
