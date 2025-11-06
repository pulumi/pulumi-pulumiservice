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

		task, err := pulumiservice.CreateTask(ctx, &pulumiservice.CreateTaskArgs{
			Content:          "Hello Neo!",
			OrganizationName: organizationName,
		})
		if err != nil {
			return err
		}
		ctx.Export("taskId", task.Id)

		return nil
	})
}
