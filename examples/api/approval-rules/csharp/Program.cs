using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";

    var approvers = new Ps.Api.PolicyGroup("approvers", new()
    {
        OrgName = organizationName,
        Name = "api-approvers",
        EntityType = "stacks",
    });

    return new Dictionary<string, object?>
    {
        ["policyGroupName"] = approvers.Name,
    };
});
