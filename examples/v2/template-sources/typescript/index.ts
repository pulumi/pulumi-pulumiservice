import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const templateSuffix = config.get("templateSuffix") ?? "dev";
const sourceUrl = config.get("sourceUrl") ?? "https://github.com/pulumi/examples";

const source = new ps.v2.OrgTemplateCollection("source", {
    orgName: serviceOrg,
    name: `v2-templates-${templateSuffix}`,
    sourceURL: sourceUrl,
});

export const collectionName = source.name;
