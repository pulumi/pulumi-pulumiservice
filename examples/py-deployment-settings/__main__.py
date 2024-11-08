import pulumi
from pulumi_pulumiservice import (
    AgentPool,
    DeploymentSettings,
    DeploymentSettingsOperationContextArgs,
    DeploymentSettingsSourceContextArgs,
    DeploymentSettingsGitSourceArgs,
    DeploymentSettingsCacheOptionsArgs,
)

config = pulumi.Config()

agent_pool = AgentPool(
    "my-agent-pool",
    organization_name="service-provider-test-org",
    name="my-test-pool",
)

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
    ),
    agent_pool_id=agent_pool.agent_pool_id,
    cache_options=DeploymentSettingsCacheOptionsArgs(
        enable=False,
    )
)
