using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var githubId = config.Get("githubIntegrationId") ?? "gh-org-integration";
    var githubEnterpriseId = config.Get("githubEnterpriseIntegrationId") ?? "ghe-org-integration";
    var gitlabId = config.Get("gitlabIntegrationId") ?? "gl-org-integration";
    var bitbucketId = config.Get("bitbucketIntegrationId") ?? "bb-org-integration";
    var azureDevOpsId = config.Get("azureDevOpsIntegrationId") ?? "ado-org-integration";

    new Ps.V2.GitHubIntegration("github", new()
    {
        OrgName = serviceOrg,
        IntegrationId = githubId,
        DisablePRComments = false,
        DisableDetailedDiff = false,
        DisableNeoSummaries = false,
        DisableCodeAccessForReviews = false,
    });

    new Ps.V2.GitHubEnterpriseIntegration("githubEnterprise", new()
    {
        OrgName = serviceOrg,
        IntegrationId = githubEnterpriseId,
        DisablePRComments = true,
        DisableDetailedDiff = false,
        DisableNeoSummaries = false,
        DisableCodeAccessForReviews = true,
    });

    new Ps.V2.GitLabIntegration("gitlab", new()
    {
        OrgName = serviceOrg,
        IntegrationId = gitlabId,
        DisablePRComments = false,
        DisableDetailedDiff = false,
        DisableNeoSummaries = true,
    });

    new Ps.V2.BitBucketIntegration("bitbucket", new()
    {
        OrgName = serviceOrg,
        IntegrationId = bitbucketId,
        DisablePRComments = false,
        DisableDetailedDiff = false,
        DisableNeoSummaries = false,
    });

    new Ps.V2.AzureDevOpsIntegration("azureDevOps", new()
    {
        OrgName = serviceOrg,
        IntegrationId = azureDevOpsId,
        DisablePRComments = true,
        DisableDetailedDiff = true,
        DisableNeoSummaries = true,
    });
});
