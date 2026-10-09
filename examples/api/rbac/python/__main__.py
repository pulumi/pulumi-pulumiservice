import pulumi
import pulumi_pulumiservice as ps
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
name_suffix = config.get("nameSuffix") or "manual"
role_description = config.get("roleDescription") or "Stack read everywhere, plus write on stacks tagged team=platform."

# Pulumi Cloud models a role in three layers: a set grants scopes, a policy
# composes sets (optionally gated on a condition), and a role composes policies.
stack_read_set = ps_api.Role(
    "stackReadSet",
    org_name=organization_name,
    name=f"api-rbac-stack-read-{name_suffix}",
    ux_purpose="set",
    resource_type="stack",
    details=ps.build_allow_permissions_output(permissions=["stack:read"]).permissions,
)

stack_write_set = ps_api.Role(
    "stackWriteSet",
    org_name=organization_name,
    name=f"api-rbac-stack-write-{name_suffix}",
    ux_purpose="set",
    resource_type="stack",
    details=ps.build_allow_permissions_output(permissions=["stack:read", "stack:write"]).permissions,
)

platform_policy = ps_api.Role(
    "platformPolicy",
    org_name=organization_name,
    name=f"api-rbac-policy-{name_suffix}",
    ux_purpose="policy",
    details=ps.build_group_permissions_output(
        entries=[
            ps.build_compose_permissions_output(permission_descriptor_ids=[stack_read_set.role_id]).permissions,
            ps.build_tag_conditional_permissions_output(
                entity_type="stack",
                tag_key="team",
                tag_value="platform",
                set_ids=[stack_write_set.role_id],
            ).permissions,
        ],
    ).permissions,
)

platform_role = ps_api.Role(
    "platformRole",
    org_name=organization_name,
    name=f"api-rbac-role-{name_suffix}",
    description=role_description,
    ux_purpose="role",
    details=ps.build_compose_permissions_output(permission_descriptor_ids=[platform_policy.role_id]).permissions,
)

rbac_team = ps_api.teams.Team(
    "rbacTeam",
    org_name=organization_name,
    name=f"api-rbac-team-{name_suffix}",
    display_name=f"api RBAC Team {name_suffix}",
    description="Team scaffold used by the api rbac example.",
)

pulumi.export("roleName", platform_role.name)
pulumi.export("teamName", rbac_team.name)
