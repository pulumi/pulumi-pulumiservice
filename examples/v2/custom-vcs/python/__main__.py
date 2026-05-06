import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
vcs_suffix = config.get("vcsSuffix") or "dev"
base_url = config.get("baseUrl") or "https://git.example.invalid"
env_ref = config.get("envRef") or "organization/vcs-credentials"

integration = ps_v2.integrations.CustomVCSIntegration(
    "integration",
    org_name=service_org,
    name=f"v2-custom-vcs-{vcs_suffix}",
    base_url=base_url,
    vcs_type="gitea",
    environment=env_ref,
)

repository = ps_v2.integrations.CustomVCSRepository(
    "repository",
    org_name=service_org,
    integration_id=integration.integration_id,
    name=f"example-repo-{vcs_suffix}",
    display_name="Example Repository",
)

pulumi.export("integrationId", integration.integration_id)
pulumi.export("repositoryId", repository.id)
