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
            // Thumbprints must match the certificate the issuer currently serves, so
            // they have no static default. Compute one with:
            //   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
            var pulumiThumbprint = config.require("pulumiThumbprint");
            var githubThumbprint = config.require("githubThumbprint");

            var pulumiIssuer = new OidcIssuer("pulumiIssuer",
                OidcIssuerArgs.builder()
                    .orgName(organizationName)
                    .name("pulumi_issuer_" + issuerSuffix)
                    .url("https://api.pulumi.com/oidc")
                    .thumbprints(List.of(pulumiThumbprint))
                    .build());

            var githubIssuer = new OidcIssuer("githubIssuer",
                OidcIssuerArgs.builder()
                    .orgName(organizationName)
                    .name("github_issuer_" + issuerSuffix)
                    .url("https://token.actions.githubusercontent.com")
                    .thumbprints(List.of(githubThumbprint))
                    .maxExpiration(maxExpiration)
                    .build());

            ctx.export("pulumiIssuerName", pulumiIssuer.name());
            ctx.export("githubIssuerName", githubIssuer.name());
        });
    }
}
