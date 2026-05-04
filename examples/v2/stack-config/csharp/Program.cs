using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var projectName = config.Get("projectName") ?? "v2-stack-config-example";
    var stackName = config.Get("stackName") ?? "dev";
    var hookUrl = config.Get("hookUrl") ?? "https://example.invalid/hooks/example";
    var envRef = config.Get("envRef") ?? "organization/credentials";

    var parentStack = new Ps.V2.Stack("parentStack", new()
    {
        OrgName = serviceOrg,
        ProjectName = projectName,
        StackName = stackName,
    });

    new Ps.V2.StackConfig("config", new()
    {
        OrgName = serviceOrg,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Environment = envRef,
    });

    new Ps.V2.StackWebhook("hook", new()
    {
        OrganizationName = serviceOrg,
        ProjectName = parentStack.ProjectName,
        StackName = parentStack.StackName,
        Name = "v2-stackhook",
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
