import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") || "service-provider-test-org2";
const pulumiIssuer = new pulumiservice.v2.OidcIssuer("pulumiIssuer", {
    orgName: serviceOrg,
    name: "pulumi_issuer",
    url: "https://api.pulumi.com/oidc",
    thumbprints: ["57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"],
});
const githubIssuer = new pulumiservice.v2.OidcIssuer("githubIssuer", {
    orgName: serviceOrg,
    name: "github_issuer",
    url: "https://token.actions.githubusercontent.com",
    thumbprints: ["caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"],
    maxExpiration: 3600,
});
export const pulumiIssuerUrl = pulumiIssuer.url;
export const githubIssuerUrl = githubIssuer.url;
