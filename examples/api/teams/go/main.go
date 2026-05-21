package main

import (
	teams "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api/teams"
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
		teamSuffix := cfg.Get("teamSuffix")
		if teamSuffix == "" {
			teamSuffix = "dev"
		}
		teamDescription := cfg.Get("teamDescription")
		if teamDescription == "" {
			teamDescription = "A team created by the api example."
		}

		team, err := teams.NewTeam(ctx, "team", &teams.TeamArgs{
			OrgName:     pulumi.String(organizationName),
			Name:        pulumi.String("api-team-" + teamSuffix),
			DisplayName: pulumi.String("api Team " + teamSuffix),
			Description: pulumi.String(teamDescription),
		})
		if err != nil {
			return err
		}

		ctx.Export("teamName", team.Name)
		return nil
	})
}
