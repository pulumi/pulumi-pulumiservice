package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.AzureDevOpsIntegration;
import com.pulumi.pulumiservice.v2.AzureDevOpsIntegrationArgs;
import com.pulumi.pulumiservice.v2.BitBucketIntegration;
import com.pulumi.pulumiservice.v2.BitBucketIntegrationArgs;
import com.pulumi.pulumiservice.v2.GitHubEnterpriseIntegration;
import com.pulumi.pulumiservice.v2.GitHubEnterpriseIntegrationArgs;
import com.pulumi.pulumiservice.v2.GitHubIntegration;
import com.pulumi.pulumiservice.v2.GitHubIntegrationArgs;
import com.pulumi.pulumiservice.v2.GitLabIntegration;
import com.pulumi.pulumiservice.v2.GitLabIntegrationArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var githubId = config.get("githubIntegrationId").orElse("gh-org-integration");
            var githubEnterpriseId = config.get("githubEnterpriseIntegrationId").orElse("ghe-org-integration");
            var gitlabId = config.get("gitlabIntegrationId").orElse("gl-org-integration");
            var bitbucketId = config.get("bitbucketIntegrationId").orElse("bb-org-integration");
            var azureDevOpsId = config.get("azureDevOpsIntegrationId").orElse("ado-org-integration");

            new GitHubIntegration("github",
                GitHubIntegrationArgs.builder()
                    .orgName(serviceOrg)
                    .integrationId(githubId)
                    .disablePRComments(false)
                    .disableDetailedDiff(false)
                    .disableNeoSummaries(false)
                    .disableCodeAccessForReviews(false)
                    .build());

            new GitHubEnterpriseIntegration("githubEnterprise",
                GitHubEnterpriseIntegrationArgs.builder()
                    .orgName(serviceOrg)
                    .integrationId(githubEnterpriseId)
                    .disablePRComments(true)
                    .disableDetailedDiff(false)
                    .disableNeoSummaries(false)
                    .disableCodeAccessForReviews(true)
                    .build());

            new GitLabIntegration("gitlab",
                GitLabIntegrationArgs.builder()
                    .orgName(serviceOrg)
                    .integrationId(gitlabId)
                    .disablePRComments(false)
                    .disableDetailedDiff(false)
                    .disableNeoSummaries(true)
                    .build());

            new BitBucketIntegration("bitbucket",
                BitBucketIntegrationArgs.builder()
                    .orgName(serviceOrg)
                    .integrationId(bitbucketId)
                    .disablePRComments(false)
                    .disableDetailedDiff(false)
                    .disableNeoSummaries(false)
                    .build());

            new AzureDevOpsIntegration("azureDevOps",
                AzureDevOpsIntegrationArgs.builder()
                    .orgName(serviceOrg)
                    .integrationId(azureDevOpsId)
                    .disablePRComments(true)
                    .disableDetailedDiff(true)
                    .disableNeoSummaries(true)
                    .build());
        });
    }
}
