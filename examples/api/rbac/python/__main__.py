import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
name_suffix = config.get("nameSuffix") or "manual"
role_description = config.get("roleDescription") or "Read-only access to stacks, created by the api rbac example."

read_only_role = ps_api.Role(
    "readOnlyRole",
    org_name=organization_name,
    name=f"api-rbac-read-only-{name_suffix}",
    description=role_description,
    ux_purpose="role",
    details={
        "__type": "PermissionDescriptorAllow",
        "permissions": ["stack:read"],
    },
)

rbac_team = ps_api.teams.Team(
    "rbacTeam",
    org_name=organization_name,
    name=f"api-rbac-team-{name_suffix}",
    display_name=f"api RBAC Team {name_suffix}",
    description="Team scaffold used by the api rbac example.",
)

pulumi.export("roleName", read_only_role.name)
pulumi.export("teamName", rbac_team.name)
