import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
team_suffix = config.get("teamSuffix") or "dev"
team_description = config.get("teamDescription") or "A team created by the api example."

team = ps_api.teams.Team(
    "team",
    org_name=organization_name,
    name=f"api-team-{team_suffix}",
    display_name=f"api Team {team_suffix}",
    description=team_description,
)

pulumi.export("teamName", team.name)
