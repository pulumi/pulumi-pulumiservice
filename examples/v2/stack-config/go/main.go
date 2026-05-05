package main

import (
	stacks "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/stacks"
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
		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "v2-stack-config-example"
		}
		stackName := cfg.Get("stackName")
		if stackName == "" {
			stackName = "dev"
		}
		hookUrl := cfg.Get("hookUrl")
		if hookUrl == "" {
			hookUrl = "https://example.invalid/hooks/example"
		}
		envRef := cfg.Get("envRef")
		if envRef == "" {
			envRef = "organization/credentials"
		}

		parentStack, err := stacks.NewStack(ctx, "parentStack", &stacks.StackArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
		})
		if err != nil {
			return err
		}

		if _, err := stacks.NewConfig(ctx, "config", &stacks.ConfigArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: parentStack.ProjectName,
			StackName:   parentStack.StackName,
			Environment: pulumi.String(envRef),
		}); err != nil {
			return err
		}

		if _, err := stacks.NewWebhook(ctx, "hook", &stacks.WebhookArgs{
			OrganizationName: pulumi.String(serviceOrg),
			ProjectName:      parentStack.ProjectName,
			StackName:        parentStack.StackName,
			Name:             pulumi.String("v2-stackhook"),
			DisplayName:      pulumi.String("Stack hook example"),
			PayloadUrl:       pulumi.String(hookUrl),
			Active:           pulumi.Bool(true),
			Format:           pulumi.String("pulumi"),
		}); err != nil {
			return err
		}

		ctx.Export("stack", parentStack.ID())
		return nil
	})
}
