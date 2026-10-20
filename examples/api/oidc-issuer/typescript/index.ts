import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const issuerSuffix = config.get("issuerSuffix") ?? "dev";
const maxExpiration = config.getNumber("maxExpiration") ?? 3600;
// Tests override the issuer URLs so that concurrent runs don't conflict;
// see oidcIssuerTestConfig in examples/examples_utils_test.go.
const pulumiIssuerUrl = config.get("pulumiIssuerUrl") ?? "https://api.pulumi.com/oidc";
const githubIssuerUrl = config.get("githubIssuerUrl") ?? "https://token.actions.githubusercontent.com";
// Thumbprints must match the certificate the issuer currently serves, so they
// have no static default. Compute one with:
//   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
const pulumiThumbprint = config.require("pulumiThumbprint");
const githubThumbprint = config.require("githubThumbprint");

const pulumiIssuer = new ps.api.auth.OidcIssuer("pulumiIssuer", {
    orgName: organizationName,
    name: `pulumi_issuer_${issuerSuffix}`,
    url: pulumiIssuerUrl,
    thumbprints: [pulumiThumbprint],
});

const githubIssuer = new ps.api.auth.OidcIssuer("githubIssuer", {
    orgName: organizationName,
    name: `github_issuer_${issuerSuffix}`,
    url: githubIssuerUrl,
    thumbprints: [githubThumbprint],
    maxExpiration: maxExpiration,
});

export const pulumiIssuerName = pulumiIssuer.name;
export const githubIssuerName = githubIssuer.name;
