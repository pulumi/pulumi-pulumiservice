"""A Python Pulumi Service Teams program"""

import pulumi
from pulumi_pulumiservice import DeploymentSettings, DeploymentSettingsOperationContextArgs, DeploymentSettingsSourceContextArgs, DeploymentSettingsGitSourceArgs

config = pulumi.Config()

settings = DeploymentSettings(
    "my-settings",
    organization="service-provider-test-org",
    project=pulumi.get_project(),
    stack=pulumi.get_stack(),
    operation_context=DeploymentSettingsOperationContextArgs(
        pre_run_commands=["echo 'pre-run'", "poetry install"],
        environment_variables={
            "MY_ENV_VAR": "my-value",
            "MY_SECRET_ENV_VAR": config.require_secret("my-secret")
        }
    ),
    source_context=DeploymentSettingsSourceContextArgs(
        git=DeploymentSettingsGitSourceArgs(
            repo_url="https://github.com/pulumi/deploy-demos.git",
            branch="main",
            repo_dir="pulumi-programs/simple-resource"
        )
    )
)
