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
			projectName = "pulumi-service-schedules-example"
		}
		stackName := cfg.Get("stackName")
		if stackName == "" {
			stackName = "dev"
		}
		scheduleCron := cfg.Get("scheduleCron")
		if scheduleCron == "" {
			scheduleCron = "0 7 * * *"
		}

		parentStack, err := stacks.NewStack(ctx, "parentStack", &stacks.StackArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
		})
		if err != nil {
			return err
		}

		parentSettings, err := deployments.NewSettings(ctx, "parentSettings", &deployments.SettingsArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
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

		nightlyDeploy, err := deployments.NewScheduledDeployment(ctx, "nightlyDeploy", &deployments.ScheduledDeploymentArgs{
			OrgName:      pulumi.String(organizationName),
			ProjectName:  pulumi.String(projectName),
			StackName:    pulumi.String(stackName),
			ScheduleCron: pulumi.String(scheduleCron),
			Request: &deployments.CreateDeploymentRequestArgs{
				Operation: pulumi.String("update"),
			},
		}, pulumi.DependsOn([]pulumi.Resource{parentSettings}))
		if err != nil {
			return err
		}

		ctx.Export("nightlyCron", nightlyDeploy.ScheduleCron)
		return nil
	})
}
