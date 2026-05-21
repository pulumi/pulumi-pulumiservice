using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "api-envcfg-example";
    var envName = config.Get("envName") ?? "api-envcfg-env";

    var draft = new Ps.Api.Esc.EnvironmentDraft("draft", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        EnvName = envName,
    });

    var settings = new Ps.Api.Esc.EnvironmentSettings("settings", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        EnvName = envName,
        DeletionProtected = true,
    });

    return new Dictionary<string, object?>
    {
        ["draftId"] = draft.ChangeRequestId,
        ["protected"] = settings.DeletionProtected,
    };
});
