import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
token_suffix = config.get("tokenSuffix") or "dev"
token_description = config.get("tokenDescription") or "example v1 access token"

team = ps_v1.teams.Team(
    "team",
    org_name=organization_name,
    name=f"v1-tokens-team-{token_suffix}",
    display_name=f"v1 Tokens Team {token_suffix}",
    description="Owner team for the v1 access-tokens example",
)

org_token = ps_v1.tokens.OrgToken(
    "orgToken",
    org_name=organization_name,
    name=f"v1-org-token-{token_suffix}",
    description=token_description,
    admin=False,
    expires=0,
)

team_token = ps_v1.tokens.TeamToken(
    "teamToken",
    org_name=organization_name,
    team_name=team.name,
    name=f"v1-team-token-{token_suffix}",
    description=token_description,
    expires=0,
)

ps_v1.tokens.PersonalToken(
    "personalToken",
    description=token_description,
    expires=0,
)

pulumi.export("orgTokenId", org_token.token_id)
pulumi.export("teamTokenId", team_token.token_id)
