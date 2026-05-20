package main

import (
	v1 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1"
	teams "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/teams"
	stacks "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/stacks"
	auth "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/auth"
	agents "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/agents"
	tokens "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/tokens"
	deployments "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/deployments"
	esc "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v1/esc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		organizationName := cfg.Get("organizationName")
		if organizationName == "" {
			organizationName = "service-provider-test-org"
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

		if _, err := v1.NewDefaultOrganization(ctx, "defaultOrg", &v1.DefaultOrganizationArgs{
			OrgName: pulumi.String(organizationName),
		}); err != nil {
			return err
		}

		if _, err := auth.NewOidcIssuer(ctx, "githubIssuer", &auth.OidcIssuerArgs{
			OrgName:       pulumi.String(organizationName),
			Name:          pulumi.String("github_issuer_" + suffix),
			Url:           pulumi.String("https://token.actions.githubusercontent.com"),
			Thumbprints:   pulumi.StringArray{pulumi.String("b41ae0832808ebc94951437bf7e92b93ccb6479364daf894d46d6001bee7a486")},
			MaxExpiration: pulumi.Int(3600),
		}); err != nil {
			return err
		}
		if _, err := auth.NewOidcIssuer(ctx, "pulumiSelfIssuer", &auth.OidcIssuerArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("pulumi_issuer_" + suffix),
			Url:         pulumi.String("https://api.pulumi.com/oidc"),
			Thumbprints: pulumi.StringArray{pulumi.String("57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da")},
		}); err != nil {
			return err
		}

		platformTeam, err := teams.NewTeam(ctx, "platformTeam", &teams.TeamArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("platform-team-" + suffix),
			DisplayName: pulumi.String("Platform Team " + suffix),
			Description: pulumi.String("Owns shared infra, runs the deployments engine."),
		})
		if err != nil {
			return err
		}

		if _, err := v1.NewRole(ctx, "stackReadonlyRole", &v1.RoleArgs{
			OrgName:     pulumi.String(organizationName),
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
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("ci-" + suffix),
			Description: pulumi.String("Used by CI/CD to deploy non-prod stacks."),
			Admin:       pulumi.Bool(false),
			Expires:     pulumi.Int(0),
		})
		if err != nil {
			return err
		}
		if _, err := tokens.NewTeamToken(ctx, "teamToken", &tokens.TeamTokenArgs{
			OrgName:     pulumi.String(organizationName),
			TeamName:    platformTeam.Name,
			Name:        pulumi.String("platform-team-token-" + suffix),
			Description: pulumi.String("Platform-team-scoped token for shared automation."),
			Expires:     pulumi.Int(0),
		}); err != nil {
			return err
		}

		runnersPool, err := agents.NewPool(ctx, "runnersPool", &agents.PoolArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("platform-runners-" + suffix),
			Description: pulumi.String("Self-hosted deployment runner pool."),
		})
		if err != nil {
			return err
		}

		templates, err := v1.NewOrgTemplateCollection(ctx, "templates", &v1.OrgTemplateCollectionArgs{
			OrgName:   pulumi.String(organizationName),
			Name:      pulumi.String("platform-templates-" + suffix),
			SourceURL: pulumi.String("https://github.com/pulumi/examples"),
		})
		if err != nil {
			return err
		}

		sharedCredentials, err := esc.NewEnvironment(ctx, "sharedCredentials", &esc.EnvironmentArgs{
			OrgName: pulumi.String(organizationName),
			Project: pulumi.String("shared"),
			Name:    pulumi.String("credentials-" + suffix),
		})
		if err != nil {
			return err
		}
		if _, err := esc.NewEnvironmentTag(ctx, "stableTag", &esc.EnvironmentTagArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: sharedCredentials.Project,
			EnvName:     sharedCredentials.Name,
			Name:        pulumi.String("stable"),
			Value:       pulumi.String("1"),
		}); err != nil {
			return err
		}

		stagingStack, err := stacks.NewStack(ctx, "stagingStack", &stacks.StackArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String("platform-app-" + suffix),
			StackName:   pulumi.String("staging"),
		})
		if err != nil {
			return err
		}
		prodStack, err := stacks.NewStack(ctx, "prodStack", &stacks.StackArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: pulumi.String("platform-app-" + suffix),
			StackName:   pulumi.String("prod"),
		})
		if err != nil {
			return err
		}

		sharedEnvRef := pulumi.Sprintf("%s/%s", sharedCredentials.Project, sharedCredentials.Name)

		if _, err := stacks.NewConfig(ctx, "stagingConfig", &stacks.ConfigArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: stagingStack.ProjectName,
			StackName:   stagingStack.StackName,
			Environment: sharedEnvRef,
		}); err != nil {
			return err
		}
		if _, err := stacks.NewConfig(ctx, "prodConfig", &stacks.ConfigArgs{
			OrgName:     pulumi.String(organizationName),
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
				OrgName:     pulumi.String(organizationName),
				ProjectName: prodStack.ProjectName,
				StackName:   prodStack.StackName,
				Name:        pulumi.String(kv.k),
				Value:       pulumi.String(kv.v),
			}); err != nil {
				return err
			}
		}

		if _, err := stacks.NewWebhook(ctx, "prodPagerDuty", &stacks.WebhookArgs{
			OrganizationName: pulumi.String(organizationName),
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
			OrgName:     pulumi.String(organizationName),
			ProjectName: stagingStack.ProjectName,
			StackName:   stagingStack.StackName,
			ExecutorContext: pulumi.Map{
				"executorImage": pulumi.Map{"reference": pulumi.String("pulumi/pulumi:latest")},
			},
		}); err != nil {
			return err
		}
		prodDeploySettings, err := deployments.NewSettings(ctx, "prodDeploySettings", &deployments.SettingsArgs{
			OrgName:     pulumi.String(organizationName),
			ProjectName: prodStack.ProjectName,
			StackName:   prodStack.StackName,
			ExecutorContext: pulumi.Map{
				"executorImage": pulumi.Map{"reference": pulumi.String("pulumi/pulumi:3-nonroot")},
			},
		})
		if err != nil {
			return err
		}

		if _, err := v1.NewGate(ctx, "credsApprovalGate", &v1.GateArgs{
			OrgName: pulumi.String(organizationName),
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
			OrgName:      pulumi.String(organizationName),
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

		if _, err := v1.NewOrganizationWebhook(ctx, "slack", &v1.OrganizationWebhookArgs{
			OrganizationName: pulumi.String(organizationName),
			Name:             pulumi.String("org-slack-" + suffix),
			DisplayName:      pulumi.String("Org Slack notifications"),
			PayloadUrl:       pulumi.String(slackWebhookUrl),
			Active:           pulumi.Bool(true),
			Format:           pulumi.String("slack"),
		}); err != nil {
			return err
		}

		if _, err := v1.NewPolicyGroup(ctx, "starterPolicyGroup", &v1.PolicyGroupArgs{
			OrgName:    pulumi.String(organizationName),
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
