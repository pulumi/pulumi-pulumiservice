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
			ExecutorContext: pulumi.Map{
				"executorImage": pulumi.String(executorImage),
			},
			OperationContext: pulumi.Map{
				"preRunCommands":       pulumi.StringArray{pulumi.String("yarn")},
				"environmentVariables": pulumi.StringMap{"TEST_VAR": pulumi.String("foo")},
				"options": pulumi.Map{
					"skipInstallDependencies": pulumi.Bool(true),
				},
			},
			SourceContext: pulumi.Map{
				"git": pulumi.Map{
					"repoUrl": pulumi.String("https://github.com/example/example.git"),
					"branch":  pulumi.String("refs/heads/main"),
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
