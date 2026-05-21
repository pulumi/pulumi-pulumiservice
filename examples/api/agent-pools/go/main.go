package main

import (
	agents "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/agents"
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
		poolSuffix := cfg.Get("poolSuffix")
		if poolSuffix == "" {
			poolSuffix = "dev"
		}
		poolDescription := cfg.Get("poolDescription")
		if poolDescription == "" {
			poolDescription = "api example agent pool"
		}

		pool, err := agents.NewPool(ctx, "pool", &agents.PoolArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-agent-pool-" + poolSuffix),
			Description: pulumi.String(poolDescription),
		})
		if err != nil {
			return err
		}

		ctx.Export("poolName", pool.Name)
		return nil
	})
}
