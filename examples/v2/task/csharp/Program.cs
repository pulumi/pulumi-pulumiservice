using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var taskSuffix = config.Get("taskSuffix") ?? "dev";
    var taskID = config.Get("taskID") ?? "example-task-id";

    var pool = new Ps.V2.Agents.Pool("pool", new()
    {
        OrgName = organizationName,
        Name = $"v2-task-pool-{taskSuffix}",
        Description = "Pool used by the v2 task example",
    });

    new Ps.V2.Agents.Task("task", new()
    {
        OrgName = organizationName,
        TaskID = taskID,
        ApprovalMode = "manual",
        PermissionMode = "default",
        Source = "api",
        PlanMode = false,
    });

    return new Dictionary<string, object?>
    {
        ["poolName"] = pool.Name,
    };
});
