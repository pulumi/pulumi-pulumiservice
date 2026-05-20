import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
name_suffix = config.get("nameSuffix") or "manual"
role_description = config.get("roleDescription") or "Read-only access to stacks, created by the v1 rbac example."

read_only_role = ps_v1.Role(
    "readOnlyRole",
    org_name=organization_name,
    name=f"v1-rbac-read-only-{name_suffix}",
    description=role_description,
    ux_purpose="role",
    details={
        "__type": "PermissionDescriptorAllow",
        "permissions": ["stack:read"],
    },
)

rbac_team = ps_v1.teams.Team(
    "rbacTeam",
    org_name=organization_name,
    name=f"v1-rbac-team-{name_suffix}",
    display_name=f"v1 RBAC Team {name_suffix}",
    description="Team scaffold used by the v1 rbac example.",
)

pulumi.export("roleName", read_only_role.name)
pulumi.export("teamName", rbac_team.name)
