import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const bucketName = config.get("bucketName") ?? "pulumi-audit-log-archive";
const region = config.get("region") ?? "us-west-2";

const exportConfig = new ps.v1.AuditLogExportConfiguration("exportConfig", {
    orgName: organizationName,
    newEnabled: true,
    newS3Configuration: {
        bucketName: bucketName,
        region: region,
        roleArn: "arn:aws:iam::123456789012:role/PulumiAuditLogExportRole",
    },
});

export const exportEnabled = exportConfig.enabled;
