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
		vcsSuffix := cfg.Get("vcsSuffix")
		if vcsSuffix == "" {
			vcsSuffix = "dev"
		}
		baseUrl := cfg.Get("baseUrl")
		if baseUrl == "" {
			baseUrl = "https://git.example.invalid"
		}
		envRef := cfg.Get("envRef")
		if envRef == "" {
			envRef = "organization/vcs-credentials"
		}

		integration, err := v2.NewCustomVCSIntegration(ctx, "integration", &v2.CustomVCSIntegrationArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-custom-vcs-" + vcsSuffix),
			BaseUrl:     pulumi.String(baseUrl),
			VcsType:     pulumi.String("gitea"),
			Environment: pulumi.String(envRef),
		})
		if err != nil {
			return err
		}

		repository, err := v2.NewCustomVCSRepository(ctx, "repository", &v2.CustomVCSRepositoryArgs{
			OrgName:       pulumi.String(serviceOrg),
			IntegrationId: integration.IntegrationId,
			Name:          pulumi.String("example-repo-" + vcsSuffix),
			DisplayName:   pulumi.String("Example Repository"),
		})
		if err != nil {
			return err
		}

		ctx.Export("integrationId", integration.IntegrationId)
		ctx.Export("repositoryId", repository.ID())
		return nil
	})
}
