using Pulumi;
using Pulumi.PulumiService;
using System;

class MyStack : Stack
{
    public MyStack()
    {
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
                Name = "testing-environment-cs",
                Yaml = new StringAsset(yaml)
            }
        );
    }
}
