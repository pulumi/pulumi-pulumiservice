using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var githubId = config.Get("githubIntegrationId") ?? "gh-org-integration";
    var githubEnterpriseId = config.Get("githubEnterpriseIntegrationId") ?? "ghe-org-integration";
    var gitlabId = config.Get("gitlabIntegrationId") ?? "gl-org-integration";
    var bitbucketId = config.Get("bitbucketIntegrationId") ?? "bb-org-integration";
    var azureDevOpsId = config.Get("azureDevOpsIntegrationId") ?? "ado-org-integration";

    new Ps.Api.Integrations.GitHubIntegration("github", new()
    {
        OrgName = organizationName,
        IntegrationId = githubId,
        DisablePRComments = false,
        DisableDetailedDiff = false,
        DisableNeoSummaries = false,
        DisableCodeAccessForReviews = false,
    });

    new Ps.Api.Integrations.GitHubEnterpriseIntegration("githubEnterprise", new()
    {
        OrgName = organizationName,
        IntegrationId = githubEnterpriseId,
        DisablePRComments = true,
        DisableDetailedDiff = false,
        DisableNeoSummaries = false,
        DisableCodeAccessForReviews = true,
    });

    new Ps.Api.Integrations.GitLabIntegration("gitlab", new()
    {
        OrgName = organizationName,
        IntegrationId = gitlabId,
        DisablePRComments = false,
        DisableDetailedDiff = false,
        DisableNeoSummaries = true,
    });

    new Ps.Api.Integrations.BitBucketIntegration("bitbucket", new()
    {
        OrgName = organizationName,
        IntegrationId = bitbucketId,
        DisablePRComments = false,
        DisableDetailedDiff = false,
        DisableNeoSummaries = false,
    });

    new Ps.Api.Integrations.AzureDevOpsIntegration("azureDevOps", new()
    {
        OrgName = organizationName,
        IntegrationId = azureDevOpsId,
        DisablePRComments = true,
        DisableDetailedDiff = true,
        DisableNeoSummaries = true,
    });
});
