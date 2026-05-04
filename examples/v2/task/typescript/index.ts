import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const taskSuffix = config.get("taskSuffix") ?? "dev";
const taskID = config.get("taskID") ?? "example-task-id";

const pool = new ps.v2.AgentPool("pool", {
    orgName: serviceOrg,
    name: `v2-task-pool-${taskSuffix}`,
    description: "Pool used by the v2 task example",
});

new ps.v2.Task("task", {
    orgName: serviceOrg,
    taskID: taskID,
    name: `v2-task-${taskSuffix}`,
    approvalMode: "manual",
    permissionMode: "maintainer",
    isShared: false,
});

export const poolName = pool.name;
