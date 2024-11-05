import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";
import { WebhookGroup, WebhookFormat, WebhookFilters } from "@pulumi/pulumiservice";

const serviceOrg = "service-provider-test-org";
let config = new pulumi.Config();

var environment = new service.Environment("environment-to-use", {
  organization: serviceOrg,
  project: "test-project",
  name: "testing-environment-ts-"+config.require("digits"),
  yaml: new pulumi.asset.StringAsset(
`values:
  myNumber: 1`
  )
})

// Organization webhook subscribed to all events
const webhookAllEvents = new service.Webhook("org-webhook-all", {
  active: true,
  displayName: "webhook-from-provider",
  organizationName: serviceOrg,
  payloadUrl: "https://google.com",
});

// Organization webhook only subscribed to environments and stacks groups
const webhook = new service.Webhook("org-webhook-groups", {
  active: true,
  displayName: "webhook-from-provider",
  organizationName: serviceOrg,
  payloadUrl: "https://google.com",
  groups: [ WebhookGroup.Environments, WebhookGroup.Stacks ]
});

// Stack webhook subscribed to a group and specific filters
const stackWebhook = new service.Webhook("stack-webhook", {
  active: true,
  displayName: "stack-webhook",
  organizationName: serviceOrg,
  projectName: pulumi.getProject(),
  stackName: pulumi.getStack(),
  payloadUrl: "https://hooks.slack.com/blahblah",
  format: WebhookFormat.Slack,
  groups: [ WebhookGroup.Stacks ],
  filters: [WebhookFilters.DeploymentStarted, WebhookFilters.DeploymentSucceeded],
})

// Environment webhook subscribed to specific filters only
const environmentWebhook = new service.Webhook("env-webhook", {
  active: true,
  displayName: "env-webhook",
  organizationName: serviceOrg,
  projectName: environment.project,
  environmentName: environment.name,
  payloadUrl: "https://example.com",
  filters: [WebhookFilters.EnvironmentRevisionCreated, WebhookFilters.ImportedEnvironmentChanged],
})
