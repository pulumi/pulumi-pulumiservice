import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const digits = config.require("digits");
const projectName = `sbc-test-${digits}`;

// An ESC environment that holds config/secrets shared by the linked stack.
const sharedCfg = new service.Environment("sharedCfg", {
    organization: organizationName,
    project: "default",
    name: `sbc-shared-cfg-${digits}`,
    yaml: new pulumi.asset.StringAsset(`values:
  app:
    greeting: "hello from ESC"
`),
});

// Mode 1: reference an existing ESC environment as the stack's remote config.
// Stack lifecycle: PSP creates the stack, links the env, and on delete keeps
// the env because PSP does not own its lifecycle.
const refStack = new service.Stack("refStack", {
    organizationName,
    projectName,
    stackName: `ref-${digits}`,
    configEnvironment: {
        project: sharedCfg.project,
        environment: sharedCfg.name,
        // Pin to the latest revision Pulumi wrote, so subsequent edits to the
        // env do not silently change this stack's config until the program is
        // re-applied.
        version: sharedCfg.revision.apply(revision => revision.toString()),
    },
});

// Mode 2: stack-managed env. The server creates an env named
// `${projectName}/${stackName}` alongside the stack, and deletes it when the
// stack is destroyed.
const autoStack = new service.Stack("autoStack", {
    organizationName,
    projectName,
    stackName: `auto-${digits}`,
    configEnvironment: {
        auto: true,
    },
});

export const refStackName = refStack.stackName;
export const autoStackName = autoStack.stackName;
