import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const nameSuffix = config.get("nameSuffix") ?? "manual";
const roleDescription = config.get("roleDescription") ?? "Environment-scoped read access, created by the api rbac-scoped example.";
const allowedEnvironmentId = config.get("allowedEnvironmentId") ?? "c5549aa1-87db-4d67-a195-455b56772900";
const deniedEnvironmentId = config.get("deniedEnvironmentId") ?? "3cb9b7ad-0848-4e0d-aeff-8e9f093fd2d9";

// Every level below is a discriminated union variant the compiler checks:
// the descriptor layer (Group/Condition/Allow) and the expression layer
// (And/Not/Equal + context/literal environment expressions).
const envMatch = (identity: pulumi.Input<string>): ps.types.input.api.PermissionExpressionEqualArgs => ({
    __type: "PermissionExpressionEqual",
    left: { __type: "PermissionExpressionEnvironment" },
    right: { __type: "PermissionLiteralExpressionEnvironment", identity },
});

const details: ps.types.input.api.PermissionDescriptorGroupArgs = {
    __type: "PermissionDescriptorGroup",
    entries: [{
        __type: "PermissionDescriptorCondition",
        condition: {
            __type: "PermissionExpressionAnd",
            left: envMatch(allowedEnvironmentId),
            right: {
                __type: "PermissionExpressionNot",
                node: envMatch(deniedEnvironmentId),
            },
        },
        subNode: {
            __type: "PermissionDescriptorAllow",
            permissions: ["environment:read", "environment:open"],
        },
    }],
};

const scopedRole = new ps.api.Role("scopedRole", {
    orgName: organizationName,
    name: `api-rbac-scoped-${nameSuffix}`,
    description: roleDescription,
    uxPurpose: "role",
    details,
});

export const roleName = scopedRole.name;
