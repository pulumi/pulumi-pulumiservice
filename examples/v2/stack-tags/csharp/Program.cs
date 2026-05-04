using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "v2-stack-tags-example";
    var stackName = config.Get("stackName") ?? "dev";
    var tagValue = config.Get("tagValue") ?? "v2-tag-value";

    var parentStack = new Ps.V2.Stack("parentStack", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        StackName = stackName,
    });

    new Ps.V2.StackTag("ownerTag", new()
    {
        OrgName = serviceOrg,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Name = "owner",
        Value = "pulumicloud-v2-example",
    });

    new Ps.V2.StackTag("customTag", new()
    {
        OrgName = serviceOrg,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Name = "purpose",
        Value = tagValue,
    });

    return new Dictionary<string, object?>
    {
        ["parent"] = parentStack.Id,
    };
});
