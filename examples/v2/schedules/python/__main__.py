import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
project_name = config.get("projectName") or "pulumi-service-schedules-example"
stack_name = config.get("stackName") or "dev"
schedule_cron = config.get("scheduleCron") or "0 7 * * *"

parent_stack = ps_v2.Stack(
    "parentStack",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
)

parent_settings = ps_v2.DeploymentSettings(
    "parentSettings",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
    source_context={"git": {"repoUrl": "https://github.com/example/example.git", "branch": "refs/heads/main"}},
    opts=pulumi.ResourceOptions(depends_on=[parent_stack]),
)

nightly_deploy = ps_v2.ScheduledDeployment(
    "nightlyDeploy",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
    schedule_cron=schedule_cron,
    request={"operation": "update"},
    opts=pulumi.ResourceOptions(depends_on=[parent_settings]),
)

pulumi.export("nightlyCron", nightly_deploy.schedule_cron)
