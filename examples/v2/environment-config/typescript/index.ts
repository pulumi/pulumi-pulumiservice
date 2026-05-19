import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "v2-envcfg-example";
const envName = config.get("envName") ?? "v2-envcfg-env";

const draft = new ps.v2.esc.EnvironmentDraft("draft", {
    orgName: organizationName,
    projectName: projectName,
    envName: envName,
});

const settings = new ps.v2.esc.EnvironmentSettings("settings", {
    orgName: organizationName,
    projectName: projectName,
    envName: envName,
    deletionProtected: true,
});

export const draftId = draft.changeRequestId;
export const protect = settings.deletionProtected;
