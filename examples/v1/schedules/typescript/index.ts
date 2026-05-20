import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "pulumi-service-schedules-example";
const stackName = config.get("stackName") ?? "dev";
const scheduleCron = config.get("scheduleCron") ?? "0 7 * * *";

const parentStack = new ps.v1.stacks.Stack("parentStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
});

const parentSettings = new ps.v1.deployments.Settings("parentSettings", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
    sourceContext: {
        git: { repoUrl: "https://github.com/example/example.git", branch: "refs/heads/main" },
    },
}, { dependsOn: [parentStack] });

const nightlyDeploy = new ps.v1.deployments.ScheduledDeployment("nightlyDeploy", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
    scheduleCron: scheduleCron,
    request: { operation: "update" },
}, { dependsOn: [parentSettings] });

export const nightlyCron = nightlyDeploy.scheduleCron;
