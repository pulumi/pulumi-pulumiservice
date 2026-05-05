import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const suffix = config.get("suffix") ?? "dev";
const prodApprovalEnabled = config.getBoolean("prodApprovalEnabled") ?? true;
const slackWebhookUrl = config.get("slackWebhookUrl") ??
    "https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap";
const pagerDutyWebhookUrl = config.get("pagerDutyWebhookUrl") ??
    "https://events.pagerduty.com/v2/enqueue";

// 1. Org-level user preference.
new ps.v2.DefaultOrganization("defaultOrg", { orgName: serviceOrg });

// 2. OIDC issuers: GitHub Actions trust + Pulumi Cloud self-trust.
new ps.v2.auth.OidcIssuer("githubIssuer", {
    orgName: serviceOrg,
    name: `github_issuer_${suffix}`,
    url: "https://token.actions.githubusercontent.com",
    thumbprints: ["caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"],
    maxExpiration: 3600,
});
new ps.v2.auth.OidcIssuer("pulumiSelfIssuer", {
    orgName: serviceOrg,
    name: `pulumi_issuer_${suffix}`,
    url: "https://api.pulumi.com/oidc",
    thumbprints: ["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
});

// 3. Platform team + a stack-readonly role.
const platformTeam = new ps.v2.teams.Team("platformTeam", {
    orgName: serviceOrg,
    name: `platform-team-${suffix}`,
    displayName: `Platform Team ${suffix}`,
    description: "Owns shared infra, runs the deployments engine.",
});
const stackReadonlyRole = new ps.v2.Role("stackReadonlyRole", {
    orgName: serviceOrg,
    name: `stack-readonly-${suffix}`,
    description: "Read-only access to stacks, scoped via the platform team.",
    uxPurpose: "role",
    details: {
        __type: "PermissionDescriptorAllow",
        permissions: ["stack:read"],
    },
});

// 4. CI machine token + team-scoped token.
const ciToken = new ps.v2.tokens.OrgToken("ciToken", {
    orgName: serviceOrg,
    name: `ci-${suffix}`,
    description: "Used by CI/CD to deploy non-prod stacks.",
    admin: false,
    expires: 0,
});
new ps.v2.tokens.TeamToken("teamToken", {
    orgName: serviceOrg,
    teamName: platformTeam.name,
    name: `platform-team-token-${suffix}`,
    description: "Platform-team-scoped token for shared automation.",
    expires: 0,
});

// 5. Self-hosted deployment runner.
const runnersPool = new ps.v2.agents.Pool("runnersPool", {
    orgName: serviceOrg,
    name: `platform-runners-${suffix}`,
    description: "Self-hosted deployment runner pool.",
});

// 6. New-project template seed.
const templates = new ps.v2.OrgTemplateCollection("templates", {
    orgName: serviceOrg,
    name: `platform-templates-${suffix}`,
    sourceURL: "https://github.com/pulumi/examples",
});

// 7. Shared ESC credentials env + a "stable" tag on it.
const sharedCredentials = new ps.v2.esc.Environment("sharedCredentials", {
    orgName: serviceOrg,
    project: "shared",
    name: `credentials-${suffix}`,
});
new ps.v2.esc.EnvironmentTag("stableTag", {
    orgName: serviceOrg,
    projectName: sharedCredentials.project,
    envName: sharedCredentials.name,
    name: "stable",
    value: "1",
});

// 8. Two stacks (staging + prod).
const stagingStack = new ps.v2.stacks.Stack("stagingStack", {
    orgName: serviceOrg,
    projectName: `platform-app-${suffix}`,
    stackName: "staging",
});
const prodStack = new ps.v2.stacks.Stack("prodStack", {
    orgName: serviceOrg,
    projectName: `platform-app-${suffix}`,
    stackName: "prod",
});

// 9. StackConfig: bind each stack to the shared ESC env.
const sharedEnvRef = pulumi.interpolate`${sharedCredentials.project}/${sharedCredentials.name}`;
new ps.v2.stacks.Config("stagingConfig", {
    orgName: serviceOrg,
    projectName: stagingStack.projectName,
    stackName: stagingStack.stackName,
    environment: sharedEnvRef,
});
new ps.v2.stacks.Config("prodConfig", {
    orgName: serviceOrg,
    projectName: prodStack.projectName,
    stackName: prodStack.stackName,
    environment: sharedEnvRef,
});

// 10. Tags on prod.
for (const [k, v] of Object.entries({
    owner: "platform-team",
    tier: "gold",
    "cost-center": "platform",
})) {
    new ps.v2.stacks.Tag(`prodTag-${k}`, {
        orgName: serviceOrg,
        projectName: prodStack.projectName,
        stackName: prodStack.stackName,
        name: k,
        value: v,
    });
}

// 11. Per-stack PagerDuty webhook on prod.
new ps.v2.stacks.Webhook("prodPagerDuty", {
    organizationName: serviceOrg,
    projectName: prodStack.projectName,
    stackName: prodStack.stackName,
    name: "prod-pagerduty",
    displayName: "prod stack PagerDuty",
    payloadUrl: pagerDutyWebhookUrl,
    active: true,
    format: "raw",
});

// 12. DeploymentSettings: different executor per environment.
new ps.v2.deployments.Settings("stagingDeploySettings", {
    orgName: serviceOrg,
    projectName: stagingStack.projectName,
    stackName: stagingStack.stackName,
    executorContext: { executorImage: { reference: "pulumi/pulumi:latest" } },
});
const prodDeploySettings = new ps.v2.deployments.Settings("prodDeploySettings", {
    orgName: serviceOrg,
    projectName: prodStack.projectName,
    stackName: prodStack.stackName,
    executorContext: { executorImage: { reference: "pulumi/pulumi:3-nonroot" } },
});

// 13. Approval gate on the credentials env.
new ps.v2.Gate("credsApprovalGate", {
    orgName: serviceOrg,
    name: `creds-approval-${suffix}`,
    enabled: prodApprovalEnabled,
    rule: {
        ruleType: "approval_required",
        numApprovalsRequired: 1,
        allowSelfApproval: false,
        requireReapprovalOnChange: true,
        eligibleApprovers: [
            { eligibilityType: "team_member", teamName: platformTeam.name },
        ],
    },
    target: {
        entityType: "environment",
        actionTypes: ["update"],
        qualifiedName: sharedEnvRef,
    },
});

// 14. Nightly redeploy of prod (depends on prodDeploySettings).
new ps.v2.deployments.ScheduledDeployment("prodNightlyDeploy", {
    orgName: serviceOrg,
    projectName: prodStack.projectName,
    stackName: prodStack.stackName,
    scheduleCron: "0 7 * * *",
    request: { operation: "update", inheritSettings: true },
}, { dependsOn: [prodDeploySettings] });

// 15. Org-level Slack webhook.
new ps.v2.OrganizationWebhook("slack", {
    organizationName: serviceOrg,
    name: `org-slack-${suffix}`,
    displayName: "Org Slack notifications",
    payloadUrl: slackWebhookUrl,
    active: true,
    format: "slack",
});

// 16. Starter PolicyGroup.
new ps.v2.PolicyGroup("starterPolicyGroup", {
    orgName: serviceOrg,
    name: `platform-policies-${suffix}`,
    entityType: "stacks",
});

// 17. Bind the stack-readonly role to the platform team.
new ps.v2.teams.Role("platformTeamRole", {
    orgName: serviceOrg,
    teamName: platformTeam.name,
    roleID: stackReadonlyRole.id,
});

export const platformTeamName = platformTeam.name;
export const ciTokenId = ciToken.tokenId;
export const agentPoolName = runnersPool.name;
export const templatesName = templates.name;
export const sharedCredsEnv = sharedEnvRef;
export const prodStackId = prodStack.id;
