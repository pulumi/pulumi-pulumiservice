using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var bucketName = config.Get("bucketName") ?? "pulumi-audit-log-archive";

    var exportConfig = new Ps.Api.AuditLogExportConfiguration("exportConfig", new()
    {
        OrgName = organizationName,
        NewEnabled = true,
        NewS3Configuration = new Ps.Api.Inputs.AuditLogsExportS3ConfigArgs
        {
            S3BucketName = bucketName,
            IamRoleArn = "arn:aws:iam::123456789012:role/PulumiAuditLogExportRole",
        },
    });

    return new Dictionary<string, object?>
    {
        ["exportEnabled"] = exportConfig.Enabled,
    };
});
