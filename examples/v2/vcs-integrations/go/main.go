package main

import (
	v2 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2"
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

		if _, err := v2.NewGitHubIntegration(ctx, "github", &v2.GitHubIntegrationArgs{
			OrgName:                     pulumi.String(serviceOrg),
			IntegrationId:               pulumi.String(githubID),
			DisablePRComments:           pulumi.Bool(false),
			DisableDetailedDiff:         pulumi.Bool(false),
			DisableNeoSummaries:         pulumi.Bool(false),
			DisableCodeAccessForReviews: pulumi.Bool(false),
		}); err != nil {
			return err
		}

		if _, err := v2.NewGitHubEnterpriseIntegration(ctx, "githubEnterprise", &v2.GitHubEnterpriseIntegrationArgs{
			OrgName:                     pulumi.String(serviceOrg),
			IntegrationId:               pulumi.String(githubEnterpriseID),
			DisablePRComments:           pulumi.Bool(true),
			DisableDetailedDiff:         pulumi.Bool(false),
			DisableNeoSummaries:         pulumi.Bool(false),
			DisableCodeAccessForReviews: pulumi.Bool(true),
		}); err != nil {
			return err
		}

		if _, err := v2.NewGitLabIntegration(ctx, "gitlab", &v2.GitLabIntegrationArgs{
			OrgName:             pulumi.String(serviceOrg),
			IntegrationId:       pulumi.String(gitlabID),
			DisablePRComments:   pulumi.Bool(false),
			DisableDetailedDiff: pulumi.Bool(false),
			DisableNeoSummaries: pulumi.Bool(true),
		}); err != nil {
			return err
		}

		if _, err := v2.NewBitBucketIntegration(ctx, "bitbucket", &v2.BitBucketIntegrationArgs{
			OrgName:             pulumi.String(serviceOrg),
			IntegrationId:       pulumi.String(bitbucketID),
			DisablePRComments:   pulumi.Bool(false),
			DisableDetailedDiff: pulumi.Bool(false),
			DisableNeoSummaries: pulumi.Bool(false),
		}); err != nil {
			return err
		}

		if _, err := v2.NewAzureDevOpsIntegration(ctx, "azureDevOps", &v2.AzureDevOpsIntegrationArgs{
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
