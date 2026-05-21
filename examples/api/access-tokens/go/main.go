package main

import (
	teams "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/teams"
	tokens "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/tokens"
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
		tokenSuffix := cfg.Get("tokenSuffix")
		if tokenSuffix == "" {
			tokenSuffix = "dev"
		}
		tokenDescription := cfg.Get("tokenDescription")
		if tokenDescription == "" {
			tokenDescription = "example api access token"
		}

		team, err := teams.NewTeam(ctx, "team", &teams.TeamArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-tokens-team-" + tokenSuffix),
			DisplayName: pulumi.String("api Tokens Team " + tokenSuffix),
			Description: pulumi.String("Owner team for the api access-tokens example"),
		})
		if err != nil {
			return err
		}

		orgToken, err := tokens.NewOrgToken(ctx, "orgToken", &tokens.OrgTokenArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-org-token-" + tokenSuffix),
			Description: pulumi.String(tokenDescription),
			Admin:       pulumi.Bool(false),
			Expires:     pulumi.Int(0),
		})
		if err != nil {
			return err
		}

		teamToken, err := tokens.NewTeamToken(ctx, "teamToken", &tokens.TeamTokenArgs{
			OrgName:     pulumi.String(organizationName),
			TeamName:    team.Name,
			Name:        pulumi.String("api-team-token-" + tokenSuffix),
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
