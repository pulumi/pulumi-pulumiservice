import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "api-stack-tags-example";
const stackName = config.get("stackName") ?? "dev";
const tagValue = config.get("tagValue") ?? "api-tag-value";

const parentStack = new ps.api.stacks.Stack("parentStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
});

new ps.api.stacks.Tag("ownerTag", {
    orgName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "owner",
    value: "pulumicloud-api-example",
});

new ps.api.stacks.Tag("customTag", {
    orgName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "purpose",
    value: tagValue,
});

export const parent = parentStack.id;
