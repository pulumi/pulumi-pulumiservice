import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "api-stack-config-example";
const stackName = config.get("stackName") ?? "dev";
const hookUrl = config.get("hookUrl") ?? "https://example.invalid/hooks/example";
const envRef = config.get("envRef") ?? "organization/credentials";

const parentStack = new ps.api.stacks.Stack("parentStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
});

new ps.api.stacks.Config("config", {
    orgName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    environment: envRef,
});

new ps.api.stacks.Webhook("hook", {
    organizationName: organizationName,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "api-stackhook",
    displayName: "Stack hook example",
    payloadUrl: hookUrl,
    active: true,
    format: "pulumi",
});

export const stack = parentStack.id;
