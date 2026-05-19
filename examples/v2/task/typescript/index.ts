import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const taskSuffix = config.get("taskSuffix") ?? "dev";
const taskID = config.get("taskID") ?? "example-task-id";

const pool = new ps.v2.agents.Pool("pool", {
    orgName: organizationName,
    name: `v2-task-pool-${taskSuffix}`,
    description: "Pool used by the v2 task example",
});

new ps.v2.agents.Task("task", {
    orgName: organizationName,
    taskID: taskID,
    approvalMode: "manual",
    permissionMode: "default",
    source: "api",
    planMode: false,
});

export const poolName = pool.name;
