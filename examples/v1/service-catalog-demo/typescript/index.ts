import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const nameSuffix = config.get("nameSuffix") ?? "demo";

const projectName = `catalog-demo-${nameSuffix}`;
const envProject = "default";
const envName = `platform-bootstrap-${nameSuffix}`;

// === Owner team (re-impl of v1 Team in v1 namespace) ===
const ownerTeam = new ps.v1.teams.Team("ownerTeam", {
    orgName: organizationName,
    name: `catalog-owner-${nameSuffix}`,
    displayName: "Catalog Demo Owner",
    description: "Owner team for the service catalog demo.",
});

const stacks = {
    checkout: new ps.v1.stacks.Stack("checkoutStack", {
        orgName: organizationName,
        projectName,
        stackName: "checkout-prod",
    }),
    search: new ps.v1.stacks.Stack("searchStack", {
        orgName: organizationName,
        projectName,
        stackName: "search-prod",
    }),
    notification: new ps.v1.stacks.Stack("notificationStack", {
        orgName: organizationName,
        projectName,
        stackName: "notification-prod",
    }),
};

new ps.v1.services.Service("checkoutService", {
    orgName: organizationName,
    name: `checkout-api-${nameSuffix}`,
    description: "Customer-facing checkout REST API. Handles cart submission, payment authorization, and order placement. SLO 99.95% / p99 250ms.",
    ownerType: "team",
    ownerName: ownerTeam.name,
    items: [
        { name: pulumi.interpolate`${projectName}/${stacks.checkout.stackName}`, type: "stack" },
    ],
    properties: [
        { key: "tier", value: "tier-1", type: "string", order: 0 },
        { key: "oncall", value: "payments-pager", type: "string", order: 1 },
        { key: "slack", value: "#payments-oncall", type: "string", order: 2 },
        { key: "runbook", value: "https://runbooks.example.com/checkout-api", type: "string", order: 3 },
    ],
});

new ps.v1.services.Service("searchService", {
    orgName: organizationName,
    name: `search-frontend-${nameSuffix}`,
    description: "Search results UI surface. Owns the search experience across web and mobile clients.",
    ownerType: "team",
    ownerName: ownerTeam.name,
    items: [
        { name: pulumi.interpolate`${projectName}/${stacks.search.stackName}`, type: "stack" },
    ],
    properties: [
        { key: "tier", value: "tier-2", type: "string", order: 0 },
        { key: "oncall", value: "discovery-pager", type: "string", order: 1 },
        { key: "slack", value: "#discovery-eng", type: "string", order: 2 },
        { key: "runbook", value: "https://runbooks.example.com/search", type: "string", order: 3 },
    ],
});

new ps.v1.services.Service("notificationService", {
    orgName: organizationName,
    name: `notification-worker-${nameSuffix}`,
    description: "Async fan-out worker for email, SMS, and push notifications. Consumes events from the platform event bus.",
    ownerType: "team",
    ownerName: ownerTeam.name,
    items: [
        { name: pulumi.interpolate`${projectName}/${stacks.notification.stackName}`, type: "stack" },
    ],
    properties: [
        { key: "tier", value: "tier-2", type: "string", order: 0 },
        { key: "oncall", value: "platform-pager", type: "string", order: 1 },
        { key: "slack", value: "#platform-oncall", type: "string", order: 2 },
    ],
});

const templates = new ps.v1.OrgTemplateCollection("templates", {
    orgName: organizationName,
    name: `platform-templates-${nameSuffix}`,
    sourceURL: "https://github.com/pulumi/templates.git",
});

const webhook = new ps.v1.OrganizationWebhook("webhook", {
    organizationName: organizationName,
    name: `bootstrap-webhook-${nameSuffix}`,
    displayName: "Platform bootstrap webhook",
    payloadUrl: "https://example.com/pulumi-webhook",
    active: true,
    format: "raw",
});

const env = new ps.v1.esc.Environment("bootstrapEnv", {
    orgName: organizationName,
    project: envProject,
    name: envName,
    yaml: [
        "values:",
        "  bootstrap:",
        "    appVersion: 1.0.0",
        "    region: us-west-2",
    ].join("\n") + "\n",
});

const envSchedule = new ps.v1.esc.EnvironmentSchedule("envSchedule", {
    orgName: organizationName,
    projectName: envProject,
    envName: envName,
    scheduleCron: "0 9 * * *",
    secretRotationRequest: { environmentPath: "" },
}, { dependsOn: [env] });

// const customRoleID = "";

// const customRole = new ps.v1.Role("customRole", {
//     orgName: organizationName,
//     roleID: customRoleID,
// }, { import: `${organizationName}/${customRoleID}` });

// new ps.v1.teams.Role("customRoleAssignment", {
//     orgName: organizationName,
//     teamName: ownerTeam.name,
//     roleID: customRoleID,
// }, {
//     import: `${organizationName}/catalog-owner-${nameSuffix}/${customRoleID}`,
//     dependsOn: [customRole],
// });

export const ownerTeamName = ownerTeam.name;
export const templateCollectionName = templates.name;
export const webhookName = webhook.name;
export const bootstrapEnvName = env.id;
export const envScheduleCron = envSchedule.scheduleCron;
