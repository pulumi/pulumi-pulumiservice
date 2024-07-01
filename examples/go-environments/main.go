package main

import (
	"strconv"

	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
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
			Organization: pulumi.String("service-provider-test-org"),
			Yaml:         pulumi.NewStringAsset(yaml),
		})
		if err != nil {
			return err
		}

		// A tag that will always be placed on the latest revision of the environment
		_, err = pulumiservice.NewEnvironmentVersionTag(ctx, "StableTag", &pulumiservice.EnvironmentVersionTagArgs{
			Organization: environment.Organization,
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
			Environment:  environment.Name,
			TagName: environment.Revision.ApplyT(func(rev int) (string, error) {
				return "v" + strconv.Itoa(rev), nil
			}).(pulumi.StringOutput),
			Revision: environment.Revision,
		}, pulumi.RetainOnDelete(true))
		if err != nil {
			return err
		}

		return nil
	})
}
