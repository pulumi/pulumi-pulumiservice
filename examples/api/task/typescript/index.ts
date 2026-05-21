import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const taskSuffix = config.get("taskSuffix") ?? "dev";
const taskID = config.get("taskID") ?? "example-task-id";

const pool = new ps.api.agents.Pool("pool", {
    orgName: organizationName,
    name: `api-task-pool-${taskSuffix}`,
    description: "Pool used by the api task example",
});

new ps.api.agents.Task("task", {
    orgName: organizationName,
    taskID: taskID,
    approvalMode: "manual",
    permissionMode: "default",
    source: "api",
    planMode: false,
});

export const poolName = pool.name;
