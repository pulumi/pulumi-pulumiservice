import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
team_suffix = config.get("teamSuffix") or "dev"
team_description = config.get("teamDescription") or "A team created by the v2 example."

team = ps_v2.teams.Team(
    "team",
    org_name=service_org,
    name=f"v2-team-{team_suffix}",
    display_name=f"v2 Team {team_suffix}",
    description=team_description,
)

pulumi.export("teamName", team.name)
