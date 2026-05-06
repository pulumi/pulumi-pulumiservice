using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";

    var def = new Ps.V2.DefaultOrganization("default", new()
    {
        OrgName = serviceOrg,
    });

    return new Dictionary<string, object?>
    {
        ["defaultOrg"] = def.OrgName,
    };
});
