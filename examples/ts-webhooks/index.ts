import * as pulumi from "@pulumi/pulumi";
import { Webhook, WebhookFormat, WebhookFilters } from "@pulumi/pulumiservice";

const serviceOrg = "service-provider-test-org";

const webhook = new Webhook("wh", {
  active: true,
  displayName: "webhook-from-provider",
  organizationName: serviceOrg,
  payloadUrl: "https://google.com",
  filters: [WebhookFilters.DeploymentStarted, WebhookFilters.DeploymentSucceeded],
});

const stackWebhook = new Webhook("stack-webhook", {
  active: true,
  displayName: "stack-webhook",
  organizationName: serviceOrg,
  projectName: pulumi.getProject(),
  stackName: pulumi.getStack(),
  payloadUrl: "https://example.com",
  format: WebhookFormat.Slack,
})

export const orgName = webhook.organizationName;
export const name = webhook.name;
export const stackWebhookName = stackWebhook.name;
export const stackWebhookProjectName = stackWebhook.projectName;
