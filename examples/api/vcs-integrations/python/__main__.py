import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
github_id = config.get("githubIntegrationId") or "gh-org-integration"
github_enterprise_id = config.get("githubEnterpriseIntegrationId") or "ghe-org-integration"
gitlab_id = config.get("gitlabIntegrationId") or "gl-org-integration"
bitbucket_id = config.get("bitbucketIntegrationId") or "bb-org-integration"
azure_devops_id = config.get("azureDevOpsIntegrationId") or "ado-org-integration"

ps_api.integrations.GitHubIntegration(
    "github",
    org_name=organization_name,
    integration_id=github_id,
    disable_pr_comments=False,
    disable_detailed_diff=False,
    disable_neo_summaries=False,
    disable_code_access_for_reviews=False,
)

ps_api.integrations.GitHubEnterpriseIntegration(
    "githubEnterprise",
    org_name=organization_name,
    integration_id=github_enterprise_id,
    disable_pr_comments=True,
    disable_detailed_diff=False,
    disable_neo_summaries=False,
    disable_code_access_for_reviews=True,
)

ps_api.integrations.GitLabIntegration(
    "gitlab",
    org_name=organization_name,
    integration_id=gitlab_id,
    disable_pr_comments=False,
    disable_detailed_diff=False,
    disable_neo_summaries=True,
)

ps_api.integrations.BitBucketIntegration(
    "bitbucket",
    org_name=organization_name,
    integration_id=bitbucket_id,
    disable_pr_comments=False,
    disable_detailed_diff=False,
    disable_neo_summaries=False,
)

ps_api.integrations.AzureDevOpsIntegration(
    "azureDevOps",
    org_name=organization_name,
    integration_id=azure_devops_id,
    disable_pr_comments=True,
    disable_detailed_diff=True,
    disable_neo_summaries=True,
)
