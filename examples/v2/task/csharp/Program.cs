using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var taskSuffix = config.Get("taskSuffix") ?? "dev";
    var taskID = config.Get("taskID") ?? "example-task-id";

    var pool = new Ps.V2.Agents.Pool("pool", new()
    {
        OrgName = serviceOrg,
        Name = $"v2-task-pool-{taskSuffix}",
        Description = "Pool used by the v2 task example",
    });

    new Ps.V2.Agents.Task("task", new()
    {
        OrgName = serviceOrg,
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
