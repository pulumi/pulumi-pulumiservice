import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const nameSuffix = config.get("nameSuffix") ?? "demo";

const projectName = `catalog-demo-${nameSuffix}`;
const envProject = "default";
const envName = `platform-bootstrap-${nameSuffix}`;

// === Owner team (re-impl of v1 Team in v2 namespace) ===
const ownerTeam = new ps.v2.teams.Team("ownerTeam", {
    orgName: serviceOrg,
    name: `catalog-owner-${nameSuffix}`,
    displayName: "Catalog Demo Owner",
    description: "Owner team for the service catalog demo.",
});

const stacks = {
    checkout: new ps.v2.stacks.Stack("checkoutStack", {
        orgName: serviceOrg,
        projectName,
        stackName: "checkout-prod",
    }),
    search: new ps.v2.stacks.Stack("searchStack", {
        orgName: serviceOrg,
        projectName,
        stackName: "search-prod",
    }),
    notification: new ps.v2.stacks.Stack("notificationStack", {
        orgName: serviceOrg,
        projectName,
        stackName: "notification-prod",
    }),
};

new ps.v2.services.Service("checkoutService", {
    orgName: serviceOrg,
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

new ps.v2.services.Service("searchService", {
    orgName: serviceOrg,
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

new ps.v2.services.Service("notificationService", {
    orgName: serviceOrg,
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

// === Curated template gallery (NEW in v2) ===
// Visible to every org member in the "New Project" UI immediately.
const templates = new ps.v2.OrgTemplateCollection("templates", {
    orgName: serviceOrg,
    name: `platform-templates-${nameSuffix}`,
    sourceURL: "https://github.com/pulumi/templates.git",
});

// === Org-wide notification webhook (NEW in v2) ===
// v1's `Webhook` was monolithic; v2 splits webhooks into typed,
// scope-specific resources. Placeholder URL — the API doesn't ping it.
const webhook = new ps.v2.OrganizationWebhook("webhook", {
    organizationName: serviceOrg,
    name: `bootstrap-webhook-${nameSuffix}`,
    displayName: "Platform bootstrap webhook",
    payloadUrl: "https://example.com/pulumi-webhook",
    active: true,
    format: "raw",
});

// === ESC environment + schedule + revision tag ===
// `yaml` is a synthesized input surfaced by the v2 dispatch when an op
// declares an application/x-yaml request body. On Create, the dispatch
// fires the JSON CreateEnvironment op then PATCH-es the yaml via Update.
const env = new ps.v2.esc.Environment("bootstrapEnv", {
    orgName: serviceOrg,
    project: envProject,
    name: envName,
    yaml: [
        "values:",
        "  bootstrap:",
        "    appVersion: 1.0.0",
        "    region: us-west-2",
    ].join("\n") + "\n",
});

// Daily secret rotation schedule. Despite the generic name, v2's
// EnvironmentSchedule is a secret-rotation schedule — it requires
// `secretRotationRequest`. Empty `environmentPath` means rotate every
// rotated secret in the env (none here, but the schedule still installs).
// envName is the literal string (v2 esc Environment surfaces no `name`
// output); `dependsOn` preserves the create-before-schedule ordering.
const envSchedule = new ps.v2.esc.EnvironmentSchedule("envSchedule", {
    orgName: serviceOrg,
    projectName: envProject,
    envName: envName,
    scheduleCron: "0 9 * * *",
    secretRotationRequest: { environmentPath: "" },
}, { dependsOn: [env] });

// const customRoleID = "";

// const customRole = new ps.v2.Role("customRole", {
//     orgName: serviceOrg,
//     roleID: customRoleID,
// }, { import: `${serviceOrg}/${customRoleID}` });

// new ps.v2.teams.Role("customRoleAssignment", {
//     orgName: serviceOrg,
//     teamName: ownerTeam.name,
//     roleID: customRoleID,
// }, {
//     import: `${serviceOrg}/catalog-owner-${nameSuffix}/${customRoleID}`,
//     dependsOn: [customRole],
// });

export const ownerTeamName = ownerTeam.name;
export const templateCollectionName = templates.name;
export const webhookName = webhook.name;
export const bootstrapEnvName = env.id;
export const envScheduleCron = envSchedule.scheduleCron;
