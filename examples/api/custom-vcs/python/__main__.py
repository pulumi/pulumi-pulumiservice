import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
vcs_suffix = config.get("vcsSuffix") or "dev"
base_url = config.get("baseUrl") or "https://git.example.invalid"
env_ref = config.get("envRef") or "organization/vcs-credentials"

integration = ps_api.integrations.CustomVCSIntegration(
    "integration",
    org_name=organization_name,
    name=f"api-custom-vcs-{vcs_suffix}",
    base_url=base_url,
    vcs_type="gitea",
    environment=env_ref,
)

repository = ps_api.integrations.CustomVCSRepository(
    "repository",
    org_name=organization_name,
    integration_id=integration.integration_id,
    name=f"example-repo-{vcs_suffix}",
    display_name="Example Repository",
)

pulumi.export("integrationId", integration.integration_id)
pulumi.export("repositoryId", repository.id)
