import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const digits = config.require("digits");
const projectName = `sbc-test-${digits}`;

// Stack-managed ESC environment. The server creates an env named
// `${projectName}/${stackName}` alongside the stack, and deletes it when the
// stack is destroyed.
const managedStack = new service.Stack("managedStack", {
    organizationName,
    projectName,
    stackName: `managed-${digits}`,
    configEnvironment: {
        managed: true,
    },
});

export const managedStackName = managedStack.stackName;
