package main

import (
	agents "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/agents"
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
		taskSuffix := cfg.Get("taskSuffix")
		if taskSuffix == "" {
			taskSuffix = "dev"
		}
		taskID := cfg.Get("taskID")
		if taskID == "" {
			taskID = "example-task-id"
		}

		pool, err := agents.NewPool(ctx, "pool", &agents.PoolArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("v1-task-pool-" + taskSuffix),
			Description: pulumi.String("Pool used by the v1 task example"),
		})
		if err != nil {
			return err
		}

		if _, err := agents.NewTask(ctx, "task", &agents.TaskArgs{
			OrgName:        pulumi.String(organizationName),
			TaskID:         pulumi.String(taskID),
			ApprovalMode:   pulumi.String("manual"),
			PermissionMode: pulumi.String("default"),
			Source:         pulumi.String("api"),
			PlanMode:       pulumi.Bool(false),
		}); err != nil {
			return err
		}

		ctx.Export("poolName", pool.Name)
		return nil
	})
}
