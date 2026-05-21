import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
token_suffix = config.get("tokenSuffix") or "dev"
token_description = config.get("tokenDescription") or "example api access token"

team = ps_api.teams.Team(
    "team",
    org_name=organization_name,
    name=f"api-tokens-team-{token_suffix}",
    display_name=f"api Tokens Team {token_suffix}",
    description="Owner team for the api access-tokens example",
)

org_token = ps_api.tokens.OrgToken(
    "orgToken",
    org_name=organization_name,
    name=f"api-org-token-{token_suffix}",
    description=token_description,
    admin=False,
    expires=0,
)

team_token = ps_api.tokens.TeamToken(
    "teamToken",
    org_name=organization_name,
    team_name=team.name,
    name=f"api-team-token-{token_suffix}",
    description=token_description,
    expires=0,
)

ps_api.tokens.PersonalToken(
    "personalToken",
    description=token_description,
    expires=0,
)

pulumi.export("orgTokenId", org_token.token_id)
pulumi.export("teamTokenId", team_token.token_id)
