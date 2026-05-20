using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "pulumi-service-stack-example";
    var stackName = config.Get("stackName") ?? "dev";
    var stackPurpose = config.Get("stackPurpose") ?? "demo";

    var exampleStack = new Ps.V1.Stacks.Stack("exampleStack", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
        Tags =
        {
            { "owner", "pulumicloud-v1-example" },
            { "purpose", stackPurpose },
        },
    });

    return new Dictionary<string, object?>
    {
        ["stackName"] = exampleStack.StackName,
    };
});
