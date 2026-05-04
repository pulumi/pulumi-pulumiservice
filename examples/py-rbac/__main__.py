"""Python Pulumi Service RBAC example.

Exercises the full RBAC flow from Python:
- Fetches org members via the getOrganizationMembers data source.
- Creates a custom OrganizationRole.
- Creates a Team whose membership is derived from the data-source output
  (filtered to known-safe fixture users so the integration test never
  re-parents a real admin) plus the authenticated caller, who Pulumi
  Cloud auto-adds on team creation.
- Binds the custom role to the team via TeamRoleAssignment.
- Adopts an existing org member via OrganizationMember and flips their
  built-in role to "admin".
- Builds an environment-scoped role using
  build_environment_scoped_permissions_output and assigns it to the
  caller via the scopedReadOnlyRole resource.
"""

import pulumi
from pulumi_pulumiservice import (
    Environment,
    OrganizationMember,
    OrganizationRole,
    Team,
    TeamRoleAssignment,
    get_current_user_output,
    build_environment_scoped_permissions_output,
    get_organization_members_output,
)

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
digits = config.require("digits")

# Fixture users the test org is known to have. Filtering the data-source
# output to this set keeps the test from putting real admins on the team or
# demoting them.
SAFE_USERS = {"pulumi-bot", "service-provider-example-user"}

# The single user we flip to "admin" below. Kept separate so the test's
# cleanup hook can reset just this user back to "member".
ROLE_CHANGE_USER = "service-provider-example-user"

# 1. Fetch org members and the authenticated caller.
current_members = get_organization_members_output(organization_name=organization_name)
current_user = get_current_user_output()

# 2. Derive the team member list dynamically: known-safe org members plus
#    the authenticated caller. Pulumi Cloud auto-adds the team creator on
#    team creation, so seeding the list with the caller keeps refresh from
#    detecting that as drift. Sorted so the order is stable across runs.
#
# Implemented as a nested apply (rather than pulumi.Output.all) because
# Output.all serializes the inner list of OrganizationMemberInfo down to
# plain dicts, which loses the .username attribute access. Nesting the
# applies preserves the typed shape that get_organization_members_output
# returns.
team_members = current_user.username.apply(
    lambda caller: current_members.members.apply(
        lambda ms: sorted(
            {m.username for m in ms if m.username in SAFE_USERS} | {caller}
        )
    )
)

# 3. Custom org role that grants stack:read. The simplest descriptor: a
#    flat ``PermissionDescriptorAllow``. The SDK uses the same PascalCase
#    ``kind`` values as Pulumi Cloud's REST API (``__type`` in the wire
#    format is renamed to ``kind`` to dodge Python's underscore-prefix
#    stripping — see pulumi/pulumi#22738).
custom_role = OrganizationRole(
    "rbacRole",
    organization_name=organization_name,
    name=f"py-rbac-read-only-{digits}",
    description="Read-only access to stacks, created by the py-rbac example.",
    permissions={
        "discriminator": "PermissionDescriptorAllow",
        "permissions": ["stack:read"],
    },
)

# 4. Team with members pulled from the data source.
team = Team(
    "rbacTeam",
    organization_name=organization_name,
    name=f"py-rbac-team-{digits}",
    team_type="pulumi",
    display_name=f"py-rbac team ({digits})",
    description="Team created by the py-rbac example.",
    members=team_members,
)

# 5. Attach the custom role to the team. The team's organization must have
#    the custom-roles feature enabled.
team_role = TeamRoleAssignment(
    "rbacTeamRoleBinding",
    organization_name=organization_name,
    team_name=team.name,
    role_id=custom_role.role_id,
)

# 6. Change the built-in org role for an existing member. Create hits the
#    409 branch, adopts the existing membership, then updates the role.
member_role_change = OrganizationMember(
    "rbacMemberRoleChange",
    organization_name=organization_name,
    username=ROLE_CHANGE_USER,
    role="admin",
)

pulumi.export("teamMemberCount", team_members.apply(len))
pulumi.export("roleId", custom_role.role_id)
pulumi.export("assignedRoleName", team_role.role_name)
pulumi.export("memberAdopted", member_role_change.adopted)
pulumi.export("memberRole", member_role_change.role)

# 7. ESC environment used to demonstrate per-env role scoping. The
#    Environment resource exposes its UUID via `environment_id`, which the
#    next role pins its grants to.
scoped_env = Environment(
    "scopedEnv",
    organization=organization_name,
    project="default",
    name=f"py-rbac-scoped-env-{digits}",
    yaml=pulumi.StringAsset(
        'values:\n  placeholder: "py-rbac scoped env fixture"\n',
    ),
)

# 8. A role that grants environment:read + environment:open ONLY on the
#    env created above. Anywhere else in the org the role grants nothing.
#    The role definition is org-scoped (resourceType defaults to "global");
#    the helper returns a ``PermissionDescriptorCondition(Equal(...),
#    Allow)`` tree — the same wire shape Pulumi Cloud's REST API uses,
#    modulo the ``__type`` → ``kind`` rename. The provider has no
#    SDK-side translation; it sends the descriptor on the wire verbatim.
scoped_role = OrganizationRole(
    "scopedReadOnlyRole",
    organization_name=organization_name,
    name=f"py-rbac-scoped-read-only-{digits}",
    description="Read+open access scoped to a single ESC environment.",
    permissions=build_environment_scoped_permissions_output(
        environment_id=scoped_env.environment_id,
        permissions=["environment:read", "environment:open"],
    ).permissions,
)

pulumi.export("scopedEnvironmentId", scoped_env.environment_id)
pulumi.export("scopedRoleId", scoped_role.role_id)
