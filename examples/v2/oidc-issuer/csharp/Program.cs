using System.Collections.Generic;
using System.Linq;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() => 
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org2";
    var pulumiIssuer = new PulumiService.V2.OidcIssuer("pulumiIssuer", new()
    {
        OrgName = serviceOrg,
        Name = "pulumi_issuer",
        Url = "https://api.pulumi.com/oidc",
        Thumbprints = new[]
        {
            "57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da",
        },
    });

    var githubIssuer = new PulumiService.V2.OidcIssuer("githubIssuer", new()
    {
        OrgName = serviceOrg,
        Name = "github_issuer",
        Url = "https://token.actions.githubusercontent.com",
        Thumbprints = new[]
        {
            "caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7",
        },
        MaxExpiration = 3600,
    });

    return new Dictionary<string, object?>
    {
        ["pulumiIssuerUrl"] = pulumiIssuer.Url,
        ["githubIssuerUrl"] = githubIssuer.Url,
    };
});

