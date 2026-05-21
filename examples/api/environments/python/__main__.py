import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
project_name = config.get("projectName") or "test-project"
env_suffix = config.get("envSuffix") or "dev"

environment = ps_api.esc.Environment(
    "environment",
    org_name=organization_name,
    project=project_name,
    name=f"testing-environment-{env_suffix}",
)

pulumi.export("environmentId", environment.id)
