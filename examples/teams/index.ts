import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const accesstokenstack = new pulumi.StackReference("pierskarsenbarg/pulumi-service-access-tokens-example-ts/dev");
const accessToken = accesstokenstack.getOutput("token");



const provider = new service.Provider("provider", {
    accessToken: accessToken
});

const team = new service.Team("team", {
    description: "This was created with Pulumi",
    displayName: "Team Awesome",
    members: ["piers3"],
    name: "providerteam",
    organisationName: "pk-demo",
    type: "pulumi"
}, { provider });