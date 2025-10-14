package main

import (
	"strconv"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		yaml := `values:
  myKey1: "myValue1"
  myNestedKey:
    myKey2: "myValue2"
    myNumber: 1`

		environment, err := pulumiservice.NewEnvironment(ctx, "testing-environment", &pulumiservice.EnvironmentArgs{
			Name:         pulumi.String("testing-environment-go-" + conf.Require("digits")),
			Project:      pulumi.String("my-project"),
			Organization: pulumi.String("service-provider-test-org"),
			Yaml:         pulumi.NewStringAsset(yaml),
		})
		if err != nil {
			return err
		}

		// A tag that will always be placed on the latest revision of the environment
		_, err = pulumiservice.NewEnvironmentVersionTag(ctx, "StableTag", &pulumiservice.EnvironmentVersionTagArgs{
			Organization: environment.Organization,
			Project:      environment.Project,
			Environment:  environment.Name,
			TagName:      pulumi.String("stable"),
			Revision:     environment.Revision,
		})
		if err != nil {
			return err
		}

		// A tag that will be placed on each new version, and remain on old revisions
		_, err = pulumiservice.NewEnvironmentVersionTag(ctx, "VersionTag", &pulumiservice.EnvironmentVersionTagArgs{
			Organization: environment.Organization,
			Project:      environment.Project,
			Environment:  environment.Name,
			TagName: environment.Revision.ApplyT(func(rev int) (string, error) {
				return "v" + strconv.Itoa(rev), nil
			}).(pulumi.StringOutput),
			Revision: environment.Revision,
		}, pulumi.RetainOnDelete(true))
		if err != nil {
			return err
		}

		// A team to use for TeamEnvironmentPermission
		team, err := pulumiservice.NewTeam(ctx, "team", &pulumiservice.TeamArgs{
			Name:             pulumi.Sprintf("brand-new-go-team-%s", conf.Require("digits")),
			Description:      pulumi.String("This was created with Pulumi"),
			DisplayName:      pulumi.String("PulumiUP Team"),
			OrganizationName: environment.Organization,
			TeamType:         pulumi.String("pulumi"),
			Members: pulumi.ToStringArray([]string{
				"pulumi-bot",
				"service-provider-example-user",
			}),
		})
		if err != nil {
			return err
		}

		_, err = pulumiservice.NewTeamEnvironmentPermission(ctx, "teamEnvironmentPermission", &pulumiservice.TeamEnvironmentPermissionArgs{
			Organization: environment.Organization,
			Team: team.Name.ApplyT(func(name *string) (string, error) {
				return *name, nil
			}).(pulumi.StringOutput),
			Environment: environment.Name,
			Project:     environment.Project,
			Permission:  pulumiservice.EnvironmentPermissionAdmin,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
