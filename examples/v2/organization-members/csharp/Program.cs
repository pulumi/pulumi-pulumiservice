using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var memberLogin = config.Get("memberLogin") ?? "pulumi-bot";
    var memberRole = config.Get("memberRole") ?? "member";

    var member = new Ps.V2.OrganizationMember("member", new()
    {
        OrgName = serviceOrg,
        UserLogin = memberLogin,
        Role = memberRole,
    });

    return new Dictionary<string, object?>
    {
        ["memberId"] = member.Id,
    };
});
