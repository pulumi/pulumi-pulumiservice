package main

import (
	insights "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/insights"
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
		accountSuffix := cfg.Get("accountSuffix")
		if accountSuffix == "" {
			accountSuffix = "dev"
		}
		insightsEnv := cfg.Get("insightsEnvironment")
		if insightsEnv == "" {
			insightsEnv = "insights/credentials"
		}

		account, err := insights.NewAccount(ctx, "account", &insights.AccountArgs{
			OrgName:      pulumi.String(serviceOrg),
			AccountName:  pulumi.String("v2-insights-" + accountSuffix),
			Provider:     pulumi.String("aws"),
			Environment:  pulumi.String(insightsEnv),
			ScanSchedule: pulumi.String("none"),
		})
		if err != nil {
			return err
		}

		if _, err := insights.NewScheduledScanSettings(ctx, "scanSettings", &insights.ScheduledScanSettingsArgs{
			OrgName:      pulumi.String(serviceOrg),
			AccountName:  account.AccountName,
			Paused:       pulumi.Bool(true),
			ScheduleCron: pulumi.String("0 6 * * *"),
		}); err != nil {
			return err
		}

		ctx.Export("accountName", account.AccountName)
		return nil
	})
}
