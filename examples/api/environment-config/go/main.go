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
			projectName = "api-envcfg-example"
		}
		envName := cfg.Get("envName")
		if envName == "" {
			envName = "api-envcfg-env"
		}

		draft, err := esc.NewEnvironmentDraft(ctx, "draft", &esc.EnvironmentDraftArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String(projectName),
			EnvName:     pulumi.String(envName),
		})
		if err != nil {
			return err
		}

		settings, err := esc.NewEnvironmentSettings(ctx, "settings", &esc.EnvironmentSettingsArgs{
			OrgName:           pulumi.String(organizationName),
			ProjectName:       pulumi.String(projectName),
			EnvName:           pulumi.String(envName),
			DeletionProtected: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		ctx.Export("draftId", draft.ChangeRequestId)
		ctx.Export("protected", settings.DeletionProtected)
		return nil
	})
}
