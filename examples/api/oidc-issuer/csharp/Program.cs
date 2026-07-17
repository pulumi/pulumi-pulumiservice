using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var issuerSuffix = config.Get("issuerSuffix") ?? "dev";
    var maxExpiration = config.GetInt32("maxExpiration") ?? 3600;
    // Thumbprints must match the certificate the issuer currently serves, so they
    // have no static default. Compute one with:
    //   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
    var pulumiThumbprint = config.Require("pulumiThumbprint");
    var githubThumbprint = config.Require("githubThumbprint");

    var pulumiIssuer = new Ps.Api.Auth.OidcIssuer("pulumiIssuer", new()
    {
        OrgName = organizationName,
        Name = $"pulumi_issuer_{issuerSuffix}",
        Url = "https://api.pulumi.com/oidc",
        Thumbprints = new[] { pulumiThumbprint },
    });

    var githubIssuer = new Ps.Api.Auth.OidcIssuer("githubIssuer", new()
    {
        OrgName = organizationName,
        Name = $"github_issuer_{issuerSuffix}",
        Url = "https://token.actions.githubusercontent.com",
        Thumbprints = new[] { githubThumbprint },
        MaxExpiration = maxExpiration,
    });

    return new Dictionary<string, object?>
    {
        ["pulumiIssuerName"] = pulumiIssuer.Name,
        ["githubIssuerName"] = githubIssuer.Name,
    };
});
