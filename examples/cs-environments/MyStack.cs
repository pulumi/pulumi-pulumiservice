using Pulumi;
using Pulumi.PulumiService;
using System;

class MyStack : Stack
{
    public MyStack()
    {
        var config = new Pulumi.Config();
        String yaml = """
        values:
          myKey1: "myValue1"
          myNestedKey:
            myKey2: "myValue2"
            myNumber: 1
        """;

        var environment = new Pulumi.PulumiService.Environment(
            "testing-environment",
            new EnvironmentArgs {
                Organization = "service-provider-test-org",
                Name = "testing-environment-cs-" + config.Require("digits"),
                Yaml = new StringAsset(yaml)
            }
        );

        // A tag that will always be placed on the latest revision of the environment
        var stableTag = new Pulumi.PulumiService.EnvironmentVersionTag(
            "StableTag",
            new EnvironmentVersionTagArgs {
                Organization = environment.Organization,
                Environment = environment.Name,
                TagName = "stable",
                Revision = environment.Revision
            }
        );

        // A tag that will be placed on each new version, and remain on old revisions
        var versionTag = new Pulumi.PulumiService.EnvironmentVersionTag(
            "VersionTag",
            new EnvironmentVersionTagArgs {
                Organization = environment.Organization,
                Environment = environment.Name,
                TagName = environment.Revision.Apply(rev => "v"+rev),
                Revision = environment.Revision
            },
            new CustomResourceOptions{
                RetainOnDelete = true
            }
        );
    }
}
