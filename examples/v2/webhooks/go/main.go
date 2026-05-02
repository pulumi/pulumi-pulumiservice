package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		serviceOrg := "service-provider-test-org"
		if param := cfg.Get("serviceOrg"); param != "" {
			serviceOrg = param
		}
		secretValue := "shhh"
		if param := cfg.Get("secretValue"); param != "" {
			secretValue = param
		}
		// Organization-scoped webhook subscribed to all events.
		orgWebhookAll, err := v2.NewOrganizationWebhook(ctx, "orgWebhookAll", &v2.OrganizationWebhookArgs{
			OrgName:          serviceOrg,
			OrganizationName: serviceOrg,
			Name:             "org-webhook-all",
			DisplayName:      "webhook-from-provider",
			PayloadUrl:       "https://google.com",
			Active:           true,
			Secret:           secretValue,
		})
		if err != nil {
			return err
		}
		// Organization-scoped webhook subscribed only to environments and stacks groups.
		_, err = v2.NewOrganizationWebhook(ctx, "orgWebhookGroups", &v2.OrganizationWebhookArgs{
			OrgName:          serviceOrg,
			OrganizationName: serviceOrg,
			Name:             "org-webhook-groups",
			DisplayName:      "webhook-from-provider",
			PayloadUrl:       "https://google.com",
			Active:           true,
			Groups: []string{
				"environments",
				"stacks",
			},
			Secret: secretValue,
		})
		if err != nil {
			return err
		}
		ctx.Export("orgWebhookId", pulumi.Any(orgWebhookAll.Id))
		return nil
	})
}
