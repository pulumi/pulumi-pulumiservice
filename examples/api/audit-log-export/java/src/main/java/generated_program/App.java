package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.AuditLogExportConfiguration;
import com.pulumi.pulumiservice.api.AuditLogExportConfigurationArgs;
import com.pulumi.pulumiservice.api.inputs.AuditLogsExportS3ConfigArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var bucketName = config.get("bucketName").orElse("pulumi-audit-log-archive");

            var exportConfig = new AuditLogExportConfiguration("exportConfig",
                AuditLogExportConfigurationArgs.builder()
                    .orgName(organizationName)
                    .newEnabled(true)
                    .newS3Configuration(AuditLogsExportS3ConfigArgs.builder()
                        .s3BucketName(bucketName)
                        .iamRoleArn("arn:aws:iam::123456789012:role/PulumiAuditLogExportRole")
                        .build())
                    .build());

            ctx.export("exportEnabled", exportConfig.enabled());
        });
    }
}
