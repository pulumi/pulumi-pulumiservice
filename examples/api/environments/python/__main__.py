import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
project_name = config.get("projectName") or "test-project"
env_suffix = config.get("envSuffix") or "dev"
tag_value = config.get("tagValue") or "env-tag-initial"

environment = ps_api.esc.Environment(
    "environment",
    org_name=organization_name,
    project=project_name,
    name=f"testing-environment-{env_suffix}",
)

environment_tag = ps_api.esc.EnvironmentTag(
    "environmentTag",
    org_name=organization_name,
    project_name=project_name,
    env_name=f"testing-environment-{env_suffix}",
    name="purpose",
    value=tag_value,
    opts=pulumi.ResourceOptions(depends_on=[environment]),
)

pulumi.export("environmentId", environment.id)
pulumi.export("environmentTagValue", environment_tag.value)
