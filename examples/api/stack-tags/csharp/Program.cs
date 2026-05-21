using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "api-stack-tags-example";
    var stackName = config.Get("stackName") ?? "dev";
    var tagValue = config.Get("tagValue") ?? "api-tag-value";

    var parentStack = new Ps.Api.Stacks.Stack("parentStack", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
    });

    new Ps.Api.Stacks.Tag("ownerTag", new()
    {
        OrgName = organizationName,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Name = "owner",
        Value = "pulumicloud-api-example",
    });

    new Ps.Api.Stacks.Tag("customTag", new()
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
