using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var bucketName = config.Get("bucketName") ?? "pulumi-audit-log-archive";
    var region = config.Get("region") ?? "us-west-2";

    var exportConfig = new Ps.V2.AuditLogExportConfiguration("exportConfig", new()
    {
        OrgName = serviceOrg,
        NewEnabled = true,
        NewS3Configuration = new Dictionary<string, object?>
        {
            ["bucketName"] = bucketName,
            ["region"] = region,
            ["roleArn"] = "arn:aws:iam::123456789012:role/PulumiAuditLogExportRole",
        }.ToImmutableDictionary(),
    });

    return new Dictionary<string, object?>
    {
        ["exportEnabled"] = exportConfig.Enabled,
    };
});
