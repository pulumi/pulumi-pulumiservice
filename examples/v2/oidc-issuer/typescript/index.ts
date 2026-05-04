import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const issuerSuffix = config.get("issuerSuffix") ?? "dev";
const maxExpiration = config.getNumber("maxExpiration") ?? 3600;

const pulumiIssuer = new ps.v2.OidcIssuer("pulumiIssuer", {
    orgName: serviceOrg,
    name: `pulumi_issuer_${issuerSuffix}`,
    url: "https://api.pulumi.com/oidc",
    thumbprints: ["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
});

const githubIssuer = new ps.v2.OidcIssuer("githubIssuer", {
    orgName: serviceOrg,
    name: `github_issuer_${issuerSuffix}`,
    url: "https://token.actions.githubusercontent.com",
    thumbprints: ["caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"],
    maxExpiration: maxExpiration,
});

export const pulumiIssuerName = pulumiIssuer.name;
export const githubIssuerName = githubIssuer.name;
