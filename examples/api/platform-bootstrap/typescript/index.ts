import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const suffix = config.get("suffix") ?? "dev";
const prodApprovalEnabled = config.getBoolean("prodApprovalEnabled") ?? true;
const slackWebhookUrl = config.get("slackWebhookUrl") ??
    "https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap";
const pagerDutyWebhookUrl = config.get("pagerDutyWebhookUrl") ??
    "https://events.pagerduty.com/v2/enqueue";

// 1. Org-level user preference.
new ps.api.DefaultOrganization("defaultOrg", { orgName: organizationName });

// 2. OIDC issuers: GitHub Actions trust + Pulumi Cloud self-trust.
new ps.api.auth.OidcIssuer("githubIssuer", {
    orgName: organizationName,
    name: `github_issuer_${suffix}`,
    url: "https://token.actions.githubusercontent.com",
    thumbprints: ["b41ae0832808ebc94951437bf7e92b93ccb6479364daf894d46d6001bee7a486"],
    maxExpiration: 3600,
});
new ps.api.auth.OidcIssuer("pulumiSelfIssuer", {
    orgName: organizationName,
    name: `pulumi_issuer_${suffix}`,
    url: "https://api.pulumi.com/oidc",
    thumbprints: ["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
});

// 3. Platform team + a stack-readonly role.
const platformTeam = new ps.api.teams.Team("platformTeam", {
    orgName: organizationName,
    name: `platform-team-${suffix}`,
    displayName: `Platform Team ${suffix}`,
    description: "Owns shared infra, runs the deployments engine.",
});
const stackReadonlyRole = new ps.api.Role("stackReadonlyRole", {
    orgName: organizationName,
    name: `stack-readonly-${suffix}`,
    description: "Read-only access to stacks, scoped via the platform team.",
    uxPurpose: "role",
    details: {
        __type: "PermissionDescriptorAllow",
        permissions: ["stack:read"],
    },
});

// 4. CI machine token + team-scoped token.
const ciToken = new ps.api.tokens.OrgToken("ciToken", {
    orgName: organizationName,
    name: `ci-${suffix}`,
    description: "Used by CI/CD to deploy non-prod stacks.",
    admin: false,
    expires: 0,
});
new ps.api.tokens.TeamToken("teamToken", {
    orgName: organizationName,
    teamName: platformTeam.name,
    name: `platform-team-token-${suffix}`,
    description: "Platform-team-scoped token for shared automation.",
    expires: 0,
});

// 5. Self-hosted deployment runner.
const runnersPool = new ps.api.agents.Pool("runnersPool", {
    orgName: organizationName,
    name: `platform-runners-${suffix}`,
    description: "Self-hosted deployment runner pool.",
});

// 6. New-project template seed.
const templates = new ps.api.OrgTemplateCollection("templates", {
    orgName: organizationName,
    name: `platform-templates-${suffix}`,
    sourceURL: "https://github.com/pulumi/examples",
});

// 7. Shared ESC credentials env + a "stable" tag on it.
const sharedCredentials = new ps.api.esc.Environment("sharedCredentials", {
    orgName: organizationName,
    project: "shared",
    name: `credentials-${suffix}`,
});
new ps.api.esc.EnvironmentTag("stableTag", {
    orgName: organizationName,
    projectName: sharedCredentials.project,
    envName: sharedCredentials.name,
    name: "stable",
    value: "1",
});

// 8. Two stacks (staging + prod).
const stagingStack = new ps.api.stacks.Stack("stagingStack", {
    orgName: organizationName,
    projectName: `platform-app-${suffix}`,
    stackName: "staging",
});
const prodStack = new ps.api.stacks.Stack("prodStack", {
    orgName: organizationName,
    projectName: `platform-app-${suffix}`,
    stackName: "prod",
});

// 9. StackConfig: bind each stack to the shared ESC env.
const sharedEnvRef = pulumi.interpolate`${sharedCredentials.project}/${sharedCredentials.name}`;
new ps.api.stacks.Config("stagingConfig", {
    orgName: organizationName,
    projectName: stagingStack.projectName,
    stackName: stagingStack.stackName,
    environment: sharedEnvRef,
});
new ps.api.stacks.Config("prodConfig", {
    orgName: organizationName,
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
    new ps.api.stacks.Tag(`prodTag-${k}`, {
        orgName: organizationName,
        projectName: prodStack.projectName,
        stackName: prodStack.stackName,
        name: k,
        value: v,
    });
}

// 11. Per-stack PagerDuty webhook on prod.
new ps.api.stacks.Webhook("prodPagerDuty", {
    organizationName: organizationName,
    projectName: prodStack.projectName,
    stackName: prodStack.stackName,
    name: "prod-pagerduty",
    displayName: "prod stack PagerDuty",
    payloadUrl: pagerDutyWebhookUrl,
    active: true,
    format: "raw",
});

// 12. DeploymentSettings: different executor per environment.
new ps.api.deployments.Settings("stagingDeploySettings", {
    orgName: organizationName,
    projectName: stagingStack.projectName,
    stackName: stagingStack.stackName,
    executorContext: { executorImage: { reference: "pulumi/pulumi:latest" } },
});
const prodDeploySettings = new ps.api.deployments.Settings("prodDeploySettings", {
    orgName: organizationName,
    projectName: prodStack.projectName,
    stackName: prodStack.stackName,
    executorContext: { executorImage: { reference: "pulumi/pulumi:3-nonroot" } },
});

// 13. Approval gate on the credentials env.
new ps.api.Gate("credsApprovalGate", {
    orgName: organizationName,
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
new ps.api.deployments.ScheduledDeployment("prodNightlyDeploy", {
    orgName: organizationName,
    projectName: prodStack.projectName,
    stackName: prodStack.stackName,
    scheduleCron: "0 7 * * *",
    request: { operation: "update", inheritSettings: true },
}, { dependsOn: [prodDeploySettings] });

// 15. Org-level Slack webhook.
new ps.api.OrganizationWebhook("slack", {
    organizationName: organizationName,
    name: `org-slack-${suffix}`,
    displayName: "Org Slack notifications",
    payloadUrl: slackWebhookUrl,
    active: true,
    format: "slack",
});

// 16. Starter PolicyGroup.
new ps.api.PolicyGroup("starterPolicyGroup", {
    orgName: organizationName,
    name: `platform-policies-${suffix}`,
    entityType: "stacks",
});

// 17. Bind the stack-readonly role to the platform team.
new ps.api.teams.Role("platformTeamRole", {
    orgName: organizationName,
    teamName: platformTeam.name,
    roleID: stackReadonlyRole.id,
});

export const platformTeamName = platformTeam.name;
export const ciTokenId = ciToken.tokenId;
export const agentPoolName = runnersPool.name;
export const templatesName = templates.name;
export const sharedCredsEnv = sharedEnvRef;
export const prodStackId = prodStack.id;
