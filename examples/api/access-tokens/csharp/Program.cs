using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var tokenSuffix = config.Get("tokenSuffix") ?? "dev";
    var tokenDescription = config.Get("tokenDescription") ?? "example api access token";

    var team = new Ps.Api.Teams.Team("team", new()
    {
        OrgName = organizationName,
        Name = $"api-tokens-team-{tokenSuffix}",
        DisplayName = $"api Tokens Team {tokenSuffix}",
        Description = "Owner team for the api access-tokens example",
    });

    var orgToken = new Ps.Api.Tokens.OrgToken("orgToken", new()
    {
        OrgName = organizationName,
        Name = $"api-org-token-{tokenSuffix}",
        Description = tokenDescription,
        Admin = false,
        Expires = 0,
    });

    var teamToken = new Ps.Api.Tokens.TeamToken("teamToken", new()
    {
        OrgName = organizationName,
        TeamName = team.Name,
        Name = $"api-team-token-{tokenSuffix}",
        Description = tokenDescription,
        Expires = 0,
    });

    new Ps.Api.Tokens.PersonalToken("personalToken", new()
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
