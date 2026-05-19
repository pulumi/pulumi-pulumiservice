import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "test-project";
const envSuffix = config.get("envSuffix") ?? "dev";

const environment = new ps.v2.esc.Environment("environment", {
    orgName: organizationName,
    project: projectName,
    name: `testing-environment-${envSuffix}`,
});

// esc:Environment exposes only id/urn — path-param inputs (orgName,
// project, name) are program-owned and don't surface on the resource.
export const environmentIdOut = environment.id;
