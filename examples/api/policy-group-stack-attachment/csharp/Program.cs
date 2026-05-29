using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var groupName = config.Get("groupName") ?? "example-attachment-group";
    var projectName = config.Get("projectName") ?? "pulumi-service-attachment-example";
    var stackName = config.Get("stackName") ?? "dev";

    var exampleStack = new Ps.Api.Stacks.Stack("exampleStack", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
    });

    var group = new Ps.Api.PolicyGroup("group", new()
    {
        OrgName = organizationName,
        Name = groupName,
        EntityType = "stacks",
    });

    var attachment = new Ps.Api.PolicyGroupStackAttachment("attachment", new()
    {
        OrgName = organizationName,
        PolicyGroup = group.Name,
        Name = exampleStack.StackName,
        RoutingProject = projectName,
    }, new CustomResourceOptions
    {
        DependsOn =
        {
            group,
            exampleStack,
        },
    });

    return new Dictionary<string, object?>
    {
        ["attachedStack"] = attachment.Name,
    };
});
