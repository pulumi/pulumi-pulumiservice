import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const issuerSuffix = config.get("issuerSuffix") ?? "dev";
const maxExpiration = config.getNumber("maxExpiration") ?? 3600;

const pulumiIssuer = new ps.api.auth.OidcIssuer("pulumiIssuer", {
    orgName: organizationName,
    name: `pulumi_issuer_${issuerSuffix}`,
    url: "https://api.pulumi.com/oidc",
    thumbprints: ["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
});

const githubIssuer = new ps.api.auth.OidcIssuer("githubIssuer", {
    orgName: organizationName,
    name: `github_issuer_${issuerSuffix}`,
    url: "https://token.actions.githubusercontent.com",
    thumbprints: ["39517789ff0132a9212bafea4dc37401eae58b1bfac9756109d14301c90a6ab5"],
    maxExpiration: maxExpiration,
});

export const pulumiIssuerName = pulumiIssuer.name;
export const githubIssuerName = githubIssuer.name;
