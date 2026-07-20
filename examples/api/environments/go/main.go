package main

import (
	esc "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/esc"
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
		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "test-project"
		}
		envSuffix := cfg.Get("envSuffix")
		if envSuffix == "" {
			envSuffix = "dev"
		}
		tagValue := cfg.Get("tagValue")
		if tagValue == "" {
			tagValue = "env-tag-initial"
		}

		environment, err := esc.NewEnvironment(ctx, "environment", &esc.EnvironmentArgs{
			OrgName: pulumi.String(organizationName),
			Project: pulumi.String(projectName),
			Name:    pulumi.String("testing-environment-" + envSuffix),
		})
		if err != nil {
			return err
		}

		environmentTag, err := esc.NewEnvironmentTag(ctx, "environmentTag", &esc.EnvironmentTagArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String(projectName),
			EnvName:     pulumi.String("testing-environment-" + envSuffix),
			Name:        pulumi.String("purpose"),
			Value:       pulumi.String(tagValue),
		}, pulumi.DependsOn([]pulumi.Resource{environment}))
		if err != nil {
			return err
		}

		ctx.Export("environmentId", environment.ID())
		ctx.Export("environmentTagValue", environmentTag.Value)
		return nil
	})
}
