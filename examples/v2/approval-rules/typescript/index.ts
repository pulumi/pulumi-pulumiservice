import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";

const approvers = new ps.v2.PolicyGroup("approvers", {
    orgName: organizationName,
    name: "v2-approvers",
    entityType: "stacks",
});

export const policyGroupName = approvers.name;
