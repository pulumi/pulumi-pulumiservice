import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
project_name = config.get("projectName") or "test-project"
env_suffix = config.get("envSuffix") or "dev"

environment = ps_v2.Environment_esc_environments(
    "environment",
    org_name=service_org,
    project=project_name,
    name=f"testing-environment-{env_suffix}",
)

pulumi.export("envName", environment.name)
