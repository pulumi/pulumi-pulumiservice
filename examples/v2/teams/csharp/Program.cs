using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var teamSuffix = config.Get("teamSuffix") ?? "dev";
    var teamDescription = config.Get("teamDescription") ?? "A team created by the v2 example.";

    var team = new Ps.V2.Teams.Team("team", new()
    {
        OrgName = organizationName,
        Name = $"v2-team-{teamSuffix}",
        DisplayName = $"v2 Team {teamSuffix}",
        Description = teamDescription,
    });

    return new Dictionary<string, object?>
    {
        ["teamName"] = team.Name,
    };
});
