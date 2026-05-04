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
		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "v2-stack-tags-example"
		}
		stackName := cfg.Get("stackName")
		if stackName == "" {
			stackName = "dev"
		}
		tagValue := cfg.Get("tagValue")
		if tagValue == "" {
			tagValue = "v2-tag-value"
		}

		parentStack, err := v2.NewStack(ctx, "parentStack", &v2.StackArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
		})
		if err != nil {
			return err
		}

		if _, err := v2.NewStackTag(ctx, "ownerTag", &v2.StackTagArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: parentStack.ProjectName,
			StackName:   parentStack.StackName,
			Name:        pulumi.String("owner"),
			Value:       pulumi.String("pulumicloud-v2-example"),
		}); err != nil {
			return err
		}

		if _, err := v2.NewStackTag(ctx, "customTag", &v2.StackTagArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: parentStack.ProjectName,
			StackName:   parentStack.StackName,
			Name:        pulumi.String("purpose"),
			Value:       pulumi.String(tagValue),
		}); err != nil {
			return err
		}

		ctx.Export("parent", parentStack.ID())
		return nil
	})
}
