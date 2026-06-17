package main

import (
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
	deployments "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/deployments"
	stacks "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/stacks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		organizationName := cfg.Get("organizationName")
		if organizationName == "" {
			organizationName = "service-provider-test-org"
		}
		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "my-new-project"
		}
		stackName := cfg.Get("stackName")
		if stackName == "" {
			stackName = "dev"
		}
		executorImage := cfg.Get("executorImage")
		if executorImage == "" {
			executorImage = "pulumi-cli"
		}

		parentStack, err := stacks.NewStack(ctx, "parentStack", &stacks.StackArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
		})
		if err != nil {
			return err
		}

		settings, err := deployments.NewSettings(ctx, "settings", &deployments.SettingsArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
			ExecutorContext: api.ExecutorSettingsRequestArgs{
				ExecutorImage: api.DockerImageRequestArgs{
					Reference: pulumi.StringPtr(executorImage),
				}.ToDockerImageRequestPtrOutput(),
			},
			OperationContext: api.OperationContextRequestArgs{
				PreRunCommands:       pulumi.ToStringArray([]string{"yarn"}),
				EnvironmentVariables: pulumi.StringMap{"TEST_VAR": pulumi.String("foo")},
				Options: api.OperationContextOptionsRequestArgs{
					SkipInstallDependencies: pulumi.BoolPtr(true),
				}.ToOperationContextOptionsRequestPtrOutput(),
			},
			SourceContext: api.SourceContextRequestArgs{
				Git: api.SourceContextGitRequestArgs{
					RepoUrl: pulumi.StringPtr("https://github.com/example/example.git"),
					Branch:  pulumi.StringPtr("refs/heads/main"),
				}.ToSourceContextGitRequestPtrOutput(),
			},
		}, pulumi.DependsOn([]pulumi.Resource{parentStack}))
		if err != nil {
			return err
		}

		_ = settings
		ctx.Export("stackId", pulumi.String(organizationName+"/"+projectName+"/"+stackName))
		return nil
	})
}
