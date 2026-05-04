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
			projectName = "test-project"
		}
		envSuffix := cfg.Get("envSuffix")
		if envSuffix == "" {
			envSuffix = "dev"
		}

		environment, err := v2.NewEnvironment_esc_environments(ctx, "environment", &v2.Environment_esc_environmentsArgs{
			OrgName: pulumi.String(serviceOrg),
			Project: pulumi.String(projectName),
			Name:    pulumi.String("testing-environment-" + envSuffix),
		})
		if err != nil {
			return err
		}

		ctx.Export("envName", environment.Name)
		return nil
	})
}
