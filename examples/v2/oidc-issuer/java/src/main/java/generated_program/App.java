package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.v2.OidcIssuer;
import com.pulumi.pulumiservice.v2.OidcIssuerArgs;
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
        final var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org2");
        var pulumiIssuer = new OidcIssuer("pulumiIssuer", OidcIssuerArgs.builder()
            .orgName(serviceOrg)
            .name("pulumi_issuer")
            .url("https://api.pulumi.com/oidc")
            .thumbprints(List.of("57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"))
            .build());

        var githubIssuer = new OidcIssuer("githubIssuer", OidcIssuerArgs.builder()
            .orgName(serviceOrg)
            .name("github_issuer")
            .url("https://token.actions.githubusercontent.com")
            .thumbprints(List.of("caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"))
            .maxExpiration(3600)
            .build());

        ctx.export("pulumiIssuerUrl", pulumiIssuer.url());
        ctx.export("githubIssuerUrl", githubIssuer.url());
    }
}
