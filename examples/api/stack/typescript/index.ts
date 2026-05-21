import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "pulumi-service-stack-example";
const stackName = config.get("stackName") ?? "dev";
const stackPurpose = config.get("stackPurpose") ?? "demo";

const exampleStack = new ps.api.stacks.Stack("exampleStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
    tags: {
        owner: "pulumicloud-api-example",
        purpose: stackPurpose,
    },
});

export const stack = exampleStack.stackName;
