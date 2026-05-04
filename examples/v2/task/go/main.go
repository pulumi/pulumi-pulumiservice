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
		taskSuffix := cfg.Get("taskSuffix")
		if taskSuffix == "" {
			taskSuffix = "dev"
		}
		taskID := cfg.Get("taskID")
		if taskID == "" {
			taskID = "example-task-id"
		}

		pool, err := v2.NewAgentPool(ctx, "pool", &v2.AgentPoolArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-task-pool-" + taskSuffix),
			Description: pulumi.String("Pool used by the v2 task example"),
		})
		if err != nil {
			return err
		}

		if _, err := v2.NewTask(ctx, "task", &v2.TaskArgs{
			OrgName:        pulumi.String(serviceOrg),
			TaskID:         pulumi.String(taskID),
			Name:           pulumi.String("v2-task-" + taskSuffix),
			ApprovalMode:   pulumi.String("manual"),
			PermissionMode: pulumi.String("maintainer"),
			IsShared:       pulumi.Bool(false),
		}); err != nil {
			return err
		}

		ctx.Export("poolName", pool.Name)
		return nil
	})
}
