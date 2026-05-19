using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var poolSuffix = config.Get("poolSuffix") ?? "dev";
    var poolDescription = config.Get("poolDescription") ?? "v2 example agent pool";

    var pool = new Ps.V2.Agents.Pool("pool", new()
    {
        OrgName = organizationName,
        Name = $"v2-agent-pool-{poolSuffix}",
        Description = poolDescription,
    });

    return new Dictionary<string, object?>
    {
        ["poolName"] = pool.Name,
    };
});
