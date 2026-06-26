import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const serviceSuffix = config.get("serviceSuffix") ?? "dev";

new ps.api.services.Service("catalogService", {
    orgName: organizationName,
    name: `api-service-${serviceSuffix}`,
    description: "An example api service catalog entry.",
    ownerType: "team",
    ownerName: "platform",
    items: [
        { type: "stack", name: "service-provider-test-org/example-app/dev" },
    ],
    properties: [
        { key: "tier", value: "gold", type: "string", order: 1 },
        { key: "oncall", value: "platform-ops", type: "string", order: 2 },
    ],
});
