import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const secretValue = config.get("secretValue") ?? "shhh";
const hookSuffix = config.get("hookSuffix") ?? "dev";

const orgWebhookAll = new ps.v1.OrganizationWebhook("orgWebhookAll", {
    organizationName: organizationName,
    name: `org-webhook-all-${hookSuffix}`,
    displayName: "webhook-from-provider",
    payloadUrl: "https://google.com",
    active: true,
    secret: secretValue,
});

const orgWebhookGroups = new ps.v1.OrganizationWebhook("orgWebhookGroups", {
    organizationName: organizationName,
    name: `org-webhook-groups-${hookSuffix}`,
    displayName: "webhook-from-provider",
    payloadUrl: "https://google.com",
    active: true,
    groups: ["environments", "stacks"],
    secret: secretValue,
});

export const orgWebhookId = orgWebhookAll.id;
export const orgWebhookGroupsId = orgWebhookGroups.id;
