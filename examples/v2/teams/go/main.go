package main

import (
	v2 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2"
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
		teamSuffix := cfg.Get("teamSuffix")
		if teamSuffix == "" {
			teamSuffix = "dev"
		}
		teamDescription := cfg.Get("teamDescription")
		if teamDescription == "" {
			teamDescription = "A team created by the v2 example."
		}

		team, err := v2.NewTeam(ctx, "team", &v2.TeamArgs{
			OrgName:     pulumi.String(serviceOrg),
			Name:        pulumi.String("v2-team-" + teamSuffix),
			DisplayName: pulumi.String("v2 Team " + teamSuffix),
			Description: pulumi.String(teamDescription),
		})
		if err != nil {
			return err
		}

		ctx.Export("teamName", team.Name)
		return nil
	})
}
