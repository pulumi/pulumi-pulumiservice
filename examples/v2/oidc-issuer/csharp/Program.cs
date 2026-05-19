using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var issuerSuffix = config.Get("issuerSuffix") ?? "dev";
    var maxExpiration = config.GetInt32("maxExpiration") ?? 3600;

    var pulumiIssuer = new Ps.V2.Auth.OidcIssuer("pulumiIssuer", new()
    {
        OrgName = serviceOrg,
        Name = $"pulumi_issuer_{issuerSuffix}",
        Url = "https://api.pulumi.com/oidc",
        Thumbprints = new[] { "57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da" },
    });

    var githubIssuer = new Ps.V2.Auth.OidcIssuer("githubIssuer", new()
    {
        OrgName = serviceOrg,
        Name = $"github_issuer_{issuerSuffix}",
        Url = "https://token.actions.githubusercontent.com",
        Thumbprints = new[] { "b41ae0832808ebc94951437bf7e92b93ccb6479364daf894d46d6001bee7a486" },
        MaxExpiration = maxExpiration,
    });

    return new Dictionary<string, object?>
    {
        ["pulumiIssuerName"] = pulumiIssuer.Name,
        ["githubIssuerName"] = githubIssuer.Name,
    };
});
