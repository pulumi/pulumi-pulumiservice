import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "my-new-project";
const stackName = config.get("stackName") ?? "dev";
const executorImage = config.get("executorImage") ?? "pulumi-cli";

// DeploymentSettings is a singleton-per-stack — ensure the stack exists first.
const parentStack = new ps.v2.Stack("parentStack", {
    orgName: serviceOrg,
    projectName: projectName,
    stackName: stackName,
});

const settings = new ps.v2.DeploymentSettings("settings", {
    orgName: serviceOrg,
    projectName: projectName,
    stackName: stackName,
    executorContext: { executorImage: executorImage },
    operationContext: {
        preRunCommands: ["yarn"],
        environmentVariables: { TEST_VAR: "foo" },
        options: { skipInstallDependencies: true },
    },
    sourceContext: {
        git: {
            repoUrl: "https://github.com/example/example.git",
            branch: "refs/heads/main",
        },
    },
}, { dependsOn: [parentStack] });

export const stackId = settings.stackName;
