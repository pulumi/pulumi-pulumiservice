import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "v2-envcfg-example";
const envName = config.get("envName") ?? "v2-envcfg-env";

const draft = new ps.v2.esc.EnvironmentDraft("draft", {
    orgName: serviceOrg,
    projectName: projectName,
    envName: envName,
});

const settings = new ps.v2.esc.EnvironmentSettings("settings", {
    orgName: serviceOrg,
    projectName: projectName,
    envName: envName,
    deletionProtected: true,
});

export const draftId = draft.changeRequestID;
export const protect = settings.deletionProtected;
