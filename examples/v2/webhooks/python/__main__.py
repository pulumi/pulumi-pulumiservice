import pulumi
import pulumi_pulumiservice as pulumiservice

config = pulumi.Config()
service_org = config.get("serviceOrg")
if service_org is None:
    service_org = "service-provider-test-org"
secret_value = config.get("secretValue")
if secret_value is None:
    secret_value = "shhh"
# Organization-scoped webhook subscribed to all events.
org_webhook_all = pulumiservice.v2.OrganizationWebhook("orgWebhookAll",
    org_name=service_org,
    organization_name=service_org,
    name=org-webhook-all,
    display_name=webhook-from-provider,
    payload_url=https://google.com,
    active=True,
    secret=secret_value)
# Organization-scoped webhook subscribed only to environments and stacks groups.
org_webhook_groups = pulumiservice.v2.OrganizationWebhook("orgWebhookGroups",
    org_name=service_org,
    organization_name=service_org,
    name=org-webhook-groups,
    display_name=webhook-from-provider,
    payload_url=https://google.com,
    active=True,
    groups=[
        environments,
        stacks,
    ],
    secret=secret_value)
pulumi.export("orgWebhookId", org_webhook_all["id"])
