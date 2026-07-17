using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var suffix = config.Get("suffix") ?? "dev";
    var prodApprovalEnabled = config.GetBoolean("prodApprovalEnabled") ?? true;
    var slackWebhookUrl = config.Get("slackWebhookUrl") ??
        "https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap";
    var pagerDutyWebhookUrl = config.Get("pagerDutyWebhookUrl") ??
        "https://events.pagerduty.com/v2/enqueue";

    new Ps.Api.DefaultOrganization("defaultOrg", new() { OrgName = organizationName });

    new Ps.Api.Auth.OidcIssuer("githubIssuer", new()
    {
        OrgName = organizationName,
        Name = $"github_issuer_{suffix}",
        Url = "https://token.actions.githubusercontent.com",
        Thumbprints = new[] { "39517789ff0132a9212bafea4dc37401eae58b1bfac9756109d14301c90a6ab5" },
        MaxExpiration = 3600,
    });
    new Ps.Api.Auth.OidcIssuer("pulumiSelfIssuer", new()
    {
        OrgName = organizationName,
        Name = $"pulumi_issuer_{suffix}",
        Url = "https://api.pulumi.com/oidc",
        Thumbprints = new[] { "57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da" },
    });

    var platformTeam = new Ps.Api.Teams.Team("platformTeam", new()
    {
        OrgName = organizationName,
        Name = $"platform-team-{suffix}",
        DisplayName = $"Platform Team {suffix}",
        Description = "Owns shared infra, runs the deployments engine.",
    });

    var roleDetails = ImmutableDictionary<string, object>.Empty
        .Add("__type", "PermissionDescriptorAllow")
        .Add("permissions", new[] { "stack:read" });
    new Ps.Api.Role("stackReadonlyRole", new()
    {
        OrgName = organizationName,
        Name = $"stack-readonly-{suffix}",
        Description = "Read-only access to stacks, scoped via the platform team.",
        UxPurpose = "role",
        Details = roleDetails,
    });

    var ciToken = new Ps.Api.Tokens.OrgToken("ciToken", new()
    {
        OrgName = organizationName,
        Name = $"ci-{suffix}",
        Description = "Used by CI/CD to deploy non-prod stacks.",
        Admin = false,
        Expires = 0,
    });
    new Ps.Api.Tokens.TeamToken("teamToken", new()
    {
        OrgName = organizationName,
        TeamName = platformTeam.Name,
        Name = $"platform-team-token-{suffix}",
        Description = "Platform-team-scoped token for shared automation.",
        Expires = 0,
    });

    var runnersPool = new Ps.Api.Agents.Pool("runnersPool", new()
    {
        OrgName = organizationName,
        Name = $"platform-runners-{suffix}",
        Description = "Self-hosted deployment runner pool.",
    });

    var templates = new Ps.Api.OrgTemplateCollection("templates", new()
    {
        OrgName = organizationName,
        Name = $"platform-templates-{suffix}",
        SourceURL = "https://github.com/pulumi/examples",
    });

    var sharedEnvName = $"credentials-{suffix}";
    var sharedCredentials = new Ps.Api.Esc.Environment("sharedCredentials", new()
    {
        OrgName = organizationName,
        Project = "shared",
        Name = sharedEnvName,
    });
    new Ps.Api.Esc.EnvironmentTag("stableTag", new()
    {
        OrgName = organizationName,
        ProjectName = "shared",
        EnvName = sharedEnvName,
        Name = "stable",
        Value = "1",
    }, new CustomResourceOptions { DependsOn = { sharedCredentials } });

    var stagingStack = new Ps.Api.Stacks.Stack("stagingStack", new()
    {
        OrgName = organizationName,
        ProjectName = $"platform-app-{suffix}",
        StackName = "staging",
    });
    var prodStack = new Ps.Api.Stacks.Stack("prodStack", new()
    {
        OrgName = organizationName,
        ProjectName = $"platform-app-{suffix}",
        StackName = "prod",
    });

    var sharedEnvRef = $"shared/{sharedEnvName}";

    new Ps.Api.Stacks.Config("stagingConfig", new()
    {
        OrgName = organizationName,
        ProjectName = stagingStack.ProjectName,
        StackName = stagingStack.StackName,
        Environment = sharedEnvRef,
    }, new CustomResourceOptions { DependsOn = { sharedCredentials } });
    new Ps.Api.Stacks.Config("prodConfig", new()
    {
        OrgName = organizationName,
        ProjectName = prodStack.ProjectName,
        StackName = prodStack.StackName,
        Environment = sharedEnvRef,
    }, new CustomResourceOptions { DependsOn = { sharedCredentials } });

    foreach (var (k, v) in new[] {
        ("owner", "platform-team"), ("tier", "gold"), ("cost-center", "platform"),
    })
    {
        new Ps.Api.Stacks.Tag($"prodTag-{k}", new()
        {
            OrgName = organizationName,
            ProjectName = prodStack.ProjectName,
            StackName = prodStack.StackName,
            Name = k,
            Value = v,
        });
    }

    new Ps.Api.Stacks.Webhook("prodPagerDuty", new()
    {
        OrganizationName = organizationName,
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
    new Ps.Api.Deployments.Settings("stagingDeploySettings", new()
    {
        OrgName = organizationName,
        ProjectName = stagingStack.ProjectName,
        StackName = stagingStack.StackName,
        ExecutorContext = stagingExecutor,
    });
    var prodExecutor = ImmutableDictionary<string, object>.Empty
        .Add("executorImage", ImmutableDictionary<string, object>.Empty.Add("reference", "pulumi/pulumi:3-nonroot"));
    var prodDeploySettings = new Ps.Api.Deployments.Settings("prodDeploySettings", new()
    {
        OrgName = organizationName,
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
    new Ps.Api.Gate("credsApprovalGate", new()
    {
        OrgName = organizationName,
        Name = $"creds-approval-{suffix}",
        Enabled = prodApprovalEnabled,
        Rule = gateRule,
        Target = gateTarget,
    }, new CustomResourceOptions { DependsOn = { sharedCredentials } });

    var deployRequest = ImmutableDictionary<string, object>.Empty
        .Add("operation", "update")
        .Add("inheritSettings", true);
    new Ps.Api.Deployments.ScheduledDeployment("prodNightlyDeploy", new()
    {
        OrgName = organizationName,
        ProjectName = prodStack.ProjectName,
        StackName = prodStack.StackName,
        ScheduleCron = "0 7 * * *",
        Request = deployRequest,
    }, new CustomResourceOptions { DependsOn = { prodDeploySettings } });

    new Ps.Api.OrganizationWebhook("slack", new()
    {
        OrganizationName = organizationName,
        Name = $"org-slack-{suffix}",
        DisplayName = "Org Slack notifications",
        PayloadUrl = slackWebhookUrl,
        Active = true,
        Format = "slack",
    });

    new Ps.Api.PolicyGroup("starterPolicyGroup", new()
    {
        OrgName = organizationName,
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
