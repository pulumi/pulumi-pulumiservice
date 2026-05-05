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

// === Default org for the demo runner (NEW in v2) ===
// Per-user preference: every user who runs `pulumi up` lands on this org
// when they next open pulumi.com.
const defaultOrg = new ps.v2.DefaultOrganization("defaultOrg", {
    orgName: serviceOrg,
});

const stacks = {
    checkout: new ps.Stack("checkoutStack", {
        organizationName: serviceOrg,
        projectName,
        stackName: "checkout-prod",
        forceDestroy: true,
    }),
    search: new ps.Stack("searchStack", {
        organizationName: serviceOrg,
        projectName,
        stackName: "search-prod",
        forceDestroy: true,
    }),
    notification: new ps.Stack("notificationStack", {
        organizationName: serviceOrg,
        projectName,
        stackName: "notification-prod",
        forceDestroy: true,
    }),
};

const checkoutService = new ps.v2.services.Service("checkoutService", {
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

const searchService = new ps.v2.services.Service("searchService", {
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

const notificationService = new ps.v2.services.Service("notificationService", {
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
// We use the v1 Environment because v2/esc:Environment takes no `yaml`
// input — it would create an empty env with no revisions.
const env = new ps.Environment("bootstrapEnv", {
    organization: serviceOrg,
    project: envProject,
    name: envName,
    yaml: new pulumi.asset.StringAsset(
        [
            "values:",
            "  bootstrap:",
            "    appVersion: 1.0.0",
            "    region: us-west-2",
        ].join("\n") + "\n",
    ),
});

// Daily secret rotation schedule. Despite the generic name, v2's
// EnvironmentSchedule is a secret-rotation schedule — it requires
// `secretRotationRequest`. Empty `environmentPath` means rotate every
// rotated secret in the env (none here, but the schedule still installs).
const envSchedule = new ps.v2.esc.EnvironmentSchedule("envSchedule", {
    orgName: serviceOrg,
    projectName: envProject,
    envName: env.name,
    scheduleCron: "0 9 * * *",
    secretRotationRequest: { environmentPath: "" },
});

// NOTE: `v2/esc:RevisionTag` was tried but its v2 metadata maps Create →
// UpdateRevisionTag (PATCH /tags/{name}), which returns 404 on a new tag.
// Provider bug — needs Create → POST. Dropped pending a fix.

export const ownerTeamName = ownerTeam.name;
export const defaultOrgName = defaultOrg.orgName;
export const checkoutServiceName = checkoutService.name;
export const searchServiceName = searchService.name;
export const notificationServiceName = notificationService.name;
export const templateCollectionName = templates.name;
export const webhookName = webhook.name;
export const bootstrapEnvName = env.name;
export const envScheduleCron = envSchedule.scheduleCron;
