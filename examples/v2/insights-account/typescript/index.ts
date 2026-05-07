import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const accountSuffix = config.get("accountSuffix") ?? "dev";
const insightsEnvironment = config.get("insightsEnvironment") ?? "insights/credentials";

const accountNameValue = `v2-insights-${accountSuffix}`;
const account = new ps.v2.insights.Account("account", {
    orgName: serviceOrg,
    accountName: accountNameValue,
    provider: "aws",
    environment: insightsEnvironment,
    scanSchedule: "none",
});

// accountName is an input (program-owned); reuse the source value.
new ps.v2.insights.ScheduledScanSettings("scanSettings", {
    orgName: serviceOrg,
    accountName: accountNameValue,
    paused: true,
    scheduleCron: "0 6 * * *",
}, { dependsOn: [account] });

export const accountName = account.name;
