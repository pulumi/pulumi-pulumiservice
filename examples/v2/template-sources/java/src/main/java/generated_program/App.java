package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.OrgTemplateCollection;
import com.pulumi.pulumiservice.v2.OrgTemplateCollectionArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var templateSuffix = config.get("templateSuffix").orElse("dev");
            var sourceUrl = config.get("sourceUrl").orElse("https://github.com/pulumi/examples");

            var source = new OrgTemplateCollection("source",
                OrgTemplateCollectionArgs.builder()
                    .orgName(organizationName)
                    .name("v2-templates-" + templateSuffix)
                    .sourceURL(sourceUrl)
                    .build());

            ctx.export("collectionName", source.name());
        });
    }
}
