import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const projectName = config.get("projectName") ?? "api-envcfg-example";
const envName = config.get("envName") ?? "api-envcfg-env";

const draft = new ps.api.esc.EnvironmentDraft("draft", {
    orgName: organizationName,
    projectName: projectName,
    envName: envName,
});

const settings = new ps.api.esc.EnvironmentSettings("settings", {
    orgName: organizationName,
    projectName: projectName,
    envName: envName,
    deletionProtected: true,
});

export const draftId = draft.changeRequestId;
export const protect = settings.deletionProtected;
