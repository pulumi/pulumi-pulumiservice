import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
name_suffix = config.get("nameSuffix") or "manual"
role_description = config.get("roleDescription") or "Read-only access to stacks, created by the v2 rbac example."

read_only_role = ps_v2.Role(
    "readOnlyRole",
    org_name=service_org,
    name=f"v2-rbac-read-only-{name_suffix}",
    description=role_description,
    ux_purpose="role",
    details={
        "__type": "PermissionDescriptorAllow",
        "permissions": ["stack:read"],
    },
)

rbac_team = ps_v2.Team(
    "rbacTeam",
    org_name=service_org,
    name=f"v2-rbac-team-{name_suffix}",
    display_name=f"v2 RBAC Team {name_suffix}",
    description="Team scaffold used by the v2 rbac example.",
)

pulumi.export("roleName", read_only_role.name)
pulumi.export("teamName", rbac_team.name)
