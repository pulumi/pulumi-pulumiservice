import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const nameSuffix = config.get("nameSuffix") ?? "manual";
const roleDescription = config.get("roleDescription") ?? "Stack read everywhere, plus write on stacks tagged team=platform.";

// Pulumi Cloud models a role in three layers: a set grants scopes, a policy
// composes sets (optionally gated on a condition), and a role composes policies.
const stackReadSet = new ps.api.Role("stackReadSet", {
    orgName: organizationName,
    name: `api-rbac-stack-read-${nameSuffix}`,
    uxPurpose: "set",
    resourceType: "stack",
    details: ps.buildAllowPermissionsOutput({ permissions: ["stack:read"] }).permissions,
});

const stackWriteSet = new ps.api.Role("stackWriteSet", {
    orgName: organizationName,
    name: `api-rbac-stack-write-${nameSuffix}`,
    uxPurpose: "set",
    resourceType: "stack",
    details: ps.buildAllowPermissionsOutput({ permissions: ["stack:read", "stack:write"] }).permissions,
});

const platformPolicy = new ps.api.Role("platformPolicy", {
    orgName: organizationName,
    name: `api-rbac-policy-${nameSuffix}`,
    uxPurpose: "policy",
    details: ps.buildGroupPermissionsOutput({
        entries: [
            ps.buildComposePermissionsOutput({ permissionDescriptorIds: [stackReadSet.roleID] }).permissions,
            ps.buildTagConditionalPermissionsOutput({
                entityType: "stack",
                tagKey: "team",
                tagValue: "platform",
                setIds: [stackWriteSet.roleID],
            }).permissions,
        ],
    }).permissions,
});

const platformRole = new ps.api.Role("platformRole", {
    orgName: organizationName,
    name: `api-rbac-role-${nameSuffix}`,
    description: roleDescription,
    uxPurpose: "role",
    details: ps.buildComposePermissionsOutput({ permissionDescriptorIds: [platformPolicy.roleID] }).permissions,
});

const rbacTeam = new ps.api.teams.Team("rbacTeam", {
    orgName: organizationName,
    name: `api-rbac-team-${nameSuffix}`,
    displayName: `api RBAC Team ${nameSuffix}`,
    description: "Team scaffold used by the api rbac example.",
});

export const roleName = platformRole.name;
export const teamName = rbacTeam.name;
