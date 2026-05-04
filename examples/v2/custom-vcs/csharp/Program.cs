using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var vcsSuffix = config.Get("vcsSuffix") ?? "dev";
    var baseUrl = config.Get("baseUrl") ?? "https://git.example.invalid";
    var envRef = config.Get("envRef") ?? "organization/vcs-credentials";

    var integration = new Ps.V2.CustomVCSIntegration("integration", new()
    {
        OrgName = serviceOrg,
        Name = $"v2-custom-vcs-{vcsSuffix}",
        BaseUrl = baseUrl,
        VcsType = "gitea",
        Environment = envRef,
    });

    var repository = new Ps.V2.CustomVCSRepository("repository", new()
    {
        OrgName = serviceOrg,
        IntegrationId = integration.IntegrationId,
        Name = $"example-repo-{vcsSuffix}",
        DisplayName = "Example Repository",
    });

    return new Dictionary<string, object?>
    {
        ["integrationId"] = integration.IntegrationId,
        ["repositoryId"] = repository.Id,
    };
});
