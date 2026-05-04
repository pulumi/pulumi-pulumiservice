using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var groupName = config.Get("groupName") ?? "example-policy-group";

    var group = new Ps.V2.PolicyGroup("group", new()
    {
        OrgName = serviceOrg,
        Name = groupName,
        EntityType = "stacks",
    });

    return new Dictionary<string, object?>
    {
        ["policyGroupName"] = group.Name,
    };
});
