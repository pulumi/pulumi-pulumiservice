import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "pulumi-service-stack-example";
const stackName = config.get("stackName") ?? "dev";
const stackPurpose = config.get("stackPurpose") ?? "demo";

const exampleStack = new ps.v2.stacks.Stack("exampleStack", {
    orgName: serviceOrg,
    projectName: projectName,
    stackName: stackName,
    tags: {
        owner: "pulumicloud-v2-example",
        purpose: stackPurpose,
    },
});

export const stack = exampleStack.stackName;
