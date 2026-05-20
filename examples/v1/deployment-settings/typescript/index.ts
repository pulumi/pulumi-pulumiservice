import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "my-new-project";
const stackName = config.get("stackName") ?? "dev";
const executorImage = config.get("executorImage") ?? "pulumi-cli";

// DeploymentSettings is a singleton-per-stack — ensure the stack exists first.
const parentStack = new ps.v1.stacks.Stack("parentStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
});

const settings = new ps.v1.deployments.Settings("settings", {
    orgName: organizationName,
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

// stackName is a path-param input; reference the source value rather than
// the resource (deployments:Settings doesn't surface stackName on state).
export const stackId = pulumi.interpolate`${organizationName}/${projectName}/${stackName}`;
