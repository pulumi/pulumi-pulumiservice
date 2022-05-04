import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const team = new service.Team("team", {
    description: "This was created with Pulumi",
    name: "brand-new-ts-team",
    displayName: "PulumiUP Team",
    organizationName: "service-provider-test-org",
    members: ["pulumi-bot", "service-provider-example-user"],
    teamType: "pulumi"
});