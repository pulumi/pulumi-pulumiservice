import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();

const settings = new service.DeploymentSettings("deployment_settings", {
    organization: "service-provider-test-org",
    project: pulumi.getProject(),
    stack: pulumi.getStack(),
    executorContext: {
        executorImage: "pulumi-cli",
    },
    operationContext: {
        preRunCommands: ["yarn"],
        environmentVariables: {
            TEST_VAR: "bar",
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
    }
});
