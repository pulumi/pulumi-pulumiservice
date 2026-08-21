import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const policyId = config.get("policyId") ?? "org";

new ps.api.auth.Policy("policy", {
    orgName: organizationName,
    policyId: policyId,
    policies: [
        { decision: "allow", tokenType: "organization", authorizedPermissions: ["standard"], rules: {} },
        { decision: "deny", tokenType: "organization", authorizedPermissions: ["admin"], rules: {} },
    ],
});
