package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		rand, err := random.NewRandomString(ctx, "random", &random.RandomStringArgs{
			Length:          pulumi.Int(5),
			Special:         pulumi.Bool(false),
		})
		if err != nil {
			return err
		}
		_, err = pulumiservice.NewTeam(ctx, "team", &pulumiservice.TeamArgs{
			Name:             pulumi.Sprintf("brand-new-go-team-%s", rand.Result),
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
