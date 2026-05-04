import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const policyId = config.get("policyId") ?? "org";

new ps.v2.AuthPolicy("policy", {
    orgName: serviceOrg,
    policyId: policyId,
    policies: [
        { decision: "allow", permission: "read", tokenType: "organization" },
        { decision: "deny", permission: "admin", tokenType: "organization" },
    ],
});
