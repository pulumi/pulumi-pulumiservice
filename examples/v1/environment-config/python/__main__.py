import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
project_name = config.get("projectName") or "v1-envcfg-example"
env_name = config.get("envName") or "v1-envcfg-env"

draft = ps_v1.esc.EnvironmentDraft(
    "draft",
    org_name=organization_name,
    project_name=project_name,
    env_name=env_name,
)

settings = ps_v1.esc.EnvironmentSettings(
    "settings",
    org_name=organization_name,
    project_name=project_name,
    env_name=env_name,
    deletion_protected=True,
)

pulumi.export("draftId", draft.change_request_id)
pulumi.export("protected", settings.deletion_protected)
