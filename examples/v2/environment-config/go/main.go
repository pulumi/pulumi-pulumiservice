package main

import (
	esc "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/esc"
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
			projectName = "v2-envcfg-example"
		}
		envName := cfg.Get("envName")
		if envName == "" {
			envName = "v2-envcfg-env"
		}

		draft, err := esc.NewEnvironmentDraft(ctx, "draft", &esc.EnvironmentDraftArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: pulumi.String(projectName),
			EnvName:     pulumi.String(envName),
		})
		if err != nil {
			return err
		}

		settings, err := esc.NewEnvironmentSettings(ctx, "settings", &esc.EnvironmentSettingsArgs{
			OrgName:           pulumi.String(serviceOrg),
			ProjectName:       pulumi.String(projectName),
			EnvName:           pulumi.String(envName),
			DeletionProtected: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		ctx.Export("draftId", draft.ChangeRequestID)
		ctx.Export("protected", settings.DeletionProtected)
		return nil
	})
}
