package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		org := conf.Require("organizationName")
		name := "test-smoke-go-" + conf.Require("digits")

		environment, err := pulumiservice.NewEnvironment(ctx, "environment", &pulumiservice.EnvironmentArgs{
			Organization: pulumi.String(org),
			Project:      pulumi.String("test-smoke"),
			Name:         pulumi.String(name),
			Yaml:         pulumi.NewStringAsset("values:\n  greeting: hello\n"),
		})
		if err != nil {
			return err
		}

		token, err := pulumiservice.NewOrgAccessToken(ctx, "token", &pulumiservice.OrgAccessTokenArgs{
			OrganizationName: pulumi.String(org),
			Name:             pulumi.String(name),
			Description:      pulumi.String("Created by the go-smoke example."),
			Admin:            pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		lookedUp := pulumiservice.LookupEnvironmentOutput(ctx, pulumiservice.LookupEnvironmentOutputArgs{
			OrganizationName: environment.Organization,
			ProjectName:      environment.Project,
			Name:             environment.Name,
		})

		ctx.Export("environmentName", lookedUp.Name())
		ctx.Export("tokenValue", token.Value)
		return nil
	})
}
