import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();

const tags = new service.StackTag("my-tag", {
    organization: "service-provider-test-org",
    project: "test-deployment-settings-proj",
    stack: "dev",
    name: "my-tag",
    value: "some-value"
})

const settings = new service.DeploymentSettings("deployment_settings", {
    organization: "service-provider-test-org",
    project: "test-deployment-settings-proj",
    stack: "dev",
    operationContext: {
        preRunCommands: ["yarn"],
        environmentVariables: {
            TEST_VAR: "foo",
            SECRET_VAR: config.requireSecret("my_secret"),
        },
        options: {
            skipInstallDependencies: true,
        }
    },
    sourceContext: {
        git: {
            repoUrl: "https://github.com/pulumi/deploy-demos.git",
            branch: "refs/heads/main",
            repoDir: "pulumi-programs/simple-resource"
        }
    }
});
