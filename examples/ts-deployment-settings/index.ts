import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";
import * as random from "@pulumi/random";

const config = new pulumi.Config();

const id = new random.RandomUuid("id");

const stack = new service.Stack("my_stack", {
    organizationName: "service-provider-test-org",
    projectName: "my-new-project",
    stackName: id.result,
    forceDestroy: false,
})

const settings = new service.DeploymentSettings("deployment_settings", {
    organization: "service-provider-test-org",
    project: stack.projectName,
    stack: stack.stackName,
    executorContext: {
        executorImage: "pulumi-cli",
    },
    operationContext: {
        preRunCommands: ["yarn"],
        environmentVariables: {
            TEST_VAR: "foo",
            SECRET_VAR: config.requireSecret("my_secret"),
        },
        options: {
            skipInstallDependencies: true,
        },
        oidc: {
            aws: {
                roleARN: "arn:aws:iam::123456789012:role/MyRole",
                duration: "1h",
                sessionName: "my-session",
            },
            gcp: {
                projectId: "my-project",
                region: "us-west1",
                workloadPoolId: "my-workload-pool",
                providerId: "my-provider",
                serviceAccount: "my-service-account",
                tokenLifetime: "30s",
            },
            azure: {
                tenantId: "my-tenant-id",
                clientId: "my-client-id",
                subscriptionId: "my-subscription-id",
            }
        }
    },
    sourceContext: {
        git: {
            repoUrl: "https://github.com/pulumi/deploy-demos.git",
            branch: "refs/heads/main",
            repoDir: "pulumi-programs/simple-resource",
            gitAuth: {
                sshAuth: {
                    sshPrivateKey: "key",
                    password: config.requireSecret("password"),
                }
            }
        }
    },
    cacheOptions: {
        enable: pulumi.secret(true),
    }
});
