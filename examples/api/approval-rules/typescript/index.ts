import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const groupName = config.get("groupName") ?? "api-approvers";

const approvers = new ps.api.PolicyGroup("approvers", {
    orgName: organizationName,
    name: groupName,
    entityType: "stacks",
});

export const policyGroupName = approvers.name;
