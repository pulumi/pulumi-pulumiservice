// TypeScript variant of canonical/07-tiered-team-access.
// Three teams × two stacks × three permission tiers. Behavioral twin of
// the sibling YAML program.

import * as pulumi from "@pulumi/pulumi";
import * as teams from "@pulumi/pulumiservice/orgs/teams";
import * as permissions from "@pulumi/pulumiservice/stacks/permissions";

const cfg = new pulumi.Config();
const organizationName = cfg.get("organizationName") ?? "service-provider-test-org";
const digits = cfg.get("digits") ?? "00000";

// Team-stack permission levels (from TeamStackPermission): 0=none,
// 101=read, 102=edit, 103=admin.
const permRead = 101;
const permEdit = 102;
const permAdmin = 103;

const platformAdmins = new teams.Team("platformAdmins", {
    organizationName,
    teamType: "pulumi",
    name: `platform-admins-${digits}`,
    displayName: "Platform Admins",
    description: "Break-glass access to everything. Keep small.",
});

const billingOwners = new teams.Team("billingOwners", {
    organizationName,
    teamType: "pulumi",
    name: `billing-owners-${digits}`,
    displayName: "Billing Service Owners",
    description: "Owns the billing service stacks end-to-end.",
});

const developers = new teams.Team("developers", {
    organizationName,
    teamType: "pulumi",
    name: `developers-${digits}`,
    displayName: "Developers (all)",
    description: "Read everything; deploy nothing without an explicit grant.",
});

// Platform stack
new permissions.TeamStackPermission("platformAdminPerm", {
    organization: organizationName, project: "platform", stack: "prod",
    team: platformAdmins.name.apply(n => n!),
    permission: permAdmin,
});
new permissions.TeamStackPermission("platformDevRead", {
    organization: organizationName, project: "platform", stack: "prod",
    team: developers.name.apply(n => n!),
    permission: permRead,
});

// Billing stack
new permissions.TeamStackPermission("billingAdminPerm", {
    organization: organizationName, project: "billing-service", stack: "prod",
    team: platformAdmins.name.apply(n => n!),
    permission: permAdmin,
});
new permissions.TeamStackPermission("billingOwnerPerm", {
    organization: organizationName, project: "billing-service", stack: "prod",
    team: billingOwners.name.apply(n => n!),
    permission: permEdit,
});
new permissions.TeamStackPermission("billingDevRead", {
    organization: organizationName, project: "billing-service", stack: "prod",
    team: developers.name.apply(n => n!),
    permission: permRead,
});
