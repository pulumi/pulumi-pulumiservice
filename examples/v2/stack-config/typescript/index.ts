import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "v2-stack-config-example";
const stackName = config.get("stackName") ?? "dev";
const hookUrl = config.get("hookUrl") ?? "https://example.invalid/hooks/example";
const envRef = config.get("envRef") ?? "organization/credentials";

const parentStack = new ps.v2.stacks.Stack("parentStack", {
    orgName: serviceOrg,
    projectName: projectName,
    stackName: stackName,
});

new ps.v2.stacks.Config("config", {
    orgName: serviceOrg,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    environment: envRef,
});

new ps.v2.stacks.Webhook("hook", {
    organizationName: serviceOrg,
    projectName: parentStack.projectName,
    stackName: parentStack.stackName,
    name: "v2-stackhook",
    displayName: "Stack hook example",
    payloadUrl: hookUrl,
    active: true,
    format: "pulumi",
});

export const stack = parentStack.id;
