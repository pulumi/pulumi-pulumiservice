import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") || "service-provider-test-org";
const secretValue = config.get("secretValue") || "shhh";
// Organization-scoped webhook subscribed to all events.
const orgWebhookAll = new pulumiservice.v2.OrganizationWebhook("orgWebhookAll", {
    orgName: serviceOrg,
    organizationName: serviceOrg,
    name: "org-webhook-all",
    displayName: "webhook-from-provider",
    payloadUrl: "https://google.com",
    active: true,
    secret: secretValue,
});
// Organization-scoped webhook subscribed only to environments and stacks groups.
const orgWebhookGroups = new pulumiservice.v2.OrganizationWebhook("orgWebhookGroups", {
    orgName: serviceOrg,
    organizationName: serviceOrg,
    name: "org-webhook-groups",
    displayName: "webhook-from-provider",
    payloadUrl: "https://google.com",
    active: true,
    groups: [
        "environments",
        "stacks",
    ],
    secret: secretValue,
});
export const orgWebhookId = orgWebhookAll.id;
