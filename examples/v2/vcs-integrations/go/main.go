package main

import (
	integrations "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/integrations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		serviceOrg := cfg.Get("serviceOrg")
		if serviceOrg == "" {
			serviceOrg = "service-provider-test-org"
		}
		githubID := cfg.Get("githubIntegrationId")
		if githubID == "" {
			githubID = "gh-org-integration"
		}
		githubEnterpriseID := cfg.Get("githubEnterpriseIntegrationId")
		if githubEnterpriseID == "" {
			githubEnterpriseID = "ghe-org-integration"
		}
		gitlabID := cfg.Get("gitlabIntegrationId")
		if gitlabID == "" {
			gitlabID = "gl-org-integration"
		}
		bitbucketID := cfg.Get("bitbucketIntegrationId")
		if bitbucketID == "" {
			bitbucketID = "bb-org-integration"
		}
		azureDevOpsID := cfg.Get("azureDevOpsIntegrationId")
		if azureDevOpsID == "" {
			azureDevOpsID = "ado-org-integration"
		}

		if _, err := integrations.NewGitHubIntegration(ctx, "github", &integrations.GitHubIntegrationArgs{
			OrgName:                     pulumi.String(serviceOrg),
			IntegrationId:               pulumi.String(githubID),
			DisablePRComments:           pulumi.Bool(false),
			DisableDetailedDiff:         pulumi.Bool(false),
			DisableNeoSummaries:         pulumi.Bool(false),
			DisableCodeAccessForReviews: pulumi.Bool(false),
		}); err != nil {
			return err
		}

		if _, err := integrations.NewGitHubEnterpriseIntegration(ctx, "githubEnterprise", &integrations.GitHubEnterpriseIntegrationArgs{
			OrgName:                     pulumi.String(serviceOrg),
			IntegrationId:               pulumi.String(githubEnterpriseID),
			DisablePRComments:           pulumi.Bool(true),
			DisableDetailedDiff:         pulumi.Bool(false),
			DisableNeoSummaries:         pulumi.Bool(false),
			DisableCodeAccessForReviews: pulumi.Bool(true),
		}); err != nil {
			return err
		}

		if _, err := integrations.NewGitLabIntegration(ctx, "gitlab", &integrations.GitLabIntegrationArgs{
			OrgName:             pulumi.String(serviceOrg),
			IntegrationId:       pulumi.String(gitlabID),
			DisablePRComments:   pulumi.Bool(false),
			DisableDetailedDiff: pulumi.Bool(false),
			DisableNeoSummaries: pulumi.Bool(true),
		}); err != nil {
			return err
		}

		if _, err := integrations.NewBitBucketIntegration(ctx, "bitbucket", &integrations.BitBucketIntegrationArgs{
			OrgName:             pulumi.String(serviceOrg),
			IntegrationId:       pulumi.String(bitbucketID),
			DisablePRComments:   pulumi.Bool(false),
			DisableDetailedDiff: pulumi.Bool(false),
			DisableNeoSummaries: pulumi.Bool(false),
		}); err != nil {
			return err
		}

		if _, err := integrations.NewAzureDevOpsIntegration(ctx, "azureDevOps", &integrations.AzureDevOpsIntegrationArgs{
			OrgName:             pulumi.String(serviceOrg),
			IntegrationId:       pulumi.String(azureDevOpsID),
			DisablePRComments:   pulumi.Bool(true),
			DisableDetailedDiff: pulumi.Bool(true),
			DisableNeoSummaries: pulumi.Bool(true),
		}); err != nil {
			return err
		}

		return nil
	})
}
