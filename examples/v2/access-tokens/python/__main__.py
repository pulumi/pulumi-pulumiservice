import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
token_suffix = config.get("tokenSuffix") or "dev"
token_description = config.get("tokenDescription") or "example v2 access token"

team = ps_v2.teams.Team(
    "team",
    org_name=service_org,
    name=f"v2-tokens-team-{token_suffix}",
    display_name=f"v2 Tokens Team {token_suffix}",
    description="Owner team for the v2 access-tokens example",
)

org_token = ps_v2.tokens.OrgToken(
    "orgToken",
    org_name=service_org,
    name=f"v2-org-token-{token_suffix}",
    description=token_description,
    admin=False,
    expires=0,
)

team_token = ps_v2.tokens.TeamToken(
    "teamToken",
    org_name=service_org,
    team_name=team.name,
    name=f"v2-team-token-{token_suffix}",
    description=token_description,
    expires=0,
)

ps_v2.tokens.PersonalToken(
    "personalToken",
    description=token_description,
    expires=0,
)

pulumi.export("orgTokenId", org_token.token_id)
pulumi.export("teamTokenId", team_token.token_id)
