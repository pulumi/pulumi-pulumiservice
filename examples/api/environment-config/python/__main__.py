import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
project_name = config.get("projectName") or "api-envcfg-example"
env_name = config.get("envName") or "api-envcfg-env"

draft = ps_api.esc.EnvironmentDraft(
    "draft",
    org_name=organization_name,
    project_name=project_name,
    env_name=env_name,
)

settings = ps_api.esc.EnvironmentSettings(
    "settings",
    org_name=organization_name,
    project_name=project_name,
    env_name=env_name,
    deletion_protected=True,
)

pulumi.export("draftId", draft.change_request_id)
pulumi.export("protected", settings.deletion_protected)
