import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
bucket_name = config.get("bucketName") or "pulumi-audit-log-archive"

export_config = ps_api.AuditLogExportConfiguration(
    "exportConfig",
    org_name=organization_name,
    new_enabled=True,
    new_s3_configuration=ps_api.AuditLogsExportS3ConfigArgs(
        s3_bucket_name=bucket_name,
        iam_role_arn="arn:aws:iam::123456789012:role/PulumiAuditLogExportRole",
    ),
)

pulumi.export("exportEnabled", export_config.enabled)
