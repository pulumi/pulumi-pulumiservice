import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const accountSuffix = config.get("accountSuffix") ?? "dev";
const insightsEnvironment = config.get("insightsEnvironment") ?? "insights/credentials";

const account = new ps.v2.Account("account", {
    orgName: serviceOrg,
    accountName: `v2-insights-${accountSuffix}`,
    provider: "aws",
    environment: insightsEnvironment,
    scanSchedule: "none",
});

new ps.v2.ScheduledScanSettings("scanSettings", {
    orgName: serviceOrg,
    accountName: account.accountName,
    paused: true,
    scheduleCron: "0 6 * * *",
});

export const accountName = account.accountName;
