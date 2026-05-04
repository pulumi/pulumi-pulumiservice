import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
project_name = config.get("projectName") or "v2-envcfg-example"
env_name = config.get("envName") or "v2-envcfg-env"

draft = ps_v2.EnvironmentDraft(
    "draft",
    org_name=service_org,
    project_name=project_name,
    env_name=env_name,
)

settings = ps_v2.EnvironmentSettings(
    "settings",
    org_name=service_org,
    project_name=project_name,
    env_name=env_name,
    deletion_protected=True,
)

pulumi.export("draftId", draft.change_request_id)
pulumi.export("protected", settings.deletion_protected)
