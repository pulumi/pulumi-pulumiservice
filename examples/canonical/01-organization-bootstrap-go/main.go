// Go variant of canonical/01-organization-bootstrap.
// Day-0 provisioning: three teams, a team-scoped CI token, a baseline
// policy group. Functionally equivalent to the sibling YAML program.

package main

import (
	"fmt"

	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/orgs/policies"
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/orgs/teams"
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/orgs/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
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

		_, err := teams.NewTeam(ctx, "admins", &teams.TeamArgs{
			OrganizationName: pulumi.String(org),
			TeamType:         pulumi.String("pulumi"),
			Name:             pulumi.String(fmt.Sprintf("admins-%s", digits)),
			DisplayName:      pulumi.String("Organization Admins"),
			Description:      pulumi.String("Full org control; rotate this membership quarterly."),
		})
		if err != nil {
			return err
		}

		deployers, err := teams.NewTeam(ctx, "deployers", &teams.TeamArgs{
			OrganizationName: pulumi.String(org),
			TeamType:         pulumi.String("pulumi"),
			Name:             pulumi.String(fmt.Sprintf("deployers-%s", digits)),
			DisplayName:      pulumi.String("CI Deployers"),
			Description:      pulumi.String("Automation-only team. Human members discouraged — use the CI token."),
		})
		if err != nil {
			return err
		}

		_, err = teams.NewTeam(ctx, "readers", &teams.TeamArgs{
			OrganizationName: pulumi.String(org),
			TeamType:         pulumi.String("pulumi"),
			Name:             pulumi.String(fmt.Sprintf("readers-%s", digits)),
			DisplayName:      pulumi.String("Developers (read-only)"),
			Description:      pulumi.String("Default team for new org members; grants stack read access."),
		})
		if err != nil {
			return err
		}

		ciToken, err := tokens.NewTeamAccessToken(ctx, "ciToken", &tokens.TeamAccessTokenArgs{
			OrganizationName: pulumi.String(org),
			TeamName:         deployers.Name.Elem(),
			Description:      pulumi.String("Used by GitHub Actions to deploy non-production stacks."),
		})
		if err != nil {
			return err
		}

		_, err = policies.NewPolicyGroup(ctx, "defaultGuardrails", &policies.PolicyGroupArgs{
			OrganizationName: pulumi.String(org),
			Name:             pulumi.String(fmt.Sprintf("baseline-guardrails-%s", digits)),
		})
		if err != nil {
			return err
		}

		ctx.Export("ciTokenValue", ciToken.Value)
		return nil
	})
}
