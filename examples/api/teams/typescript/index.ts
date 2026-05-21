import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const teamSuffix = config.get("teamSuffix") ?? "dev";
const teamDescription = config.get("teamDescription") ?? "A team created by the api example.";

const team = new ps.api.teams.Team("team", {
    orgName: organizationName,
    name: `api-team-${teamSuffix}`,
    displayName: `api Team ${teamSuffix}`,
    description: teamDescription,
});

export const teamName = team.name;
