using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";

    var def = new Ps.V2.DefaultOrganization("default", new()
    {
        OrgName = organizationName,
    });

    return new Dictionary<string, object?>
    {
        ["defaultOrg"] = organizationName,
        ["defaultOrgGitHubLogin"] = def.GitHubLogin,
    };
});
