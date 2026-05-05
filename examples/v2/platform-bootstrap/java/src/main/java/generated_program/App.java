package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.resources.CustomResourceOptions;
import com.pulumi.pulumiservice.v2.*;
import com.pulumi.pulumiservice.v2_deployments.*;
import com.pulumi.pulumiservice.v2_stacks.*;
import com.pulumi.pulumiservice.v2_agents.*;
// v2_teams imported explicitly (not wildcarded) so v2_teams.Role doesn't
// collide with v2.Role (the org-level FGA role used below).
import com.pulumi.pulumiservice.v2_teams.Team;
import com.pulumi.pulumiservice.v2_teams.TeamArgs;
import com.pulumi.pulumiservice.v2_auth.OidcIssuer;
import com.pulumi.pulumiservice.v2_auth.OidcIssuerArgs;
import com.pulumi.pulumiservice.v2_esc.Environment;
import com.pulumi.pulumiservice.v2_esc.EnvironmentArgs;
import com.pulumi.pulumiservice.v2_esc.EnvironmentTag;
import com.pulumi.pulumiservice.v2_esc.EnvironmentTagArgs;
import com.pulumi.pulumiservice.v2_tokens.OrgToken;
import com.pulumi.pulumiservice.v2_tokens.OrgTokenArgs;
import com.pulumi.pulumiservice.v2_tokens.TeamToken;
import com.pulumi.pulumiservice.v2_tokens.TeamTokenArgs;
import com.pulumi.pulumiservice.v2_deployments.Settings;
import com.pulumi.pulumiservice.v2_deployments.SettingsArgs;
import com.pulumi.pulumiservice.v2_deployments.ScheduledDeployment;
import com.pulumi.pulumiservice.v2_deployments.ScheduledDeploymentArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var suffix = config.get("suffix").orElse("dev");
            var prodApprovalEnabled = config.getBoolean("prodApprovalEnabled").orElse(true);
            var slackWebhookUrl = config.get("slackWebhookUrl")
                .orElse("https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap");
            var pagerDutyWebhookUrl = config.get("pagerDutyWebhookUrl")
                .orElse("https://events.pagerduty.com/v2/enqueue");

            new DefaultOrganization("defaultOrg",
                DefaultOrganizationArgs.builder().orgName(serviceOrg).build());

            new OidcIssuer("githubIssuer",
                OidcIssuerArgs.builder()
                    .orgName(serviceOrg)
                    .name("github_issuer_" + suffix)
                    .url("https://token.actions.githubusercontent.com")
                    .thumbprints(List.of("caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"))
                    .maxExpiration(3600)
                    .build());
            new OidcIssuer("pulumiSelfIssuer",
                OidcIssuerArgs.builder()
                    .orgName(serviceOrg)
                    .name("pulumi_issuer_" + suffix)
                    .url("https://api.pulumi.com/oidc")
                    .thumbprints(List.of("57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"))
                    .build());

            var platformTeam = new Team("platformTeam",
                TeamArgs.builder()
                    .orgName(serviceOrg)
                    .name("platform-team-" + suffix)
                    .displayName("Platform Team " + suffix)
                    .description("Owns shared infra, runs the deployments engine.")
                    .build());

            new Role("stackReadonlyRole",
                RoleArgs.builder()
                    .orgName(serviceOrg)
                    .name("stack-readonly-" + suffix)
                    .description("Read-only access to stacks, scoped via the platform team.")
                    .uxPurpose("role")
                    .details(Map.of(
                        "__type", "PermissionDescriptorAllow",
                        "permissions", List.of("stack:read")))
                    .build());

            var ciToken = new OrgToken("ciToken",
                OrgTokenArgs.builder()
                    .orgName(serviceOrg)
                    .name("ci-" + suffix)
                    .description("Used by CI/CD to deploy non-prod stacks.")
                    .admin(false)
                    .expires(0)
                    .build());
            new TeamToken("teamToken",
                TeamTokenArgs.builder()
                    .orgName(serviceOrg)
                    .teamName(platformTeam.name())
                    .name("platform-team-token-" + suffix)
                    .description("Platform-team-scoped token for shared automation.")
                    .expires(0)
                    .build());

            var runnersPool = new Pool("runnersPool",
                PoolArgs.builder()
                    .orgName(serviceOrg)
                    .name("platform-runners-" + suffix)
                    .description("Self-hosted deployment runner pool.")
                    .build());

            var templates = new OrgTemplateCollection("templates",
                OrgTemplateCollectionArgs.builder()
                    .orgName(serviceOrg)
                    .name("platform-templates-" + suffix)
                    .sourceURL("https://github.com/pulumi/examples")
                    .build());

            var sharedCredentials = new Environment("sharedCredentials",
                EnvironmentArgs.builder()
                    .orgName(serviceOrg)
                    .project("shared")
                    .name("credentials-" + suffix)
                    .build());
            new EnvironmentTag("stableTag",
                EnvironmentTagArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(sharedCredentials.project())
                    .envName(sharedCredentials.name())
                    .name("stable")
                    .value("1")
                    .build());

            var stagingStack = new Stack("stagingStack",
                StackArgs.builder()
                    .orgName(serviceOrg)
                    .projectName("platform-app-" + suffix)
                    .stackName("staging")
                    .build());
            var prodStack = new Stack("prodStack",
                StackArgs.builder()
                    .orgName(serviceOrg)
                    .projectName("platform-app-" + suffix)
                    .stackName("prod")
                    .build());

            Output<String> sharedEnvRef = Output.format("%s/%s",
                sharedCredentials.project(), sharedCredentials.name());

            new Config("stagingConfig",
                ConfigArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(stagingStack.projectName())
                    .stackName(stagingStack.stackName())
                    .environment(sharedEnvRef)
                    .build());
            new Config("prodConfig",
                ConfigArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(prodStack.projectName())
                    .stackName(prodStack.stackName())
                    .environment(sharedEnvRef)
                    .build());

            for (var entry : Map.of("owner", "platform-team", "tier", "gold", "cost-center", "platform").entrySet()) {
                new Tag("prodTag-" + entry.getKey(),
                    TagArgs.builder()
                        .orgName(serviceOrg)
                        .projectName(prodStack.projectName())
                        .stackName(prodStack.stackName())
                        .name(entry.getKey())
                        .value(entry.getValue())
                        .build());
            }

            new Webhook("prodPagerDuty",
                WebhookArgs.builder()
                    .organizationName(serviceOrg)
                    .projectName(prodStack.projectName())
                    .stackName(prodStack.stackName())
                    .name("prod-pagerduty")
                    .displayName("prod stack PagerDuty")
                    .payloadUrl(pagerDutyWebhookUrl)
                    .active(true)
                    .format("raw")
                    .build());

            new Settings("stagingDeploySettings",
                SettingsArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(stagingStack.projectName())
                    .stackName(stagingStack.stackName())
                    .executorContext(Map.of("executorImage", Map.of("reference", "pulumi/pulumi:latest")))
                    .build());
            var prodDeploySettings = new Settings("prodDeploySettings",
                SettingsArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(prodStack.projectName())
                    .stackName(prodStack.stackName())
                    .executorContext(Map.of("executorImage", Map.of("reference", "pulumi/pulumi:3-nonroot")))
                    .build());

            new Gate("credsApprovalGate",
                GateArgs.builder()
                    .orgName(serviceOrg)
                    .name("creds-approval-" + suffix)
                    .enabled(prodApprovalEnabled)
                    .rule(Map.of(
                        "ruleType", "approval_required",
                        "numApprovalsRequired", 1,
                        "allowSelfApproval", false,
                        "requireReapprovalOnChange", true,
                        "eligibleApprovers", List.of(
                            Map.of("eligibilityType", "team_member", "teamName", platformTeam.name()))))
                    .target(Map.of(
                        "entityType", "environment",
                        "actionTypes", List.of("update"),
                        "qualifiedName", sharedEnvRef))
                    .build());

            new ScheduledDeployment("prodNightlyDeploy",
                ScheduledDeploymentArgs.builder()
                    .orgName(serviceOrg)
                    .projectName(prodStack.projectName())
                    .stackName(prodStack.stackName())
                    .scheduleCron("0 7 * * *")
                    .request(Map.of("operation", "update", "inheritSettings", true))
                    .build(),
                CustomResourceOptions.builder().dependsOn(prodDeploySettings).build());

            new OrganizationWebhook("slack",
                OrganizationWebhookArgs.builder()
                    .organizationName(serviceOrg)
                    .name("org-slack-" + suffix)
                    .displayName("Org Slack notifications")
                    .payloadUrl(slackWebhookUrl)
                    .active(true)
                    .format("slack")
                    .build());

            new PolicyGroup("starterPolicyGroup",
                PolicyGroupArgs.builder()
                    .orgName(serviceOrg)
                    .name("platform-policies-" + suffix)
                    .entityType("stacks")
                    .build());

            ctx.export("platformTeamName", platformTeam.name());
            ctx.export("ciTokenId", ciToken.tokenId());
            ctx.export("agentPoolName", runnersPool.name());
            ctx.export("templatesName", templates.name());
            ctx.export("sharedCredsEnv", sharedEnvRef);
            ctx.export("prodStackId", prodStack.id());
        });
    }
}
