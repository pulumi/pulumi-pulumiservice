using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var suffix = config.Get("suffix") ?? "dev";
    var prodApprovalEnabled = config.GetBoolean("prodApprovalEnabled") ?? true;
    var slackWebhookUrl = config.Get("slackWebhookUrl") ??
        "https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap";
    var pagerDutyWebhookUrl = config.Get("pagerDutyWebhookUrl") ??
        "https://events.pagerduty.com/v2/enqueue";

    new Ps.V2.DefaultOrganization("defaultOrg", new() { OrgName = serviceOrg });

    new Ps.V2.Auth.OidcIssuer("githubIssuer", new()
    {
        OrgName = serviceOrg,
        Name = $"github_issuer_{suffix}",
        Url = "https://token.actions.githubusercontent.com",
        Thumbprints = new[] { "caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7" },
        MaxExpiration = 3600,
    });
    new Ps.V2.Auth.OidcIssuer("pulumiSelfIssuer", new()
    {
        OrgName = serviceOrg,
        Name = $"pulumi_issuer_{suffix}",
        Url = "https://api.pulumi.com/oidc",
        Thumbprints = new[] { "57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da" },
    });

    var platformTeam = new Ps.V2.Teams.Team("platformTeam", new()
    {
        OrgName = serviceOrg,
        Name = $"platform-team-{suffix}",
        DisplayName = $"Platform Team {suffix}",
        Description = "Owns shared infra, runs the deployments engine.",
    });

    var roleDetails = ImmutableDictionary<string, object>.Empty
        .Add("__type", "PermissionDescriptorAllow")
        .Add("permissions", new[] { "stack:read" });
    new Ps.V2.Role("stackReadonlyRole", new()
    {
        OrgName = serviceOrg,
        Name = $"stack-readonly-{suffix}",
        Description = "Read-only access to stacks, scoped via the platform team.",
        UxPurpose = "role",
        Details = roleDetails,
    });

    var ciToken = new Ps.V2.Tokens.OrgToken("ciToken", new()
    {
        OrgName = serviceOrg,
        Name = $"ci-{suffix}",
        Description = "Used by CI/CD to deploy non-prod stacks.",
        Admin = false,
        Expires = 0,
    });
    new Ps.V2.Tokens.TeamToken("teamToken", new()
    {
        OrgName = serviceOrg,
        TeamName = platformTeam.Name,
        Name = $"platform-team-token-{suffix}",
        Description = "Platform-team-scoped token for shared automation.",
        Expires = 0,
    });

    var runnersPool = new Ps.V2.Agents.Pool("runnersPool", new()
    {
        OrgName = serviceOrg,
        Name = $"platform-runners-{suffix}",
        Description = "Self-hosted deployment runner pool.",
    });

    var templates = new Ps.V2.OrgTemplateCollection("templates", new()
    {
        OrgName = serviceOrg,
        Name = $"platform-templates-{suffix}",
        SourceURL = "https://github.com/pulumi/examples",
    });

    var sharedCredentials = new Ps.V2.Esc.Environment("sharedCredentials", new()
    {
        OrgName = serviceOrg,
        Project = "shared",
        Name = $"credentials-{suffix}",
    });
    new Ps.V2.Esc.EnvironmentTag("stableTag", new()
    {
        OrgName = serviceOrg,
        ProjectName = sharedCredentials.Project,
        EnvName = sharedCredentials.Name,
        Name = "stable",
        Value = "1",
    });

    var stagingStack = new Ps.V2.Stacks.Stack("stagingStack", new()
    {
        OrgName = serviceOrg,
        ProjectName = $"platform-app-{suffix}",
        StackName = "staging",
    });
    var prodStack = new Ps.V2.Stacks.Stack("prodStack", new()
    {
        OrgName = serviceOrg,
        ProjectName = $"platform-app-{suffix}",
        StackName = "prod",
    });

    var sharedEnvRef = Output.Format($"{sharedCredentials.Project}/{sharedCredentials.Name}");

    new Ps.V2.Stacks.Config("stagingConfig", new()
    {
        OrgName = serviceOrg,
        ProjectName = stagingStack.ProjectName,
        StackName = stagingStack.StackName,
        Environment = sharedEnvRef,
    });
    new Ps.V2.Stacks.Config("prodConfig", new()
    {
        OrgName = serviceOrg,
        ProjectName = prodStack.ProjectName,
        StackName = prodStack.StackName,
        Environment = sharedEnvRef,
    });

    foreach (var (k, v) in new[] {
        ("owner", "platform-team"), ("tier", "gold"), ("cost-center", "platform"),
    })
    {
        new Ps.V2.Stacks.Tag($"prodTag-{k}", new()
        {
            OrgName = serviceOrg,
            ProjectName = prodStack.ProjectName,
            StackName = prodStack.StackName,
            Name = k,
            Value = v,
        });
    }

    new Ps.V2.Stacks.Webhook("prodPagerDuty", new()
    {
        OrganizationName = serviceOrg,
        ProjectName = prodStack.ProjectName,
        StackName = prodStack.StackName,
        Name = "prod-pagerduty",
        DisplayName = "prod stack PagerDuty",
        PayloadUrl = pagerDutyWebhookUrl,
        Active = true,
        Format = "raw",
    });

    var stagingExecutor = ImmutableDictionary<string, object>.Empty
        .Add("executorImage", ImmutableDictionary<string, object>.Empty.Add("reference", "pulumi/pulumi:latest"));
    new Ps.V2.Deployments.Settings("stagingDeploySettings", new()
    {
        OrgName = serviceOrg,
        ProjectName = stagingStack.ProjectName,
        StackName = stagingStack.StackName,
        ExecutorContext = stagingExecutor,
    });
    var prodExecutor = ImmutableDictionary<string, object>.Empty
        .Add("executorImage", ImmutableDictionary<string, object>.Empty.Add("reference", "pulumi/pulumi:3-nonroot"));
    var prodDeploySettings = new Ps.V2.Deployments.Settings("prodDeploySettings", new()
    {
        OrgName = serviceOrg,
        ProjectName = prodStack.ProjectName,
        StackName = prodStack.StackName,
        ExecutorContext = prodExecutor,
    });

    var gateRule = ImmutableDictionary<string, object>.Empty
        .Add("ruleType", "approval_required")
        .Add("numApprovalsRequired", 1)
        .Add("allowSelfApproval", false)
        .Add("requireReapprovalOnChange", true)
        .Add("eligibleApprovers", new[] {
            ImmutableDictionary<string, object>.Empty
                .Add("eligibilityType", "team_member")
                .Add("teamName", platformTeam.Name),
        });
    var gateTarget = ImmutableDictionary<string, object>.Empty
        .Add("entityType", "environment")
        .Add("actionTypes", new[] { "update" })
        .Add("qualifiedName", sharedEnvRef);
    new Ps.V2.Gate("credsApprovalGate", new()
    {
        OrgName = serviceOrg,
        Name = $"creds-approval-{suffix}",
        Enabled = prodApprovalEnabled,
        Rule = gateRule,
        Target = gateTarget,
    });

    var deployRequest = ImmutableDictionary<string, object>.Empty
        .Add("operation", "update")
        .Add("inheritSettings", true);
    new Ps.V2.Deployments.ScheduledDeployment("prodNightlyDeploy", new()
    {
        OrgName = serviceOrg,
        ProjectName = prodStack.ProjectName,
        StackName = prodStack.StackName,
        ScheduleCron = "0 7 * * *",
        Request = deployRequest,
    }, new CustomResourceOptions { DependsOn = { prodDeploySettings } });

    new Ps.V2.OrganizationWebhook("slack", new()
    {
        OrganizationName = serviceOrg,
        Name = $"org-slack-{suffix}",
        DisplayName = "Org Slack notifications",
        PayloadUrl = slackWebhookUrl,
        Active = true,
        Format = "slack",
    });

    new Ps.V2.PolicyGroup("starterPolicyGroup", new()
    {
        OrgName = serviceOrg,
        Name = $"platform-policies-{suffix}",
        EntityType = "stacks",
    });

    return new Dictionary<string, object?>
    {
        ["platformTeamName"] = platformTeam.Name,
        ["ciTokenId"] = ciToken.TokenId,
        ["agentPoolName"] = runnersPool.Name,
        ["templatesName"] = templates.Name,
        ["sharedCredsEnv"] = sharedEnvRef,
        ["prodStackId"] = prodStack.Id,
    };
});
