import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const groupName = config.get("groupName") ?? "example-policy-group";

const group = new ps.v2.PolicyGroup("group", {
    orgName: organizationName,
    name: groupName,
    entityType: "stacks",
});

export const policyGroupName = group.name;
