import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const groupName = config.get("groupName") ?? "example-attachment-group";
const projectName = config.get("projectName") ?? "pulumi-service-attachment-example";
const stackName = config.get("stackName") ?? "dev";

const stack = new ps.api.stacks.Stack("exampleStack", {
    orgName: organizationName,
    projectName: projectName,
    stackName: stackName,
});

const group = new ps.api.PolicyGroup("group", {
    orgName: organizationName,
    name: groupName,
    entityType: "stacks",
});

const attachment = new ps.api.PolicyGroupStackAttachment("attachment", {
    orgName: organizationName,
    policyGroup: group.name,
    name: stack.stackName,
    routingProject: projectName,
}, {
    dependsOn: [group, stack],
});

export const attachedStack = attachment.name;
