import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const tokenSuffix = config.get("tokenSuffix") ?? "dev";
const tokenDescription = config.get("tokenDescription") ?? "example api access token";

const team = new ps.api.teams.Team("team", {
    orgName: organizationName,
    name: `api-tokens-team-${tokenSuffix}`,
    displayName: `api Tokens Team ${tokenSuffix}`,
    description: "Owner team for the api access-tokens example",
});

const orgToken = new ps.api.tokens.OrgToken("orgToken", {
    orgName: organizationName,
    name: `api-org-token-${tokenSuffix}`,
    description: tokenDescription,
    admin: false,
    expires: 0,
});

const teamToken = new ps.api.tokens.TeamToken("teamToken", {
    orgName: organizationName,
    teamName: team.name,
    name: `api-team-token-${tokenSuffix}`,
    description: tokenDescription,
    expires: 0,
});

new ps.api.tokens.PersonalToken("personalToken", {
    description: tokenDescription,
    expires: 0,
});

export const orgTokenId = orgToken.tokenId;
export const teamTokenId = teamToken.tokenId;
