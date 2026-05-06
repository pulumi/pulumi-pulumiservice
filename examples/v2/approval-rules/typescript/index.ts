import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";

const approvers = new ps.v2.PolicyGroup("approvers", {
    orgName: serviceOrg,
    name: "v2-approvers",
    entityType: "stacks",
});

export const policyGroupName = approvers.name;
