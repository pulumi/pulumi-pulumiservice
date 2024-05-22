package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		_, err := pulumiservice.NewEnvironment(ctx, "testing-environment", &pulumiservice.EnvironmentArgs{
			Name:         pulumi.String("testing-environment-go"),
			Organization: pulumi.String("service-provider-test-org"),
			Yaml: pulumi.NewStringAsset(`
            values:
              myKey1: "myValue1"
              myNestedKey:
                myKey2: "myValue2"
                myNumber: 1`),
		})
		if err != nil {
			return err
		}
		return nil
	})
}
