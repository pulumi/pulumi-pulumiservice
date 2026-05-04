import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "test-project";
const envSuffix = config.get("envSuffix") ?? "dev";

const environment = new ps.v2.Environment_esc_environments("environment", {
    orgName: serviceOrg,
    project: projectName,
    name: `testing-environment-${envSuffix}`,
});

export const envNameOut = environment.name;
