using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var vcsSuffix = config.Get("vcsSuffix") ?? "dev";
    var baseUrl = config.Get("baseUrl") ?? "https://git.example.invalid";
    var envRef = config.Get("envRef") ?? "organization/vcs-credentials";

    var integration = new Ps.V1.Integrations.CustomVCSIntegration("integration", new()
    {
        OrgName = organizationName,
        Name = $"v1-custom-vcs-{vcsSuffix}",
        BaseUrl = baseUrl,
        VcsType = "gitea",
        Environment = envRef,
    });

    var repository = new Ps.V1.Integrations.CustomVCSRepository("repository", new()
    {
        OrgName = organizationName,
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
