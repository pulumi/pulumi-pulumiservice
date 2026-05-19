import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const accountSuffix = config.get("accountSuffix") ?? "dev";
const insightsEnvironment = config.get("insightsEnvironment") ?? "insights/credentials";

const accountNameValue = `v2-insights-${accountSuffix}`;
const account = new ps.v2.insights.Account("account", {
    orgName: organizationName,
    accountName: accountNameValue,
    provider: "aws",
    environment: insightsEnvironment,
    scanSchedule: "none",
});

// accountName is an input (program-owned); reuse the source value.
new ps.v2.insights.ScheduledScanSettings("scanSettings", {
    orgName: organizationName,
    accountName: accountNameValue,
    paused: true,
    scheduleCron: "0 6 * * *",
}, { dependsOn: [account] });

export const accountName = account.name;
