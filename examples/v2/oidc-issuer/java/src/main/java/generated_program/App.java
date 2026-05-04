package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.OidcIssuer;
import com.pulumi.pulumiservice.v2.OidcIssuerArgs;

import java.util.List;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var issuerSuffix = config.get("issuerSuffix").orElse("dev");
            var maxExpiration = config.getInteger("maxExpiration").orElse(3600);

            var pulumiIssuer = new OidcIssuer("pulumiIssuer",
                OidcIssuerArgs.builder()
                    .orgName(serviceOrg)
                    .name("pulumi_issuer_" + issuerSuffix)
                    .url("https://api.pulumi.com/oidc")
                    .thumbprints(List.of("57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"))
                    .build());

            var githubIssuer = new OidcIssuer("githubIssuer",
                OidcIssuerArgs.builder()
                    .orgName(serviceOrg)
                    .name("github_issuer_" + issuerSuffix)
                    .url("https://token.actions.githubusercontent.com")
                    .thumbprints(List.of("caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"))
                    .maxExpiration(maxExpiration)
                    .build());

            ctx.export("pulumiIssuerName", pulumiIssuer.name());
            ctx.export("githubIssuerName", githubIssuer.name());
        });
    }
}
