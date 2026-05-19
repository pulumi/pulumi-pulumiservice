import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const githubIntegrationId = config.get("githubIntegrationId") ?? "gh-org-integration";
const githubEnterpriseIntegrationId = config.get("githubEnterpriseIntegrationId") ?? "ghe-org-integration";
const gitlabIntegrationId = config.get("gitlabIntegrationId") ?? "gl-org-integration";
const bitbucketIntegrationId = config.get("bitbucketIntegrationId") ?? "bb-org-integration";
const azureDevOpsIntegrationId = config.get("azureDevOpsIntegrationId") ?? "ado-org-integration";

new ps.v2.integrations.GitHubIntegration("github", {
    orgName: organizationName,
    integrationId: githubIntegrationId,
    disablePRComments: false,
    disableDetailedDiff: false,
    disableNeoSummaries: false,
    disableCodeAccessForReviews: false,
});

new ps.v2.integrations.GitHubEnterpriseIntegration("githubEnterprise", {
    orgName: organizationName,
    integrationId: githubEnterpriseIntegrationId,
    disablePRComments: true,
    disableDetailedDiff: false,
    disableNeoSummaries: false,
    disableCodeAccessForReviews: true,
});

new ps.v2.integrations.GitLabIntegration("gitlab", {
    orgName: organizationName,
    integrationId: gitlabIntegrationId,
    disablePRComments: false,
    disableDetailedDiff: false,
    disableNeoSummaries: true,
});

new ps.v2.integrations.BitBucketIntegration("bitbucket", {
    orgName: organizationName,
    integrationId: bitbucketIntegrationId,
    disablePRComments: false,
    disableDetailedDiff: false,
    disableNeoSummaries: false,
});

new ps.v2.integrations.AzureDevOpsIntegration("azureDevOps", {
    orgName: organizationName,
    integrationId: azureDevOpsIntegrationId,
    disablePRComments: true,
    disableDetailedDiff: true,
    disableNeoSummaries: true,
});
