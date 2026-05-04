import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
project_name = config.get("projectName") or "my-new-project"
stack_name = config.get("stackName") or "dev"
executor_image = config.get("executorImage") or "pulumi-cli"

parent_stack = ps_v2.Stack(
    "parentStack",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
)

settings = ps_v2.DeploymentSettings(
    "settings",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
    executor_context={"executorImage": executor_image},
    operation_context={
        "preRunCommands": ["yarn"],
        "environmentVariables": {"TEST_VAR": "foo"},
        "options": {"skipInstallDependencies": True},
    },
    source_context={
        "git": {
            "repoUrl": "https://github.com/example/example.git",
            "branch": "refs/heads/main",
        },
    },
    opts=pulumi.ResourceOptions(depends_on=[parent_stack]),
)

pulumi.export("stackId", settings.stack_name)
