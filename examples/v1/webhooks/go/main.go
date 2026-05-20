package main

import (
	v1 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1"
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
		secretValue := cfg.Get("secretValue")
		if secretValue == "" {
			secretValue = "shhh"
		}
		hookSuffix := cfg.Get("hookSuffix")
		if hookSuffix == "" {
			hookSuffix = "dev"
		}

		orgAll, err := v1.NewOrganizationWebhook(ctx, "orgWebhookAll", &v1.OrganizationWebhookArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("org-webhook-all-" + hookSuffix),
			DisplayName:      pulumi.String("webhook-from-provider"),
			PayloadUrl:       pulumi.String("https://google.com"),
			Active:           pulumi.Bool(true),
			Secret:           pulumi.String(secretValue),
		})
		if err != nil {
			return err
		}

		orgGroups, err := v1.NewOrganizationWebhook(ctx, "orgWebhookGroups", &v1.OrganizationWebhookArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("org-webhook-groups-" + hookSuffix),
			DisplayName:      pulumi.String("webhook-from-provider"),
			PayloadUrl:       pulumi.String("https://google.com"),
			Active:           pulumi.Bool(true),
			Groups:           pulumi.StringArray{pulumi.String("environments"), pulumi.String("stacks")},
			Secret:           pulumi.String(secretValue),
		})
		if err != nil {
			return err
		}

		ctx.Export("orgWebhookId", orgAll.ID())
		ctx.Export("orgWebhookGroupsId", orgGroups.ID())
		return nil
	})
}
