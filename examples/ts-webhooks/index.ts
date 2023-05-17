import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const serviceOrg = "service-provider-test-org";

const webhook = new service.Webhook("wh", {
  active: true,
  displayName: "webhook-from-provider",
  organizationName: serviceOrg,
  payloadUrl: "https://google.com",
});

const stackWebhook = new service.Webhook("stack-webhook", {
  active: true,
  displayName: "stack-webhook",
  organizationName: serviceOrg,
  projectName: pulumi.getProject(),
  stackName: pulumi.getStack(),
  payloadUrl: "https://example.com",
})

export const orgName = webhook.organizationName;
export const name = webhook.name;
export const stackWebhookName = stackWebhook.name;
export const stackWebhookProjectName = stackWebhook.projectName;
