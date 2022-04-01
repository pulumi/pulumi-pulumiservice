import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const team = new service.Team("team", {
    description: "This was created with Pulumi",
    displayName: "Team Awesome",
    members: ["piers3"],
    name: "providerteam",
    organisationName: "pk-demo",
    type: "pulumi"
});