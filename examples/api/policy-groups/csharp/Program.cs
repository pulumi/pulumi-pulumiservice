using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var groupName = config.Get("groupName") ?? "example-policy-group";

    var group = new Ps.Api.PolicyGroup("group", new()
    {
        OrgName = organizationName,
        Name = groupName,
        EntityType = "stacks",
    });

    return new Dictionary<string, object?>
    {
        ["policyGroupName"] = group.Name,
    };
});
