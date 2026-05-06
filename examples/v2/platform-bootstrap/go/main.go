package main

import (
	v2 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2"
	teams "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/teams"
	stacks "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/stacks"
	auth "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/auth"
	agents "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/agents"
	tokens "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/tokens"
	deployments "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/deployments"
	esc "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/esc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		serviceOrg := cfg.Get("serviceOrg")
		if serviceOrg == "" {
			serviceOrg = "service-provider-test-org"
		}
		suffix := cfg.Get("suffix")
		if suffix == "" {
			suffix = "dev"
		}
		prodApprovalEnabled := true
		if cfg.GetBool("prodApprovalEnabled") {
			prodApprovalEnabled = cfg.GetBool("prodApprovalEnabled")
		}
		slackWebhookUrl := cfg.Get("slackWebhookUrl")
		if slackWebhookUrl == "" {
			slackWebhookUrl = "https://hooks.slack.com/services/T00000000/B00000000/v2platformbootstrap"
		}
		pagerDutyWebhookUrl := cfg.Get("pagerDutyWebhookUrl")
		if pagerDutyWebhookUrl == "" {
			pagerDutyWebhookUrl = "https://events.pagerduty.com/v2/enqueue"
		}

		if _, err := v2.NewDefaultOrganization(ctx, "defaultOrg", &v2.DefaultOrganizationArgs{
			OrgName: pulumi.String(serviceOrg),
		}); err != nil {
			return err
		}

		if _, err := auth.NewOidcIssuer(ctx, "githubIssuer", &auth.OidcIssuerArgs{
			OrgName:       pulumi.String(serviceOrg),
			Name:          pulumi.String("github_issuer_" + suffix),
			Url:           pulumi.String("https://token.actions.githubusercontent.com"),
			Thumbprints:   pulumi.StringArray{pulumi.String("caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7")},
			MaxExpiration: pulumi.Int(3600),
		}); err != nil {
			return err
		}
		if _, err := auth.NewOidcIssuer(ctx, "pulumiSelfIssuer", &auth.OidcIssuerArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("pulumi_issuer_" + suffix),
			Url:         pulumi.String("https://api.pulumi.com/oidc"),
			Thumbprints: pulumi.StringArray{pulumi.String("57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da")},
		}); err != nil {
			return err
		}

		platformTeam, err := teams.NewTeam(ctx, "platformTeam", &teams.TeamArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("platform-team-" + suffix),
			DisplayName: pulumi.String("Platform Team " + suffix),
			Description: pulumi.String("Owns shared infra, runs the deployments engine."),
		})
		if err != nil {
			return err
		}

		if _, err := v2.NewRole(ctx, "stackReadonlyRole", &v2.RoleArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("stack-readonly-" + suffix),
			Description: pulumi.String("Read-only access to stacks, scoped via the platform team."),
			UxPurpose:   pulumi.String("role"),
			Details: pulumi.Map{
				"__type":      pulumi.String("PermissionDescriptorAllow"),
				"permissions": pulumi.StringArray{pulumi.String("stack:read")},
			},
		}); err != nil {
			return err
		}

		ciToken, err := tokens.NewOrgToken(ctx, "ciToken", &tokens.OrgTokenArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("ci-" + suffix),
			Description: pulumi.String("Used by CI/CD to deploy non-prod stacks."),
			Admin:       pulumi.Bool(false),
			Expires:     pulumi.Int(0),
		})
		if err != nil {
			return err
		}
		if _, err := tokens.NewTeamToken(ctx, "teamToken", &tokens.TeamTokenArgs{
			OrgName:     pulumi.String(serviceOrg),
			TeamName:    platformTeam.Name,
			Name:        pulumi.String("platform-team-token-" + suffix),
			Description: pulumi.String("Platform-team-scoped token for shared automation."),
			Expires:     pulumi.Int(0),
		}); err != nil {
			return err
		}

		runnersPool, err := agents.NewPool(ctx, "runnersPool", &agents.PoolArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("platform-runners-" + suffix),
			Description: pulumi.String("Self-hosted deployment runner pool."),
		})
		if err != nil {
			return err
		}

		templates, err := v2.NewOrgTemplateCollection(ctx, "templates", &v2.OrgTemplateCollectionArgs{
			OrgName:   pulumi.String(serviceOrg),
			Name:      pulumi.String("platform-templates-" + suffix),
			SourceURL: pulumi.String("https://github.com/pulumi/examples"),
		})
		if err != nil {
			return err
		}

		sharedCredentials, err := esc.NewEnvironment(ctx, "sharedCredentials", &esc.EnvironmentArgs{
			OrgName: pulumi.String(serviceOrg),
			Project: pulumi.String("shared"),
			Name:    pulumi.String("credentials-" + suffix),
		})
		if err != nil {
			return err
		}
		if _, err := esc.NewEnvironmentTag(ctx, "stableTag", &esc.EnvironmentTagArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: sharedCredentials.Project,
			EnvName:     sharedCredentials.Name,
			Name:        pulumi.String("stable"),
			Value:       pulumi.String("1"),
		}); err != nil {
			return err
		}

		stagingStack, err := stacks.NewStack(ctx, "stagingStack", &stacks.StackArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: pulumi.String("platform-app-" + suffix),
			StackName:   pulumi.String("staging"),
		})
		if err != nil {
			return err
		}
		prodStack, err := stacks.NewStack(ctx, "prodStack", &stacks.StackArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: pulumi.String("platform-app-" + suffix),
			StackName:   pulumi.String("prod"),
		})
		if err != nil {
			return err
		}

		sharedEnvRef := pulumi.Sprintf("%s/%s", sharedCredentials.Project, sharedCredentials.Name)

		if _, err := stacks.NewConfig(ctx, "stagingConfig", &stacks.ConfigArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: stagingStack.ProjectName,
			StackName:   stagingStack.StackName,
			Environment: sharedEnvRef,
		}); err != nil {
			return err
		}
		if _, err := stacks.NewConfig(ctx, "prodConfig", &stacks.ConfigArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: prodStack.ProjectName,
			StackName:   prodStack.StackName,
			Environment: sharedEnvRef,
		}); err != nil {
			return err
		}

		for _, kv := range []struct{ k, v string }{
			{"owner", "platform-team"},
			{"tier", "gold"},
			{"cost-center", "platform"},
		} {
			if _, err := stacks.NewTag(ctx, "prodTag-"+kv.k, &stacks.TagArgs{
				OrgName:     pulumi.String(serviceOrg),
				ProjectName: prodStack.ProjectName,
				StackName:   prodStack.StackName,
				Name:        pulumi.String(kv.k),
				Value:       pulumi.String(kv.v),
			}); err != nil {
				return err
			}
		}

		if _, err := stacks.NewWebhook(ctx, "prodPagerDuty", &stacks.WebhookArgs{
			OrganizationName: pulumi.String(serviceOrg),
			ProjectName:      prodStack.ProjectName,
			StackName:        prodStack.StackName,
			Name:             pulumi.String("prod-pagerduty"),
			DisplayName:      pulumi.String("prod stack PagerDuty"),
			PayloadUrl:       pulumi.String(pagerDutyWebhookUrl),
			Active:           pulumi.Bool(true),
			Format:           pulumi.String("raw"),
		}); err != nil {
			return err
		}

		if _, err := deployments.NewSettings(ctx, "stagingDeploySettings", &deployments.SettingsArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: stagingStack.ProjectName,
			StackName:   stagingStack.StackName,
			ExecutorContext: pulumi.Map{
				"executorImage": pulumi.Map{"reference": pulumi.String("pulumi/pulumi:latest")},
			},
		}); err != nil {
			return err
		}
		prodDeploySettings, err := deployments.NewSettings(ctx, "prodDeploySettings", &deployments.SettingsArgs{
			OrgName:     pulumi.String(serviceOrg),
			ProjectName: prodStack.ProjectName,
			StackName:   prodStack.StackName,
			ExecutorContext: pulumi.Map{
				"executorImage": pulumi.Map{"reference": pulumi.String("pulumi/pulumi:3-nonroot")},
			},
		})
		if err != nil {
			return err
		}

		if _, err := v2.NewGate(ctx, "credsApprovalGate", &v2.GateArgs{
			OrgName: pulumi.String(serviceOrg),
			Name:    pulumi.String("creds-approval-" + suffix),
			Enabled: pulumi.Bool(prodApprovalEnabled),
			Rule: pulumi.Map{
				"ruleType":                  pulumi.String("approval_required"),
				"numApprovalsRequired":      pulumi.Int(1),
				"allowSelfApproval":         pulumi.Bool(false),
				"requireReapprovalOnChange": pulumi.Bool(true),
				"eligibleApprovers": pulumi.Array{
					pulumi.Map{
						"eligibilityType": pulumi.String("team_member"),
						"teamName":        platformTeam.Name,
					},
				},
			},
			Target: pulumi.Map{
				"entityType":    pulumi.String("environment"),
				"actionTypes":   pulumi.StringArray{pulumi.String("update")},
				"qualifiedName": sharedEnvRef,
			},
		}); err != nil {
			return err
		}

		if _, err := deployments.NewScheduledDeployment(ctx, "prodNightlyDeploy", &deployments.ScheduledDeploymentArgs{
			OrgName:      pulumi.String(serviceOrg),
			ProjectName:  prodStack.ProjectName,
			StackName:    prodStack.StackName,
			ScheduleCron: pulumi.String("0 7 * * *"),
			Request: pulumi.Map{
				"operation":       pulumi.String("update"),
				"inheritSettings": pulumi.Bool(true),
			},
		}, pulumi.DependsOn([]pulumi.Resource{prodDeploySettings})); err != nil {
			return err
		}

		if _, err := v2.NewOrganizationWebhook(ctx, "slack", &v2.OrganizationWebhookArgs{
			OrganizationName: pulumi.String(serviceOrg),
			Name:             pulumi.String("org-slack-" + suffix),
			DisplayName:      pulumi.String("Org Slack notifications"),
			PayloadUrl:       pulumi.String(slackWebhookUrl),
			Active:           pulumi.Bool(true),
			Format:           pulumi.String("slack"),
		}); err != nil {
			return err
		}

		if _, err := v2.NewPolicyGroup(ctx, "starterPolicyGroup", &v2.PolicyGroupArgs{
			OrgName:    pulumi.String(serviceOrg),
			Name:       pulumi.String("platform-policies-" + suffix),
			EntityType: pulumi.String("stacks"),
		}); err != nil {
			return err
		}

		ctx.Export("platformTeamName", platformTeam.Name)
		ctx.Export("ciTokenId", ciToken.TokenId)
		ctx.Export("agentPoolName", runnersPool.Name)
		ctx.Export("templatesName", templates.Name)
		ctx.Export("sharedCredsEnv", sharedEnvRef)
		ctx.Export("prodStackId", prodStack.ID())
		return nil
	})
}
