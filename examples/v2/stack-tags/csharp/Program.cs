using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "v2-stack-tags-example";
    var stackName = config.Get("stackName") ?? "dev";
    var tagValue = config.Get("tagValue") ?? "v2-tag-value";

    var parentStack = new Ps.V2.Stacks.Stack("parentStack", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
    });

    new Ps.V2.Stacks.Tag("ownerTag", new()
    {
        OrgName = organizationName,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Name = "owner",
        Value = "pulumicloud-v2-example",
    });

    new Ps.V2.Stacks.Tag("customTag", new()
    {
        OrgName = organizationName,
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
