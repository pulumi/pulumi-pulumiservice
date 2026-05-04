using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";

    var approvers = new Ps.V2.PolicyGroup("approvers", new()
    {
        OrgName = serviceOrg,
        Name = "v2-approvers",
        EntityType = "stacks",
    });

    return new Dictionary<string, object?>
    {
        ["policyGroupName"] = approvers.Name,
    };
});
