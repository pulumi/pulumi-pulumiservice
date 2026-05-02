# Webhook examples ported from examples/ts-webhooks to the v2 namespace.
#
# v1 had a single `pulumiservice:Webhook` whose scope was driven by which
# combination of organizationName / projectName / stackName / envName fields
# you set. v2 splits these into separate resource types so the scope is
# explicit at the type level.
#
# Convert this file to a runnable language:
#   pulumi convert --from pcl --language typescript --out ../../v2-ts-webhooks
#   pulumi convert --from pcl --language python     --out ../../v2-py-webhooks
#   pulumi convert --from pcl --language yaml       --out ../../v2-yaml-webhooks

config serviceOrg "string" {
  default = "service-provider-test-org"
}
config secretValue "string" {
  default = "shhh"
}

# Organization-scoped webhook subscribed to all events.
resource orgWebhookAll "pulumiservice:v2:OrganizationWebhook" {
  orgName          = serviceOrg
  organizationName = serviceOrg
  name             = "org-webhook-all"
  displayName      = "webhook-from-provider"
  payloadUrl       = "https://google.com"
  active           = true
  secret           = secretValue
}

# Organization-scoped webhook subscribed only to environments and stacks groups.
resource orgWebhookGroups "pulumiservice:v2:OrganizationWebhook" {
  orgName          = serviceOrg
  organizationName = serviceOrg
  name             = "org-webhook-groups"
  displayName      = "webhook-from-provider"
  payloadUrl       = "https://google.com"
  active           = true
  groups           = ["environments", "stacks"]
  secret           = secretValue
}

output orgWebhookId {
  value = orgWebhookAll.id
}
