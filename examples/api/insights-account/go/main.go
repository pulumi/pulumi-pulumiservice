package main

import (
	insights "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/insights"
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
		accountSuffix := cfg.Get("accountSuffix")
		if accountSuffix == "" {
			accountSuffix = "dev"
		}
		insightsEnv := cfg.Get("insightsEnvironment")
		if insightsEnv == "" {
			insightsEnv = "insights/credentials"
		}

		accountNameValue := "api-insights-" + accountSuffix
		account, err := insights.NewAccount(ctx, "account", &insights.AccountArgs{
			OrgName:      pulumi.String(organizationName),
			AccountName:  pulumi.String(accountNameValue),
			Provider:     pulumi.String("aws"),
			Environment:  pulumi.String(insightsEnv),
			ScanSchedule: pulumi.String("none"),
		})
		if err != nil {
			return err
		}

		if _, err := insights.NewScheduledScanSettings(ctx, "scanSettings", &insights.ScheduledScanSettingsArgs{
			OrgName:      pulumi.String(organizationName),
			AccountName:  pulumi.String(accountNameValue),
			Paused:       pulumi.Bool(true),
			ScheduleCron: pulumi.String("0 6 * * *"),
		}, pulumi.DependsOn([]pulumi.Resource{account})); err != nil {
			return err
		}

		ctx.Export("accountName", account.Name)
		return nil
	})
}
