import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
account_suffix = config.get("accountSuffix") or "dev"
insights_environment = config.get("insightsEnvironment") or "insights/credentials"

account = ps_v2.insights.Account(
    "account",
    org_name=service_org,
    account_name=f"v2-insights-{account_suffix}",
    provider="aws",
    environment=insights_environment,
    scan_schedule="none",
)

ps_v2.insights.ScheduledScanSettings(
    "scanSettings",
    org_name=service_org,
    account_name=account.account_name,
    paused=True,
    schedule_cron="0 6 * * *",
)

pulumi.export("accountName", account.account_name)
