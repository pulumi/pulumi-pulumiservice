import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const tokenSuffix = config.get("tokenSuffix") ?? "dev";
const tokenDescription = config.get("tokenDescription") ?? "example v1 access token";

const team = new ps.v1.teams.Team("team", {
    orgName: organizationName,
    name: `v1-tokens-team-${tokenSuffix}`,
    displayName: `v1 Tokens Team ${tokenSuffix}`,
    description: "Owner team for the v1 access-tokens example",
});

const orgToken = new ps.v1.tokens.OrgToken("orgToken", {
    orgName: organizationName,
    name: `v1-org-token-${tokenSuffix}`,
    description: tokenDescription,
    admin: false,
    expires: 0,
});

const teamToken = new ps.v1.tokens.TeamToken("teamToken", {
    orgName: organizationName,
    teamName: team.name,
    name: `v1-team-token-${tokenSuffix}`,
    description: tokenDescription,
    expires: 0,
});

new ps.v1.tokens.PersonalToken("personalToken", {
    description: tokenDescription,
    expires: 0,
});

export const orgTokenId = orgToken.tokenId;
export const teamTokenId = teamToken.tokenId;
