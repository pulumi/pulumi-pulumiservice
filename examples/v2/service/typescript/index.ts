import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const serviceSuffix = config.get("serviceSuffix") ?? "dev";

new ps.v2.Service("catalogService", {
    orgName: serviceOrg,
    name: `v2-service-${serviceSuffix}`,
    description: "An example v2 service catalog entry.",
    ownerType: "team",
    ownerName: "platform",
    items: [
        { kind: "stack", ref: "service-provider-test-org/example-app/dev" },
    ],
    properties: [
        { key: "tier", value: "gold" },
        { key: "oncall", value: "platform-ops" },
    ],
});
