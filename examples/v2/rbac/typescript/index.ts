import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const nameSuffix = config.get("nameSuffix") ?? "manual";
const roleDescription = config.get("roleDescription") ?? "Read-only access to stacks, created by the v2 rbac example.";

const readOnlyRole = new ps.v2.Role("readOnlyRole", {
    orgName: organizationName,
    name: `v2-rbac-read-only-${nameSuffix}`,
    description: roleDescription,
    uxPurpose: "role",
    details: {
        __type: "PermissionDescriptorAllow",
        permissions: ["stack:read"],
    },
});

const rbacTeam = new ps.v2.teams.Team("rbacTeam", {
    orgName: organizationName,
    name: `v2-rbac-team-${nameSuffix}`,
    displayName: `v2 RBAC Team ${nameSuffix}`,
    description: "Team scaffold used by the v2 rbac example.",
});

export const roleName = readOnlyRole.name;
export const teamName = rbacTeam.name;
