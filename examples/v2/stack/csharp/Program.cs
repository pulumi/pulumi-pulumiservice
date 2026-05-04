using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "pulumi-service-stack-example";
    var stackName = config.Get("stackName") ?? "dev";
    var stackPurpose = config.Get("stackPurpose") ?? "demo";

    var exampleStack = new Ps.V2.Stack("exampleStack", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        StackName = stackName,
        Tags =
        {
            { "owner", "pulumicloud-v2-example" },
            { "purpose", stackPurpose },
        },
    });

    return new Dictionary<string, object?>
    {
        ["stackName"] = exampleStack.StackName,
    };
});
