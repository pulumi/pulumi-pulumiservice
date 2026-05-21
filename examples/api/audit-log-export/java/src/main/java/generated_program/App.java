package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.AuditLogExportConfiguration;
import com.pulumi.pulumiservice.api.AuditLogExportConfigurationArgs;

import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var bucketName = config.get("bucketName").orElse("pulumi-audit-log-archive");
            var region = config.get("region").orElse("us-west-2");

            var exportConfig = new AuditLogExportConfiguration("exportConfig",
                AuditLogExportConfigurationArgs.builder()
                    .orgName(organizationName)
                    .newEnabled(true)
                    .newS3Configuration(Map.of(
                        "bucketName", bucketName,
                        "region", region,
                        "roleArn", "arn:aws:iam::123456789012:role/PulumiAuditLogExportRole"))
                    .build());

            ctx.export("exportEnabled", exportConfig.enabled());
        });
    }
}
