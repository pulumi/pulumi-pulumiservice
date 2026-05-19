import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const issuerSuffix = config.get("issuerSuffix") ?? "dev";
const maxExpiration = config.getNumber("maxExpiration") ?? 3600;

const pulumiIssuer = new ps.v2.auth.OidcIssuer("pulumiIssuer", {
    orgName: organizationName,
    name: `pulumi_issuer_${issuerSuffix}`,
    url: "https://api.pulumi.com/oidc",
    thumbprints: ["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
});

const githubIssuer = new ps.v2.auth.OidcIssuer("githubIssuer", {
    orgName: organizationName,
    name: `github_issuer_${issuerSuffix}`,
    url: "https://token.actions.githubusercontent.com",
    thumbprints: ["b41ae0832808ebc94951437bf7e92b93ccb6479364daf894d46d6001bee7a486"],
    maxExpiration: maxExpiration,
});

export const pulumiIssuerName = pulumiIssuer.name;
export const githubIssuerName = githubIssuer.name;
