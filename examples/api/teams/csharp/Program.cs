using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var teamSuffix = config.Get("teamSuffix") ?? "dev";
    var teamDescription = config.Get("teamDescription") ?? "A team created by the api example.";

    var team = new Ps.Api.Teams.Team("team", new()
    {
        OrgName = organizationName,
        Name = $"api-team-{teamSuffix}",
        DisplayName = $"api Team {teamSuffix}",
        Description = teamDescription,
    });

    return new Dictionary<string, object?>
    {
        ["teamName"] = team.Name,
    };
});
