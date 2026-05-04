import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "v2-stack-tags-example";
const stackName = config.get("stackName") ?? "dev";
const tagValue = config.get("tagValue") ?? "v2-tag-value";

const parentStack = new ps.v2.Stack("parentStack", {
    orgName: serviceOrg,
    projectName: projectName,
    stackName: stackName,
});

new ps.v2.StackTag("ownerTag", {
    orgName: serviceOrg,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "owner",
    value: "pulumicloud-v2-example",
});

new ps.v2.StackTag("customTag", {
    orgName: serviceOrg,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "purpose",
    value: tagValue,
});

export const parent = parentStack.id;
