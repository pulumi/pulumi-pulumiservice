package main

import (
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
			ExecutorContext: &deployments.ExecutorSettingsRequestArgs{
				ExecutorImage: pulumi.String(executorImage),
			},
			OperationContext: &deployments.OperationContextRequestArgs{
				PreRunCommands:       pulumi.StringArray{pulumi.String("yarn")},
				EnvironmentVariables: pulumi.Map{"TEST_VAR": pulumi.String("foo")},
				Options: &deployments.OperationContextOptionsRequestArgs{
					SkipInstallDependencies: pulumi.Bool(true),
				},
			},
			SourceContext: &deployments.SourceContextRequestArgs{
				Git: &deployments.SourceContextGitRequestArgs{
					RepoUrl: pulumi.String("https://github.com/example/example.git"),
					Branch:  pulumi.String("refs/heads/main"),
				},
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
