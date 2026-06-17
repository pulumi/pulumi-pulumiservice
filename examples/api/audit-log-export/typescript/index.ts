import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const bucketName = config.get("bucketName") ?? "pulumi-audit-log-archive";

const exportConfig = new ps.api.AuditLogExportConfiguration("exportConfig", {
    orgName: organizationName,
    newEnabled: true,
    newS3Configuration: {
        s3BucketName: bucketName,
        iamRoleArn: "arn:aws:iam::123456789012:role/PulumiAuditLogExportRole",
    },
});

export const exportEnabled = exportConfig.enabled;
