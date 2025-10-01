package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		config := config.New(ctx, "")
		organizationName := config.Require("organizationName")
		if organizationName == "" {
			organizationName = ctx.Organization()
		}

		task, err := pulumiservice.NewNeoTask(ctx, "exampleTask", &pulumiservice.NeoTaskArgs{
			Content:          pulumi.String("Hello Neo!"),
			OrganizationName: pulumi.String(organizationName),
		})
		if err != nil {
			return err
		}
		ctx.Export("taskId", task.ID())

		return nil
	})
}
