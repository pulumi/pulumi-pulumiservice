package main

import (
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
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
		groupName := cfg.Get("groupName")
		if groupName == "" {
			groupName = "example-attachment-group"
		}
		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "pulumi-service-attachment-example"
		}
		stackName := cfg.Get("stackName")
		if stackName == "" {
			stackName = "dev"
		}

		exampleStack, err := stacks.NewStack(ctx, "exampleStack", &stacks.StackArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String(projectName),
			StackName:   pulumi.String(stackName),
		})
		if err != nil {
			return err
		}

		group, err := api.NewPolicyGroup(ctx, "group", &api.PolicyGroupArgs{
			OrgName:    pulumi.String(organizationName),
			Name:       pulumi.String(groupName),
			EntityType: pulumi.String("stacks"),
		})
		if err != nil {
			return err
		}

		attachment, err := api.NewPolicyGroupStackAttachment(ctx, "attachment", &api.PolicyGroupStackAttachmentArgs{
			OrgName:        pulumi.String(organizationName),
			PolicyGroup:    group.Name,
			Name:           exampleStack.StackName,
			RoutingProject: pulumi.String(projectName),
		}, pulumi.DependsOn([]pulumi.Resource{group, exampleStack}))
		if err != nil {
			return err
		}

		ctx.Export("attachedStack", attachment.Name)
		return nil
	})
}
