import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const accesstokenstack = new pulumi.StackReference("pierskarsenbarg/pulumi-service-access-tokens-example-ts/dev");
const accessToken = accesstokenstack.getOutput("token");

const provider = new service.Provider("provider", {
    accessToken: accessToken
});

const webhook = new service.Webhook("wh", {
    active: true,
    displayName: "webhook-from-provider",
    organizationName: "pk-demo",
    payloadUrl: "https://google.com",
}, {provider})

export const orgname = webhook.organizationName;
export const name = webhook.name;