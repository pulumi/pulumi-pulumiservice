// Go variant of canonical/07-tiered-team-access.
// Three teams × two stacks × three permission tiers. Behavioral twin of
// the sibling YAML program.

package main

import (
	"fmt"

	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/orgs/teams"
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/stacks/permissions"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Team-stack permission levels: 0=none, 101=read, 102=edit, 103=admin.
const (
	permRead  = 101
	permEdit  = 102
	permAdmin = 103
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		org := cfg.Get("organizationName")
		if org == "" {
			org = "service-provider-test-org"
		}
		digits := cfg.Get("digits")
		if digits == "" {
			digits = "00000"
		}

		platformAdmins, err := teams.NewTeam(ctx, "platformAdmins", &teams.TeamArgs{
			OrganizationName: pulumi.String(org),
			TeamType:         pulumi.String("pulumi"),
			Name:             pulumi.String(fmt.Sprintf("platform-admins-%s", digits)),
			DisplayName:      pulumi.String("Platform Admins"),
			Description:      pulumi.String("Break-glass access to everything. Keep small."),
		})
		if err != nil {
			return err
		}
		billingOwners, err := teams.NewTeam(ctx, "billingOwners", &teams.TeamArgs{
			OrganizationName: pulumi.String(org),
			TeamType:         pulumi.String("pulumi"),
			Name:             pulumi.String(fmt.Sprintf("billing-owners-%s", digits)),
			DisplayName:      pulumi.String("Billing Service Owners"),
			Description:      pulumi.String("Owns the billing service stacks end-to-end."),
		})
		if err != nil {
			return err
		}
		developers, err := teams.NewTeam(ctx, "developers", &teams.TeamArgs{
			OrganizationName: pulumi.String(org),
			TeamType:         pulumi.String("pulumi"),
			Name:             pulumi.String(fmt.Sprintf("developers-%s", digits)),
			DisplayName:      pulumi.String("Developers (all)"),
			Description:      pulumi.String("Read everything; deploy nothing without an explicit grant."),
		})
		if err != nil {
			return err
		}

		grant := func(name, project string, team pulumi.StringOutput, perm int) error {
			_, err := permissions.NewTeamStackPermission(ctx, name, &permissions.TeamStackPermissionArgs{
				Organization: pulumi.String(org),
				Project:      pulumi.String(project),
				Stack:        pulumi.String("prod"),
				Team:         team,
				Permission:   pulumi.Int(perm),
			})
			return err
		}

		for _, g := range []struct {
			name, project string
			team          pulumi.StringOutput
			perm          int
		}{
			{"platformAdminPerm", "platform", platformAdmins.Name.Elem(), permAdmin},
			{"platformDevRead", "platform", developers.Name.Elem(), permRead},
			{"billingAdminPerm", "billing-service", platformAdmins.Name.Elem(), permAdmin},
			{"billingOwnerPerm", "billing-service", billingOwners.Name.Elem(), permEdit},
			{"billingDevRead", "billing-service", developers.Name.Elem(), permRead},
		} {
			if err := grant(g.name, g.project, g.team, g.perm); err != nil {
				return err
			}
		}
		return nil
	})
}
