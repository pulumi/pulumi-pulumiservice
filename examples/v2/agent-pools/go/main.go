package main

import (
	agents "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/agents"
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
		poolSuffix := cfg.Get("poolSuffix")
		if poolSuffix == "" {
			poolSuffix = "dev"
		}
		poolDescription := cfg.Get("poolDescription")
		if poolDescription == "" {
			poolDescription = "v2 example agent pool"
		}

		pool, err := agents.NewPool(ctx, "pool", &agents.PoolArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-agent-pool-" + poolSuffix),
			Description: pulumi.String(poolDescription),
		})
		if err != nil {
			return err
		}

		ctx.Export("poolName", pool.Name)
		return nil
	})
}
