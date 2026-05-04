import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const groupName = config.get("groupName") ?? "example-policy-group";

const group = new ps.v2.PolicyGroup("group", {
    orgName: serviceOrg,
    name: groupName,
    entityType: "stacks",
});

export const policyGroupName = group.name;
