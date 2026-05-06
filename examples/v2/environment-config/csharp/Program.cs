using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "v2-envcfg-example";
    var envName = config.Get("envName") ?? "v2-envcfg-env";

    var draft = new Ps.V2.Esc.EnvironmentDraft("draft", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        EnvName = envName,
    });

    var settings = new Ps.V2.Esc.EnvironmentSettings("settings", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        EnvName = envName,
        DeletionProtected = true,
    });

    return new Dictionary<string, object?>
    {
        ["draftId"] = draft.ChangeRequestID,
        ["protected"] = settings.DeletionProtected,
    };
});
