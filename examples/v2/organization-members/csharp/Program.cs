using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var memberLogin = config.Get("memberLogin") ?? "pulumi-bot";
    var memberRole = config.Get("memberRole") ?? "member";

    var member = new Ps.V2.OrganizationMember("member", new()
    {
        OrgName = organizationName,
        UserLogin = memberLogin,
        Role = memberRole,
    });

    return new Dictionary<string, object?>
    {
        ["memberId"] = member.Id,
    };
});
