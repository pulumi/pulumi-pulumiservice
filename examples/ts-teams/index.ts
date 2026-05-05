import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const nameSuffix = config.get("nameSuffix") ?? "";
const teamName = nameSuffix ? `brand-new-ts-team-${nameSuffix}` : "brand-new-ts-team";

const team = new service.Team("team", {
    description: "This was created with Pulumi",
    name: teamName,
    displayName: "PulumiUP Team",
    organizationName: process.env.PULUMI_TEST_OWNER || "service-provider-test-org",
    members: ["pulumi-bot", "service-provider-example-user"],
    teamType: "pulumi"
});
