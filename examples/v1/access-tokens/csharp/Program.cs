using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var tokenSuffix = config.Get("tokenSuffix") ?? "dev";
    var tokenDescription = config.Get("tokenDescription") ?? "example v1 access token";

    var team = new Ps.V1.Teams.Team("team", new()
    {
        OrgName = organizationName,
        Name = $"v1-tokens-team-{tokenSuffix}",
        DisplayName = $"v1 Tokens Team {tokenSuffix}",
        Description = "Owner team for the v1 access-tokens example",
    });

    var orgToken = new Ps.V1.Tokens.OrgToken("orgToken", new()
    {
        OrgName = organizationName,
        Name = $"v1-org-token-{tokenSuffix}",
        Description = tokenDescription,
        Admin = false,
        Expires = 0,
    });

    var teamToken = new Ps.V1.Tokens.TeamToken("teamToken", new()
    {
        OrgName = organizationName,
        TeamName = team.Name,
        Name = $"v1-team-token-{tokenSuffix}",
        Description = tokenDescription,
        Expires = 0,
    });

    new Ps.V1.Tokens.PersonalToken("personalToken", new()
    {
        Description = tokenDescription,
        Expires = 0,
    });

    return new Dictionary<string, object?>
    {
        ["orgTokenId"] = orgToken.TokenId,
        ["teamTokenId"] = teamToken.TokenId,
    };
});
