import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

let config = new pulumi.Config();
let stackName = "test-stack-" + config.require("digits");
let envName = "test-env-" + config.require("digits");

// Deployment Settings are required to be setup before schedules can be
// Note the `DependsOn` option in all of the schedules
var settings = new service.DeploymentSettings("deployment-settings", {
    organization: "service-provider-test-org",
    project: "pulumi-service-schedules-example-ts",
    stack: stackName,
    sourceContext: {
        git: {
            repoUrl: "https://github.com/example.git",
            branch: "refs/heads/main"
        }
    }
})

// Environment to create rotations on
var environment = new service.Environment("testing-environment", {
  organization: "service-provider-test-org",
  project: "my-project",
  name: envName,
  yaml: new pulumi.asset.StringAsset(
`values:
  myNumber: 1`
  )
})

// Schedule that runs drift every Sunday midnight, but does NOT remediate it 
var drift = new service.DriftSchedule("drift-schedule", {
    organization: "service-provider-test-org",
    project: "pulumi-service-schedules-example-ts",
    stack: stackName,
    scheduleCron: "0 0 * * 0",
    autoRemediate: false
}, { dependsOn: settings })

// Schedule to destroy stack resources on Jan 1, 2026, but NOT delete the stack itself
var ttl = new service.TtlSchedule("ttl-schedule", {
    organization: "service-provider-test-org",
    project: "pulumi-service-schedules-example-ts",
    stack: stackName,
    timestamp: "2026-01-01T00:00:00Z",
    deleteAfterDestroy: false
}, { dependsOn: settings })

// Schedule that runs `pulumi up` every Sunday midnight
var deployment_up = new service.DeploymentSchedule("deployment-schedule-up", {
    organization: "service-provider-test-org",
    project: "pulumi-service-schedules-example-ts",
    stack: stackName,
    scheduleCron: "0 0 * * 0",
    pulumiOperation: service.PulumiOperation.Update
}, { dependsOn: settings })

// Schedule that runs `pulumi preview` once on Jan 1, 2026
var deployment_preview = new service.DeploymentSchedule("deployment-schedule-preview", {
    organization: "service-provider-test-org",
    project: "pulumi-service-schedules-example-ts",
    stack: stackName,
    timestamp: "2026-01-01T00:00:00Z",
    pulumiOperation: service.PulumiOperation.Preview
}, { dependsOn: settings })

// Schedule that runs environment secret rotation every Sunday midnight
var rotation_schedule = new service.EnvironmentRotationSchedule("environment-rotation-schedule", {
    organization: environment.organization,
    project: environment.project,
    environment: environment.name,
    scheduleCron: "0 0 * * 0",
})