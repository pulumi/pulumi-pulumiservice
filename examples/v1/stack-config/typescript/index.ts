import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "v1-stack-config-example";
const stackName = config.get("stackName") ?? "dev";
const hookUrl = config.get("hookUrl") ?? "https://example.invalid/hooks/example";
const envRef = config.get("envRef") ?? "organization/credentials";

const parentStack = new ps.v1.stacks.Stack("parentStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
});

new ps.v1.stacks.Config("config", {
    orgName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    environment: envRef,
});

new ps.v1.stacks.Webhook("hook", {
    organizationName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "v1-stackhook",
    displayName: "Stack hook example",
    payloadUrl: hookUrl,
    active: true,
    format: "pulumi",
});

export const stack = parentStack.id;
