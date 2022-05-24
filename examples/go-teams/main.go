package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		_, err := pulumiservice.NewTeam(ctx, "team", &pulumiservice.TeamArgs{
			Name:             pulumi.String("brand-new-go-team"),
			Description:      pulumi.String("This was created with Pulumi"),
			DisplayName:      pulumi.String("PulumiUP Team"),
			OrganizationName: pulumi.String("service-provider-test-org"),
			TeamType:         pulumi.String("pulumi"),
			Members: pulumi.ToStringArray([]string{
				"pulumi-bot",
				"service-provider-example-user",
			}),
		})
		if err != nil {
			return err
		}
		return nil
	})
}
