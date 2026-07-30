import * as pulumiservice from "@pulumi/pulumiservice";

const org = process.env.POC_ORG ?? "poc-org";

// Fully typed recursive union: Group -> entries[Allow]. The __type literal is
// compiler-checked; a typo or missing tag fails tsc.
const role = new pulumiservice.api.Role("poc-union-role", {
    orgName: org,
    name: "poc-union-role",
    uxPurpose: "set",
    resourceType: "global",
    description: "POC: discriminated union permission tree",
    details: {
        __type: "PermissionDescriptorGroup",
        entries: [
            {
                __type: "PermissionDescriptorAllow",
                permissions: ["organization:read_usage"],
            },
        ],
    },
});

export const roleId = role.id;
export const detailsOut = role.details;
