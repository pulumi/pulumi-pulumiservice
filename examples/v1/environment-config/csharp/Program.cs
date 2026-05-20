using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "v1-envcfg-example";
    var envName = config.Get("envName") ?? "v1-envcfg-env";

    var draft = new Ps.V1.Esc.EnvironmentDraft("draft", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        EnvName = envName,
    });

    var settings = new Ps.V1.Esc.EnvironmentSettings("settings", new()
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
