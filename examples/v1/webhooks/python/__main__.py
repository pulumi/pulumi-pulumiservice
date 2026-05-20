import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
secret_value = config.get("secretValue") or "shhh"
hook_suffix = config.get("hookSuffix") or "dev"

org_webhook_all = ps_v1.OrganizationWebhook(
    "orgWebhookAll",
    organization_name=organization_name,
    name=f"org-webhook-all-{hook_suffix}",
    display_name="webhook-from-provider",
    payload_url="https://google.com",
    active=True,
    secret=secret_value,
)

org_webhook_groups = ps_v1.OrganizationWebhook(
    "orgWebhookGroups",
    organization_name=organization_name,
    name=f"org-webhook-groups-{hook_suffix}",
    display_name="webhook-from-provider",
    payload_url="https://google.com",
    active=True,
    groups=["environments", "stacks"],
    secret=secret_value,
)

pulumi.export("orgWebhookId", org_webhook_all.id)
pulumi.export("orgWebhookGroupsId", org_webhook_groups.id)
