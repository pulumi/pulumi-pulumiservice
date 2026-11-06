import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const org = config.require("organizationName");
const name = "test-smoke-ts-" + config.require("digits");

const environment = new pulumiservice.Environment("environment", {
    organization: org,
    project: "test-smoke",
    name: name,
    yaml: new pulumi.asset.StringAsset("values:\n  greeting: hello\n"),
});

const token = new pulumiservice.OrgAccessToken("token", {
    organizationName: org,
    name: name,
    description: "Created by the ts-smoke example.",
    admin: false,
});

const lookedUp = pulumiservice.getEnvironmentOutput({
    organizationName: environment.organization,
    projectName: environment.project,
    name: environment.name,
});

export const environmentName = lookedUp.name;
export const tokenValue = token.value;
