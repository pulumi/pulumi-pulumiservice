import * as service from "@pulumi/pulumiservice";

const team = new service.Team("team", {
    description: "test from provider",
    displayName: "my new team",
    members: ["piers3"],
    name: "providerteam",
    organisationName: "pk-demo",
    type: "pulumi"
})