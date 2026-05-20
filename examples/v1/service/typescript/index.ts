import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const serviceSuffix = config.get("serviceSuffix") ?? "dev";

new ps.v1.services.Service("catalogService", {
    orgName: organizationName,
    name: `v1-service-${serviceSuffix}`,
    description: "An example v1 service catalog entry.",
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
