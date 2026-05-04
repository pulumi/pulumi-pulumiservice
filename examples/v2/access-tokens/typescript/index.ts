import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const tokenSuffix = config.get("tokenSuffix") ?? "dev";
const tokenDescription = config.get("tokenDescription") ?? "example v2 access token";

const team = new ps.v2.Team("team", {
    orgName: serviceOrg,
    name: `v2-tokens-team-${tokenSuffix}`,
    displayName: `v2 Tokens Team ${tokenSuffix}`,
    description: "Owner team for the v2 access-tokens example",
});

const orgToken = new ps.v2.OrgToken("orgToken", {
    orgName: serviceOrg,
    name: `v2-org-token-${tokenSuffix}`,
    description: tokenDescription,
    admin: false,
    expires: 0,
});

const teamToken = new ps.v2.TeamToken("teamToken", {
    orgName: serviceOrg,
    teamName: team.name,
    name: `v2-team-token-${tokenSuffix}`,
    description: tokenDescription,
    expires: 0,
});

new ps.v2.PersonalToken("personalToken", {
    description: tokenDescription,
    expires: 0,
});

export const orgTokenId = orgToken.tokenId;
export const teamTokenId = teamToken.tokenId;
