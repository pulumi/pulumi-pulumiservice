using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "v1-stack-config-example";
    var stackName = config.Get("stackName") ?? "dev";
    var hookUrl = config.Get("hookUrl") ?? "https://example.invalid/hooks/example";
    var envRef = config.Get("envRef") ?? "organization/credentials";

    var parentStack = new Ps.V1.Stacks.Stack("parentStack", new()
    {
        OrgName = organizationName,
        ProjectName = projectName,
        StackName = stackName,
    });

    new Ps.V1.Stacks.Config("config", new()
    {
        OrgName = organizationName,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Environment = envRef,
    });

    new Ps.V1.Stacks.Webhook("hook", new()
    {
        OrganizationName = organizationName,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Name = "v1-stackhook",
        DisplayName = "Stack hook example",
        PayloadUrl = hookUrl,
        Active = true,
        Format = "pulumi",
    });

    return new Dictionary<string, object?>
    {
        ["stack"] = parentStack.Id,
    };
});
