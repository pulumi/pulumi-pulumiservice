// TypeScript variant of canonical/04-deployment-pipeline.
// A git-driven Pulumi Deployments pipeline: source + executor settings,
// drift detection on a cron, TTL on ephemeral stacks. Behavioral twin
// of the sibling YAML program.

import * as pulumi from "@pulumi/pulumi";
import * as deployments from "@pulumi/pulumiservice/stacks/deployments";

const cfg = new pulumi.Config();
const organizationName = cfg.get("organizationName") ?? "service-provider-test-org";
const project = cfg.get("project") ?? "infrastructure";
const stack = cfg.get("stack") ?? "production";

new deployments.Settings("deploymentSettings", {
    organization: organizationName,
    project,
    stack,
    sourceContext: {
        git: {
            repoUrl: "https://github.com/acme-corp/infrastructure.git",
            branch: "refs/heads/main",
            repoDir: "stacks/production",
        },
    },
    github: {
        repository: "acme-corp/infrastructure",
        deployCommits: true,
        previewPullRequests: true,
        pullRequestTemplate: true,
    },
    executorContext: {
        executorImage: { image: "pulumi/pulumi:latest" },
    },
    operationContext: {
        preRunCommands: ["npm ci"],
        environmentVariables: { NODE_ENV: "production" },
    },
});

new deployments.DriftSchedule("driftCheck", {
    organization: organizationName,
    project,
    stack,
    scheduleCron: "0 */6 * * *",
    autoRemediate: false,
});

new deployments.TtlSchedule("ephemeralTtl", {
    organization: organizationName,
    project,
    stack,
    timestamp: "2026-12-31T00:00:00Z",
    deleteAfterDestroy: true,
});
