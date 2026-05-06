package main

import (
	teams "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/teams"
	tokens "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2/tokens"
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
		tokenSuffix := cfg.Get("tokenSuffix")
		if tokenSuffix == "" {
			tokenSuffix = "dev"
		}
		tokenDescription := cfg.Get("tokenDescription")
		if tokenDescription == "" {
			tokenDescription = "example v2 access token"
		}

		team, err := teams.NewTeam(ctx, "team", &teams.TeamArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-tokens-team-" + tokenSuffix),
			DisplayName: pulumi.String("v2 Tokens Team " + tokenSuffix),
			Description: pulumi.String("Owner team for the v2 access-tokens example"),
		})
		if err != nil {
			return err
		}

		orgToken, err := tokens.NewOrgToken(ctx, "orgToken", &tokens.OrgTokenArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-org-token-" + tokenSuffix),
			Description: pulumi.String(tokenDescription),
			Admin:       pulumi.Bool(false),
			Expires:     pulumi.Int(0),
		})
		if err != nil {
			return err
		}

		teamToken, err := tokens.NewTeamToken(ctx, "teamToken", &tokens.TeamTokenArgs{
			OrgName:     pulumi.String(serviceOrg),
			TeamName:    team.Name,
			Name:        pulumi.String("v2-team-token-" + tokenSuffix),
			Description: pulumi.String(tokenDescription),
			Expires:     pulumi.Int(0),
		})
		if err != nil {
			return err
		}

		_, err = tokens.NewPersonalToken(ctx, "personalToken", &tokens.PersonalTokenArgs{
			Description: pulumi.String(tokenDescription),
			Expires:     pulumi.Int(0),
		})
		if err != nil {
			return err
		}

		ctx.Export("orgTokenId", orgToken.TokenId)
		ctx.Export("teamTokenId", teamToken.TokenId)
		return nil
	})
}
