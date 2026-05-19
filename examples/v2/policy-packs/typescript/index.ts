import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";

const pack = new ps.v2.PolicyPack("pack", {
    orgName: organizationName,
    name: "v2-example-policy-pack",
    displayName: "v2 example policy pack",
    description: "Demo policy pack created via v2 metadata-driven provider.",
    policies: [{
        name: "no-public-buckets",
        description: "Reject S3 buckets with public ACLs",
        enforcementLevel: "advisory",
    }],
});

export const policyPackName = pack.name;
export const policyPackVersion = pack.version;
