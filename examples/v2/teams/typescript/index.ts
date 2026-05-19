import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const teamSuffix = config.get("teamSuffix") ?? "dev";
const teamDescription = config.get("teamDescription") ?? "A team created by the v2 example.";

const team = new ps.v2.teams.Team("team", {
    orgName: organizationName,
    name: `v2-team-${teamSuffix}`,
    displayName: `v2 Team ${teamSuffix}`,
    description: teamDescription,
});

export const teamName = team.name;
