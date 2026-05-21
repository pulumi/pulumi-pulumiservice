import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const nameSuffix = config.get("nameSuffix") ?? "manual";
const roleDescription = config.get("roleDescription") ?? "Read-only access to stacks, created by the api rbac example.";

const readOnlyRole = new ps.api.Role("readOnlyRole", {
    orgName: organizationName,
    name: `api-rbac-read-only-${nameSuffix}`,
    description: roleDescription,
    uxPurpose: "role",
    details: {
        __type: "PermissionDescriptorAllow",
        permissions: ["stack:read"],
    },
});

const rbacTeam = new ps.api.teams.Team("rbacTeam", {
    orgName: organizationName,
    name: `api-rbac-team-${nameSuffix}`,
    displayName: `api RBAC Team ${nameSuffix}`,
    description: "Team scaffold used by the api rbac example.",
});

export const roleName = readOnlyRole.name;
export const teamName = rbacTeam.name;
