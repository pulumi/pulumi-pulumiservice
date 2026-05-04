import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const vcsSuffix = config.get("vcsSuffix") ?? "dev";
const baseUrl = config.get("baseUrl") ?? "https://git.example.invalid";
const envRef = config.get("envRef") ?? "organization/vcs-credentials";

const integration = new ps.v2.CustomVCSIntegration("integration", {
    orgName: serviceOrg,
    name: `v2-custom-vcs-${vcsSuffix}`,
    baseUrl: baseUrl,
    vcsType: "gitea",
    environment: envRef,
});

const repository = new ps.v2.CustomVCSRepository("repository", {
    orgName: serviceOrg,
    integrationId: integration.integrationId,
    name: `example-repo-${vcsSuffix}`,
    displayName: "Example Repository",
});

export const integrationId = integration.integrationId;
export const repositoryId = repository.id;
