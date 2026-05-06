package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.OrganizationWebhook;
import com.pulumi.pulumiservice.v2.OrganizationWebhookArgs;

import java.util.List;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var secretValue = config.get("secretValue").orElse("shhh");
            var hookSuffix = config.get("hookSuffix").orElse("dev");

            var orgWebhookAll = new OrganizationWebhook("orgWebhookAll",
                OrganizationWebhookArgs.builder()
                    .organizationName(serviceOrg)
                    .name("org-webhook-all-" + hookSuffix)
                    .displayName("webhook-from-provider")
                    .payloadUrl("https://google.com")
                    .active(true)
                    .secret(secretValue)
                    .build());

            var orgWebhookGroups = new OrganizationWebhook("orgWebhookGroups",
                OrganizationWebhookArgs.builder()
                    .organizationName(serviceOrg)
                    .name("org-webhook-groups-" + hookSuffix)
                    .displayName("webhook-from-provider")
                    .payloadUrl("https://google.com")
                    .active(true)
                    .groups(List.of("environments", "stacks"))
                    .secret(secretValue)
                    .build());

            ctx.export("orgWebhookId", orgWebhookAll.id());
            ctx.export("orgWebhookGroupsId", orgWebhookGroups.id());
        });
    }
}
