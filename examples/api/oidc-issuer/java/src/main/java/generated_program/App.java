package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_auth.OidcIssuer;
import com.pulumi.pulumiservice.api_auth.OidcIssuerArgs;

import java.util.List;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var issuerSuffix = config.get("issuerSuffix").orElse("dev");
            var maxExpiration = config.getInteger("maxExpiration").orElse(3600);

            var pulumiIssuer = new OidcIssuer("pulumiIssuer",
                OidcIssuerArgs.builder()
                    .orgName(organizationName)
                    .name("pulumi_issuer_" + issuerSuffix)
                    .url("https://api.pulumi.com/oidc")
                    .thumbprints(List.of("57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"))
                    .build());

            var githubIssuer = new OidcIssuer("githubIssuer",
                OidcIssuerArgs.builder()
                    .orgName(organizationName)
                    .name("github_issuer_" + issuerSuffix)
                    .url("https://token.actions.githubusercontent.com")
                    .thumbprints(List.of("39517789ff0132a9212bafea4dc37401eae58b1bfac9756109d14301c90a6ab5"))
                    .maxExpiration(maxExpiration)
                    .build());

            ctx.export("pulumiIssuerName", pulumiIssuer.name());
            ctx.export("githubIssuerName", githubIssuer.name());
        });
    }
}
