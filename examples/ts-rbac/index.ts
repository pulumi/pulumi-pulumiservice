import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.require("organizationName");
const targetUsername = config.require("targetUsername");
const nameSuffix = config.get("nameSuffix") ?? "manual";

// A custom organization-level role that grants stack read access.
const readOnlyRole = new service.OrganizationRole("readOnlyRole", {
    organizationName,
    name: `ts-rbac-read-only-${nameSuffix}`,
    description: "Read-only access to stacks, created by the ts-rbac example.",
    permissions: {
        __type: "PermissionDescriptorAllow",
        permissions: ["stack:read"],
    },
});

// A team that will receive the custom role. Pulumi Cloud adds the team
// creator as the first member automatically, so we seed `members` with the
// current user to keep refresh from detecting that as drift.
const currentUser = service.getCurrentUserOutput();
const teamName = `ts-rbac-team-${nameSuffix}`;
const rbacTeam = new service.Team("rbacTeam", {
    organizationName,
    name: teamName,
    teamType: "pulumi",
    displayName: `ts-rbac team (${nameSuffix})`,
    description: "Team created by the ts-rbac example.",
    members: [currentUser.username],
});

// Assign the custom role to the team. The team's organization must have
// the custom-roles feature enabled.
const rbacTeamRoleBinding = new service.TeamRoleAssignment("rbacTeamRoleBinding", {
    organizationName,
    teamName: rbacTeam.name.apply(n => n ?? teamName),
    roleId: readOnlyRole.roleId,
});

// Exercise OrganizationMember via adoption: if the user already exists in
// the org, Create hits 409 -> adopts the membership (adopted=true), then
// assigns the custom role. Destroying readOnlyRole above (force=true)
// revokes the assignment on teardown.
const rbacMember = new service.OrganizationMember("rbacMember", {
    organizationName,
    username: targetUsername,
    roleId: readOnlyRole.roleId,
});

// Data source: list every member of the organization.
const currentMembers = service.getOrganizationMembersOutput({ organizationName });

// Data source: discover the permission-scope catalogue. Customers use this
// to find the valid scope names to put in OrganizationRole.permissions.
const availableScopes = service.getOrganizationRoleScopesOutput({ organizationName });

// Data source: single-member lookup by username.
const memberByUsername = service.getOrganizationMemberOutput({
    organizationName,
    username: targetUsername,
});

export const roleId = readOnlyRole.roleId;
export const roleVersion = readOnlyRole.version;
export const assignedRoleName = rbacTeamRoleBinding.roleName;
export const firstMember = currentMembers.members[0].username;
export const firstScope = availableScopes.scopes[0].name;
export const memberAdopted = rbacMember.adopted;
export const memberAssignedRole = rbacMember.roleName;
export const lookedUpByUsernameRole = memberByUsername.role;
