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
			projectName = "pulumi-service-stack-example"
		}
		stackName := cfg.Get("stackName")
		if stackName == "" {
			stackName = "dev"
		}
		stackPurpose := cfg.Get("stackPurpose")
		if stackPurpose == "" {
			stackPurpose = "demo"
		}

		exampleStack, err := stacks.NewStack(ctx, "exampleStack", &stacks.StackArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
			Tags: pulumi.Map{
				"owner":   pulumi.String("pulumicloud-v2-example"),
				"purpose": pulumi.String(stackPurpose),
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("stackName", exampleStack.StackName)
		return nil
	})
}
