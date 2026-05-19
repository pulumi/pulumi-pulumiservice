import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
bucket_name = config.get("bucketName") or "pulumi-audit-log-archive"
region = config.get("region") or "us-west-2"

export_config = ps_v2.AuditLogExportConfiguration(
    "exportConfig",
    org_name=organization_name,
    new_enabled=True,
    new_s3_configuration={
        "bucketName": bucket_name,
        "region": region,
        "roleArn": "arn:aws:iam::123456789012:role/PulumiAuditLogExportRole",
    },
)

pulumi.export("exportEnabled", export_config.enabled)
