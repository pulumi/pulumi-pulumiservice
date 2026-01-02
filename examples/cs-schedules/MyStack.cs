using System;
using Pulumi;
using Pulumi.PulumiService;
using Pulumi.PulumiService.Inputs;

class MyStack : Pulumi.Stack
{
    public MyStack()
    {
        var config = new Pulumi.Config();
        var stackName = "test-stack-" + config.Require("digits");
        var envName = "test-env-" + config.Require("digits");
        String yaml = """
        values:
          myKey1: "myValue1"
        """;

        // Deployment Settings are required to be setup before schedules can be
        // Note the `DependsOn` option in all of the schedules
        var settings = new DeploymentSettings(
            "Deployment Settings",
            new DeploymentSettingsArgs
            {
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = stackName,
                SourceContext = new DeploymentSettingsSourceContextArgs
                {
                    Git = new DeploymentSettingsGitSourceArgs
                    {
                        RepoUrl = "https://github.com/example.git",
                        Branch = "refs/heads/main"
                    }
                }
            }
        );

        // Environment to create rotations on
        var environment = new Pulumi.PulumiService.Environment(
            "testing-environment",
            new EnvironmentArgs
            {
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Name = envName,
                Yaml = new StringAsset(yaml)
            }
        );

        // Schedule that runs drift every Sunday midnight, but does NOT remediate it
        var drift = new DriftSchedule(
            "drift-schedule",
            new DriftScheduleArgs
            {
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = stackName,
                ScheduleCron = "0 0 * * 0",
                AutoRemediate = false
            },
            new CustomResourceOptions
            {
                DependsOn = { settings }
            }
        );

        // Schedule to destroy stack resources on Jan 1, 2099, but NOT delete the stack itself
        var ttl = new TtlSchedule(
            "ttl-schedule",
            new TtlScheduleArgs
            {
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = stackName,
                Timestamp = "2099-01-01T00:00:00Z",
                DeleteAfterDestroy = false
            },
            new CustomResourceOptions
            {
                DependsOn = { settings }
            }
        );

        // Schedule that runs `pulumi up` every Sunday midnight
        var deploymentUp = new DeploymentSchedule(
            "deployment-schedule-up",
            new DeploymentScheduleArgs
            {
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = stackName,
                ScheduleCron = "0 0 * * 0",
                PulumiOperation = PulumiOperation.Update
            },
            new CustomResourceOptions
            {
                DependsOn = { settings }
            }
        );

        // Schedule that runs `pulumi preview` once on Jan 1, 2099
        var deploymentPreview = new DeploymentSchedule(
            "deployment-schedule-preview",
            new DeploymentScheduleArgs
            {
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = stackName,
                Timestamp = "2099-01-01T00:00:00Z",
                PulumiOperation = PulumiOperation.Preview
            },
            new CustomResourceOptions
            {
                DependsOn = { settings }
            }
        );
        
        // Schedule that runs environment secret rotation every Sunday midnight
        var environmentRotation = new EnvironmentRotationSchedule(
            "environment-rotation-schedule",
            new EnvironmentRotationScheduleArgs{
                Organization = environment.Organization,
                Project = environment.Project,
                Environment = environment.Name,
                ScheduleCron = "0 0 * * 0",
            }
        );
    }
}
