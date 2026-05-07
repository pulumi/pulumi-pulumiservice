using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var accountSuffix = config.Get("accountSuffix") ?? "dev";
    var insightsEnvironment = config.Get("insightsEnvironment") ?? "insights/credentials";

    var accountNameValue = $"v2-insights-{accountSuffix}";
    var account = new Ps.V2.Insights.Account("account", new()
    {
        OrgName = serviceOrg,
        AccountName = accountNameValue,
        Provider = "aws",
        Environment = insightsEnvironment,
        ScanSchedule = "none",
    });

    new Ps.V2.Insights.ScheduledScanSettings("scanSettings", new()
    {
        OrgName = serviceOrg,
        AccountName = accountNameValue,
        Paused = true,
        ScheduleCron = "0 6 * * *",
    }, new CustomResourceOptions { DependsOn = { account } });

    return new Dictionary<string, object?>
    {
        ["accountName"] = account.Name,
    };
});
