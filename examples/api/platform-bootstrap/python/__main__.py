import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
suffix = config.get("suffix") or "dev"
prod_approval_enabled = config.get_bool("prodApprovalEnabled")
if prod_approval_enabled is None:
    prod_approval_enabled = True
slack_webhook_url = config.get("slackWebhookUrl") or \
    "https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap"
pager_duty_webhook_url = config.get("pagerDutyWebhookUrl") or \
    "https://events.pagerduty.com/v2/enqueue"

ps_api.DefaultOrganization("defaultOrg", org_name=organization_name)

ps_api.auth.OidcIssuer(
    "githubIssuer",
    org_name=organization_name,
    name=f"github_issuer_{suffix}",
    url="https://token.actions.githubusercontent.com",
    thumbprints=["39517789ff0132a9212bafea4dc37401eae58b1bfac9756109d14301c90a6ab5"],
    max_expiration=3600,
)
ps_api.auth.OidcIssuer(
    "pulumiSelfIssuer",
    org_name=organization_name,
    name=f"pulumi_issuer_{suffix}",
    url="https://api.pulumi.com/oidc",
    thumbprints=["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
)

platform_team = ps_api.teams.Team(
    "platformTeam",
    org_name=organization_name,
    name=f"platform-team-{suffix}",
    display_name=f"Platform Team {suffix}",
    description="Owns shared infra, runs the deployments engine.",
)

ps_api.Role(
    "stackReadonlyRole",
    org_name=organization_name,
    name=f"stack-readonly-{suffix}",
    description="Read-only access to stacks, scoped via the platform team.",
    ux_purpose="role",
    details={
        "__type": "PermissionDescriptorAllow",
        "permissions": ["stack:read"],
    },
)

ci_token = ps_api.tokens.OrgToken(
    "ciToken",
    org_name=organization_name,
    name=f"ci-{suffix}",
    description="Used by CI/CD to deploy non-prod stacks.",
    admin=False,
    expires=0,
)
ps_api.tokens.TeamToken(
    "teamToken",
    org_name=organization_name,
    team_name=platform_team.name,
    name=f"platform-team-token-{suffix}",
    description="Platform-team-scoped token for shared automation.",
    expires=0,
)

runners_pool = ps_api.agents.Pool(
    "runnersPool",
    org_name=organization_name,
    name=f"platform-runners-{suffix}",
    description="Self-hosted deployment runner pool.",
)

templates = ps_api.OrgTemplateCollection(
    "templates",
    org_name=organization_name,
    name=f"platform-templates-{suffix}",
    source_url="https://github.com/pulumi/examples",
)

shared_credentials = ps_api.esc.Environment(
    "sharedCredentials",
    org_name=organization_name,
    project="shared",
    name=f"credentials-{suffix}",
)
ps_api.esc.EnvironmentTag(
    "stableTag",
    org_name=organization_name,
    project_name="shared",
    env_name=f"credentials-{suffix}",
    name="stable",
    value="1",
    opts=pulumi.ResourceOptions(depends_on=[shared_credentials]),
)

staging_stack = ps_api.stacks.Stack(
    "stagingStack",
    org_name=organization_name,
    project_name=f"platform-app-{suffix}",
    stack_name="staging",
)
prod_stack = ps_api.stacks.Stack(
    "prodStack",
    org_name=organization_name,
    project_name=f"platform-app-{suffix}",
    stack_name="prod",
)

shared_env_ref = f"shared/credentials-{suffix}"

ps_api.stacks.Config(
    "stagingConfig",
    org_name=organization_name,
    project_name=staging_stack.project_name,
    stack_name=staging_stack.stack_name,
    environment=shared_env_ref,
    opts=pulumi.ResourceOptions(depends_on=[shared_credentials]),
)
ps_api.stacks.Config(
    "prodConfig",
    org_name=organization_name,
    project_name=prod_stack.project_name,
    stack_name=prod_stack.stack_name,
    environment=shared_env_ref,
    opts=pulumi.ResourceOptions(depends_on=[shared_credentials]),
)

for k, v in [("owner", "platform-team"), ("tier", "gold"), ("cost-center", "platform")]:
    ps_api.stacks.Tag(
        f"prodTag-{k}",
        org_name=organization_name,
        project_name=prod_stack.project_name,
        stack_name=prod_stack.stack_name,
        name=k,
        value=v,
    )

ps_api.stacks.Webhook(
    "prodPagerDuty",
    organization_name=organization_name,
    project_name=prod_stack.project_name,
    stack_name=prod_stack.stack_name,
    name="prod-pagerduty",
    display_name="prod stack PagerDuty",
    payload_url=pager_duty_webhook_url,
    active=True,
    format="raw",
)

ps_api.deployments.Settings(
    "stagingDeploySettings",
    org_name=organization_name,
    project_name=staging_stack.project_name,
    stack_name=staging_stack.stack_name,
    executor_context={"executorImage": {"reference": "pulumi/pulumi:latest"}},
)
prod_deploy_settings = ps_api.deployments.Settings(
    "prodDeploySettings",
    org_name=organization_name,
    project_name=prod_stack.project_name,
    stack_name=prod_stack.stack_name,
    executor_context={"executorImage": {"reference": "pulumi/pulumi:3-nonroot"}},
)

ps_api.Gate(
    "credsApprovalGate",
    org_name=organization_name,
    name=f"creds-approval-{suffix}",
    enabled=prod_approval_enabled,
    rule={
        "ruleType": "approval_required",
        "numApprovalsRequired": 1,
        "allowSelfApproval": False,
        "requireReapprovalOnChange": True,
        "eligibleApprovers": [
            {"eligibilityType": "team_member", "teamName": platform_team.name},
        ],
    },
    target={
        "entityType": "environment",
        "actionTypes": ["update"],
        "qualifiedName": shared_env_ref,
    },
    opts=pulumi.ResourceOptions(depends_on=[shared_credentials]),
)

ps_api.deployments.ScheduledDeployment(
    "prodNightlyDeploy",
    org_name=organization_name,
    project_name=prod_stack.project_name,
    stack_name=prod_stack.stack_name,
    schedule_cron="0 7 * * *",
    request={"operation": "update", "inheritSettings": True},
    opts=pulumi.ResourceOptions(depends_on=[prod_deploy_settings]),
)

ps_api.OrganizationWebhook(
    "slack",
    organization_name=organization_name,
    name=f"org-slack-{suffix}",
    display_name="Org Slack notifications",
    payload_url=slack_webhook_url,
    active=True,
    format="slack",
)

ps_api.PolicyGroup(
    "starterPolicyGroup",
    org_name=organization_name,
    name=f"platform-policies-{suffix}",
    entity_type="stacks",
)

pulumi.export("platformTeamName", platform_team.name)
pulumi.export("ciTokenId", ci_token.token_id)
pulumi.export("agentPoolName", runners_pool.name)
pulumi.export("templatesName", templates.name)
pulumi.export("sharedCredsEnv", shared_env_ref)
pulumi.export("prodStackId", prod_stack.id)
