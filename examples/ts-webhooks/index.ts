import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const webhook = new service.Webhook("wh", {
  active: true,
  displayName: "webhook-from-provider",
  organizationName: "service-provider-test-org",
  payloadUrl: "https://google.com",
});

export const orgname = webhook.organizationName;
export const name = webhook.name;
