package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.v2.OrganizationWebhook;
import com.pulumi.pulumiservice.v2.OrganizationWebhookArgs;
import java.util.List;
import java.util.ArrayList;
import java.util.Map;
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    public static void stack(Context ctx) {
        final var config = ctx.config();
        final var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
        final var secretValue = config.get("secretValue").orElse("shhh");
        // Organization-scoped webhook subscribed to all events.
        var orgWebhookAll = new OrganizationWebhook("orgWebhookAll", OrganizationWebhookArgs.builder()
            .orgName(serviceOrg)
            .organizationName(serviceOrg)
            .name("org-webhook-all")
            .displayName("webhook-from-provider")
            .payloadUrl("https://google.com")
            .active(true)
            .secret(secretValue)
            .build());

        // Organization-scoped webhook subscribed only to environments and stacks groups.
        var orgWebhookGroups = new OrganizationWebhook("orgWebhookGroups", OrganizationWebhookArgs.builder()
            .orgName(serviceOrg)
            .organizationName(serviceOrg)
            .name("org-webhook-groups")
            .displayName("webhook-from-provider")
            .payloadUrl("https://google.com")
            .active(true)
            .groups(List.of(            
                "environments",
                "stacks"))
            .secret(secretValue)
            .build());

        ctx.export("orgWebhookId", orgWebhookAll.id());
    }
}
