using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var accountSuffix = config.Get("accountSuffix") ?? "dev";
    var insightsEnvironment = config.Get("insightsEnvironment") ?? "insights/credentials";

    var account = new Ps.V2.Account("account", new()
    {
        OrgName = serviceOrg,
        AccountName = $"v2-insights-{accountSuffix}",
        Provider = "aws",
        Environment = insightsEnvironment,
        ScanSchedule = "none",
    });

    new Ps.V2.ScheduledScanSettings("scanSettings", new()
    {
        OrgName = serviceOrg,
        AccountName = account.AccountName,
        Paused = true,
        ScheduleCron = "0 6 * * *",
    });

    return new Dictionary<string, object?>
    {
        ["accountName"] = account.AccountName,
    };
});
