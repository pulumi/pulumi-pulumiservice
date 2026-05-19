import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
account_suffix = config.get("accountSuffix") or "dev"
insights_environment = config.get("insightsEnvironment") or "insights/credentials"

account_name_value = f"v2-insights-{account_suffix}"
account = ps_v2.insights.Account(
    "account",
    org_name=organization_name,
    account_name=account_name_value,
    provider="aws",
    environment=insights_environment,
    scan_schedule="none",
)

ps_v2.insights.ScheduledScanSettings(
    "scanSettings",
    org_name=organization_name,
    account_name=account_name_value,
    paused=True,
    schedule_cron="0 6 * * *",
    opts=pulumi.ResourceOptions(depends_on=[account]),
)

pulumi.export("accountName", account.name)
