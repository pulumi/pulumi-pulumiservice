import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
team_suffix = config.get("teamSuffix") or "dev"
team_description = config.get("teamDescription") or "A team created by the v1 example."

team = ps_v1.teams.Team(
    "team",
    org_name=organization_name,
    name=f"v1-team-{team_suffix}",
    display_name=f"v1 Team {team_suffix}",
    description=team_description,
)

pulumi.export("teamName", team.name)
