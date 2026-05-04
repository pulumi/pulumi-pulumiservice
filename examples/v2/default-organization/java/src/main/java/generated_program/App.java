package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.DefaultOrganization;
import com.pulumi.pulumiservice.v2.DefaultOrganizationArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");

            var def = new DefaultOrganization("default",
                DefaultOrganizationArgs.builder()
                    .orgName(serviceOrg)
                    .build());

            ctx.export("defaultOrg", def.orgName());
        });
    }
}
