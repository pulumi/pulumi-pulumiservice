// TypeScript variant of canonical/01-organization-bootstrap.
// Day-0 provisioning: three teams, a team-scoped CI token, a baseline
// policy group. Functionally equivalent to the sibling YAML program.

import * as pulumi from "@pulumi/pulumi";
import * as teams from "@pulumi/pulumiservice/orgs/teams";
import * as tokens from "@pulumi/pulumiservice/orgs/tokens";
import * as policies from "@pulumi/pulumiservice/orgs/policies";

const cfg = new pulumi.Config();
const organizationName = cfg.get("organizationName") ?? "service-provider-test-org";
const digits = cfg.get("digits") ?? "00000";

const admins = new teams.Team("admins", {
    organizationName,
    teamType: "pulumi",
    name: `admins-${digits}`,
    displayName: "Organization Admins",
    description: "Full org control; rotate this membership quarterly.",
});

const deployers = new teams.Team("deployers", {
    organizationName,
    teamType: "pulumi",
    name: `deployers-${digits}`,
    displayName: "CI Deployers",
    description: "Automation-only team. Human members discouraged — use the CI token.",
});

new teams.Team("readers", {
    organizationName,
    teamType: "pulumi",
    name: `readers-${digits}`,
    displayName: "Developers (read-only)",
    description: "Default team for new org members; grants stack read access.",
});

const ciToken = new tokens.TeamAccessToken("ciToken", {
    organizationName,
    teamName: deployers.name.apply(n => n!),
    description: "Used by GitHub Actions to deploy non-production stacks.",
});

new policies.PolicyGroup("defaultGuardrails", {
    organizationName,
    name: `baseline-guardrails-${digits}`,
});

export const ciTokenValue = ciToken.value;
