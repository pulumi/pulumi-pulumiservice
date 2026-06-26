import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const policyId = config.get("policyId") ?? "org";

new ps.api.auth.Policy("policy", {
    orgName: organizationName,
    policyId: policyId,
    policies: [
        { decision: "allow", authorizedPermissions: ["read"], tokenType: "organization", rules: {} },
        { decision: "deny", authorizedPermissions: ["admin"], tokenType: "organization", rules: {} },
    ],
});
