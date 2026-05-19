import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "v2-stack-tags-example";
const stackName = config.get("stackName") ?? "dev";
const tagValue = config.get("tagValue") ?? "v2-tag-value";

const parentStack = new ps.v2.stacks.Stack("parentStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
});

new ps.v2.stacks.Tag("ownerTag", {
    orgName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "owner",
    value: "pulumicloud-v2-example",
});

new ps.v2.stacks.Tag("customTag", {
    orgName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "purpose",
    value: tagValue,
});

export const parent = parentStack.id;
