import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const githubIntegrationId = config.get("githubIntegrationId") ?? "gh-org-integration";
const githubEnterpriseIntegrationId = config.get("githubEnterpriseIntegrationId") ?? "ghe-org-integration";
const gitlabIntegrationId = config.get("gitlabIntegrationId") ?? "gl-org-integration";
const bitbucketIntegrationId = config.get("bitbucketIntegrationId") ?? "bb-org-integration";
const azureDevOpsIntegrationId = config.get("azureDevOpsIntegrationId") ?? "ado-org-integration";

new ps.v2.integrations.GitHubIntegration("github", {
    orgName: serviceOrg,
    integrationId: githubIntegrationId,
    disablePRComments: false,
    disableDetailedDiff: false,
    disableNeoSummaries: false,
    disableCodeAccessForReviews: false,
});

new ps.v2.integrations.GitHubEnterpriseIntegration("githubEnterprise", {
    orgName: serviceOrg,
    integrationId: githubEnterpriseIntegrationId,
    disablePRComments: true,
    disableDetailedDiff: false,
    disableNeoSummaries: false,
    disableCodeAccessForReviews: true,
});

new ps.v2.integrations.GitLabIntegration("gitlab", {
    orgName: serviceOrg,
    integrationId: gitlabIntegrationId,
    disablePRComments: false,
    disableDetailedDiff: false,
    disableNeoSummaries: true,
});

new ps.v2.integrations.BitBucketIntegration("bitbucket", {
    orgName: serviceOrg,
    integrationId: bitbucketIntegrationId,
    disablePRComments: false,
    disableDetailedDiff: false,
    disableNeoSummaries: false,
});

new ps.v2.integrations.AzureDevOpsIntegration("azureDevOps", {
    orgName: serviceOrg,
    integrationId: azureDevOpsIntegrationId,
    disablePRComments: true,
    disableDetailedDiff: true,
    disableNeoSummaries: true,
});
