using Pulumi;
using Pulumi.PulumiService;
using Pulumi.PulumiService.Inputs;
using Pulumi.Random;

class MyStack : Stack
{
    public MyStack()
    {
        // Deployment Settings are required to be setup before schedules can be
        // Note the `DependsOn` option in all of the schedules
        var settings = new DeploymentSettings(
            "Deployment Settings",
            new DeploymentSettingsArgs{
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = "test-stack",
                SourceContext = new DeploymentSettingsSourceContextArgs{
                    Git = new DeploymentSettingsGitSourceArgs{
                        RepoUrl = "https://github.com/example.git",
                        Branch = "refs/heads/main"
                    }
                }
            }
        );

        // Schedule that runs drift every Sunday midnight, but does NOT remediate it
        var drift = new DriftSchedule(
            "drift-schedule",
            new DriftScheduleArgs
            {
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = "test-stack",
                ScheduleCron = "0 0 * * 0",
                AutoRemediate = false
            }, 
            new CustomResourceOptions
            {
              DependsOn = { settings }
            }
        );

        // Schedule to destroy stack resources on Jan 1, 2026, but NOT delete the stack itself
        var ttl = new TtlSchedule(
            "ttl-schedule",
            new TtlScheduleArgs{
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = "test-stack",
                Timestamp = "2026-01-01T00:00:00Z",
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
            new DeploymentScheduleArgs{
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = "test-stack",
                ScheduleCron = "0 0 * * 0",
                PulumiOperation = PulumiOperation.Update
            }, 
            new CustomResourceOptions
            {
              DependsOn = { settings }
            }
        );

        // Schedule that runs `pulumi preview` once on Jan 1, 2026
        var deploymentPreview = new DeploymentSchedule(
            "deployment-schedule-preview",
            new DeploymentScheduleArgs{
                Organization = "service-provider-test-org",
                Project = "cs-schedules",
                Stack = "test-stack",
                Timestamp = "2026-01-01T00:00:00Z",
                PulumiOperation = PulumiOperation.Preview
            }, 
            new CustomResourceOptions
            {
              DependsOn = { settings }
            }
        );
    }
}
